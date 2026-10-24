package terminal

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"runtime"
	"time"

	"cloudpivot/internal/dto"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

const (
	defaultSSHHost = "localhost"
	defaultSSHPort = "22"
	sshTimeout     = 10 * time.Second
)

// ConnectSSH 建立 SSH 连接
// 支持密码认证和私钥认证（keyPath 优先）
func ConnectSSH(host string, port int, username, password, keyPath string) (*ssh.Client, error) {
	if host == "" {
		host = defaultSSHHost
	}
	if port == 0 {
		port = 22
	}
	if username == "" {
		if u, err := user.Current(); err == nil {
			username = u.Username
		}
	}

	var authMethods []ssh.AuthMethod

	// 私钥认证优先
	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("read private key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// 密码认证
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method provided (password or key_path required)")
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // MVP：不校验主机密钥
		Timeout:         sshTimeout,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		// Windows 开发环境下 SSH 通常不可用，给出明确提示
		if runtime.GOOS == "windows" {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return nil, fmt.Errorf("SSH 连接超时 (%s)：Windows 开发环境默认未启动 SSH 服务，请安装 OpenSSH Server 或在 Linux 环境运行", addr)
			}
			if opErr, ok := err.(*net.OpError); ok {
				return nil, fmt.Errorf("SSH 连接失败 (%s)：%v（Windows 开发环境需启用 OpenSSH Server，或部署到 Linux 目标机）", addr, opErr.Err)
			}
		}
		return nil, fmt.Errorf("ssh dial %s: %w", addr, err)
	}

	return client, nil
}

// StartPTY 请求 PTY 并启动交互式 shell
// 返回 stdin 写入器和 stdout/stderr 合并读取器
func StartPTY(sshClient *ssh.Client, cols, rows int) (*ssh.Session, io.WriteCloser, io.Reader, error) {
	session, err := sshClient.NewSession()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("new ssh session: %w", err)
	}

	// 请求 PTY
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}

	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		session.Close()
		return nil, nil, nil, fmt.Errorf("request pty: %w", err)
	}

	// 设置 stdin
	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, nil, nil, fmt.Errorf("stdin pipe: %w", err)
	}

	// 合并 stdout 和 stderr（PTY 模式下 stderr 自动合并到 stdout）
	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, nil, nil, fmt.Errorf("stdout pipe: %w", err)
	}

	// 启动 shell
	if err := session.Shell(); err != nil {
		session.Close()
		return nil, nil, nil, fmt.Errorf("start shell: %w", err)
	}

	return session, stdin, stdout, nil
}

// ResizePTY 调整 PTY 窗口大小
func ResizePTY(session *ssh.Session, cols, rows int) error {
	if session == nil {
		return fmt.Errorf("session is nil")
	}
	if cols <= 0 || rows <= 0 {
		return fmt.Errorf("invalid dimensions: cols=%d rows=%d", cols, rows)
	}
	return session.WindowChange(rows, cols)
}

// StreamOutput 将 SSH stdout 持续推送到 WebSocket
// 阻塞运行，直到输出结束或连接关闭
// RawMode（ZMODEM 透传）时使用 BinaryMessage 原样转发；RecordFile 开启时同步录制
func StreamOutput(conn *websocket.Conn, stdout io.Reader, session *TerminalSession) {
	buf := make([]byte, 4096)
	for {
		if session.IsClosed() {
			return
		}
		n, err := stdout.Read(buf)
		if n > 0 {
			session.Touch()

			// 会话录制（T-006）
			session.mu.Lock()
			recordFile := session.RecordFile
			session.mu.Unlock()
			if recordFile != nil {
				_, _ = recordFile.Write(buf[:n])
			}

			session.mu.Lock()
			closed := session.closed
			rawMode := session.RawMode
			session.mu.Unlock()
			if closed {
				return
			}

			if rawMode {
				// ZMODEM 原始字节透传（T-005）
				if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
					return
				}
				continue
			}

			msg := dto.TerminalMessage{
				Type: "output",
				Data: string(buf[:n]),
			}
			payload, _ := json.Marshal(msg)
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		}
		if err != nil {
			if err != io.EOF {
				errMsg := dto.TerminalMessage{
					Type: "error",
					Data: err.Error(),
				}
				payload, _ := json.Marshal(errMsg)
				_ = conn.WriteMessage(websocket.TextMessage, payload)
			}
			return
		}
	}
}
