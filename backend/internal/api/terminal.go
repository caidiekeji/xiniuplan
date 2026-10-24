package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	termsvc "cloudpivot/internal/service/terminal"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

var terminalManager = termsvc.NewSessionManager()

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 校验 Origin：生产环境仅允许同源
	CheckOrigin: func(r *http.Request) bool {
		if gin.Mode() != gin.ReleaseMode {
			return true // 开发环境允许所有来源
		}
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // 非浏览器请求（如 curl），无 Origin 头
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		return strings.EqualFold(u.Host, r.Host)
	},
}

func init() {
	// 启动会话清理：每 5 分钟检查，清理超过 30 分钟未活跃的会话
	terminalManager.StartCleanupLoop(5*time.Minute, 30*time.Minute)
}

// RegisterTerminalWSRoutes 注册终端 WebSocket 路由与增强功能接口
func RegisterTerminalWSRoutes(r *gin.RouterGroup) {
	r.GET("/ws/terminal", middleware.RequirePermission(model.PermTerminalView), handleTerminalWS)

	// 快捷命令面板（T-004）
	r.GET("/terminal/quick-commands", middleware.RequirePermission(model.PermTerminalView), h_ListQuickCommands)

	// 会话录制（T-006）
	r.GET("/terminal/records", middleware.RequirePermission(model.PermTerminalView), h_ListTerminalRecords)
	r.GET("/terminal/records/:name", middleware.RequirePermission(model.PermTerminalView), h_GetTerminalRecord)
	r.DELETE("/terminal/records/:name", middleware.RequirePermission(model.PermTerminalCreate), h_DeleteTerminalRecord)
}

// h_ListQuickCommands GET /terminal/quick-commands
func h_ListQuickCommands(c *gin.Context) {
	common.Success(c, termsvc.QuickCommands)
}

// h_ListTerminalRecords GET /terminal/records
func h_ListTerminalRecords(c *gin.Context) {
	records, err := termsvc.ListRecords()
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, records)
}

// h_GetTerminalRecord GET /terminal/records/:name - 回放录制内容
func h_GetTerminalRecord(c *gin.Context) {
	content, err := termsvc.ReadRecord(c.Param("name"))
	if err != nil {
		common.Fail(c, common.CodeNotFound, err.Error())
		return
	}
	common.Success(c, gin.H{"name": c.Param("name"), "content": content})
}

// h_DeleteTerminalRecord DELETE /terminal/records/:name
func h_DeleteTerminalRecord(c *gin.Context) {
	if err := termsvc.DeleteRecord(c.Param("name")); err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, gin.H{"deleted": true})
}

func handleTerminalWS(c *gin.Context) {
	sessionID := c.Query("session_id")
	cols, _ := strconv.Atoi(c.DefaultQuery("cols", "80"))
	rows, _ := strconv.Atoi(c.DefaultQuery("rows", "24"))
	// raw=1：原始字节透传（ZMODEM 支持，T-005）；record=1：会话录制（T-006）
	rawMode := c.Query("raw") == "1"
	record := c.Query("record") == "1"

	// 升级 WebSocket
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Warn("[terminal] WebSocket upgrade failed", "error", err)
		return
	}

	var session *termsvc.TerminalSession

	if sessionID != "" {
		// 尝试恢复已有会话
		if existing, ok := terminalManager.Get(sessionID); ok && !existing.IsClosed() {
			session = existing
			session.SetConn(conn)
			session.Touch()

			// 发送会话恢复确认
			ack, _ := json.Marshal(dto.TerminalMessage{
				Type: "output",
				Data: "\r\n[会话已恢复] " + session.ID + "\r\n",
			})
			_ = conn.WriteMessage(websocket.TextMessage, ack)
		}
	}

	if session == nil {
		// 创建新会话：SSH 连接参数通过升级后的首条消息传入
		// （不再经 URL query，避免密码进入 nginx/网关访问日志）
		session = terminalManager.Create()
		session.Conn = conn

		sshClient, err := connectSSHFromFirstMessage(conn)
		if err != nil {
			errMsg, _ := json.Marshal(dto.TerminalMessage{
				Type: "error",
				Data: "SSH 连接失败: " + err.Error(),
			})
			_ = conn.WriteMessage(websocket.TextMessage, errMsg)
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			terminalManager.Close(session.ID)
			return
		}
		session.SSHClient = sshClient

		// 启动 PTY
		sshSession, stdin, stdout, err := termsvc.StartPTY(sshClient, cols, rows)
		if err != nil {
			errMsg, _ := json.Marshal(dto.TerminalMessage{
				Type: "error",
				Data: "启动 PTY 失败: " + err.Error(),
			})
			_ = conn.WriteMessage(websocket.TextMessage, errMsg)
			_ = sshClient.Close()
			terminalManager.Close(session.ID)
			return
		}
		session.SSHSession = sshSession
		session.Stdin = stdin
		session.RawMode = rawMode

		// 会话录制（T-006）
		if record {
			if path, err := session.EnableRecording(); err != nil {
				slog.Warn("[terminal] enable recording failed", "error", err)
			} else {
				slog.Info("[terminal] recording enabled", "session", session.ID, "file", path)
			}
		}

		// 发送会话创建确认
		ack, _ := json.Marshal(dto.TerminalMessage{
			Type: "output",
			Data: "\r\n[终端已连接] session_id=" + session.ID + "\r\n",
		})
		_ = conn.WriteMessage(websocket.TextMessage, ack)

		// 启动输出流转发 goroutine
		go termsvc.StreamOutput(conn, stdout, session)
	}

	// 设置读取超时用于心跳检测
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		session.Touch()
		return nil
	})

	// 启动心跳 ping
	go heartbeat(conn, session)

	// 读取消息循环
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("[terminal] read error", "error", err)
			}
			break
		}
		session.Touch()

		var msg dto.TerminalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			// 非 JSON 消息，直接当作输入写入
			if session.Stdin != nil {
				_, _ = session.Stdin.Write(message)
			}
			continue
		}

		switch msg.Type {
		case "input":
			if session.Stdin != nil {
				_, _ = session.Stdin.Write([]byte(msg.Data))
			}
		case "resize":
			if session.SSHSession != nil && msg.Cols > 0 && msg.Rows > 0 {
				_ = termsvc.ResizePTY(session.SSHSession, msg.Cols, msg.Rows)
			}
		case "connect":
			// 恢复会话时前端仍会发一次 connect 参数：已建立连接，忽略即可
		case "close":
			goto cleanup
		}
	}

cleanup:
	terminalManager.Close(session.ID)
}

// connectSSHFromFirstMessage 等待并解析升级后的首条消息（SSH 连接参数），然后建立 SSH 连接。
// 带超时保护，防止客户端升级后不发参数长期占用会话。
func connectSSHFromFirstMessage(conn *websocket.Conn) (*ssh.Client, error) {
	_ = conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("未收到连接参数: %w", err)
	}
	_ = conn.SetReadDeadline(time.Time{}) // 恢复默认超时

	var req dto.CreateSessionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, fmt.Errorf("连接参数格式错误: %w", err)
	}

	return termsvc.ConnectSSH(req.Host, req.Port, req.Username, req.Password, req.KeyPath)
}

// heartbeat 定期发送 ping 保持连接
func heartbeat(conn *websocket.Conn, session *termsvc.TerminalSession) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if session.IsClosed() {
			return
		}
		session.Lock()
		err := conn.WriteMessage(websocket.PingMessage, nil)
		session.Unlock()
		if err != nil {
			return
		}
	}
}
