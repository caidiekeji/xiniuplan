package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	dockerpkg "cloudpivot/pkg/docker"

	"github.com/gin-gonic/gin"
)

// RegisterImagePullRoutes 注册镜像拉取路由
//
// 拉取是长耗时操作（大镜像几分钟到几十分钟），必须流式返回进度，
// 否则前端只能拿到一个挂死的请求 —— 与容器日志跟随同一个道理。
func RegisterImagePullRoutes(r *gin.RouterGroup) {
	r.GET("/images/registries", middleware.RequirePermission(model.PermContainerView), ListRegistries)
	r.POST("/images/pull", middleware.RequirePermission(model.PermContainerUpdate), PullImage)
}

// pullRequest 拉取请求
type pullRequest struct {
	// Image 镜像引用，如 nginx:1.25
	Image string `json:"image"`
	// Mirror 使用的加速源；空表示直连（或走设置里的第一个加速源）
	Mirror string `json:"mirror"`
	// Username / Password 私有仓库凭据（可选）
	Username string `json:"username"`
	Password string `json:"password"`
}

// ListRegistries GET /images/registries - 获取可用的加速源列表
//
// 返回设置里配置的加速源 + 直连选项，供拉取弹窗下拉选择。
func ListRegistries(c *gin.Context) {
	mirrors := dockerpkg.ParseMirrorList(settingValue(c, "docker_registry_mirrors"))
	private := strings.TrimSpace(settingValue(c, "docker_registry_private"))

	// 直连放在最前：配了加速源也不代表用户一定想走加速
	options := []dockerpkg.RegistryOption{
		{ID: "", Name: "直连 Docker Hub"},
	}
	for _, m := range mirrors {
		options = append(options, dockerpkg.RegistryOption{ID: m, Name: m})
	}
	if private != "" {
		options = append(options, dockerpkg.RegistryOption{ID: "@private", Name: private + "（私有仓库）"})
	}

	common.Success(c, gin.H{
		"registries":   options,
		"has_private":  private != "",
		"private_addr": private,
		"private_user": strings.TrimSpace(settingValue(c, "docker_registry_user")),
	})
}

// settingValue 读取单个设置值（不存在的键返回空串）
func settingValue(c *gin.Context, key string) string {
	if systemSvcRef == nil {
		return ""
	}
	settings, err := systemSvcRef.GetSettings()
	if err != nil {
		return ""
	}
	return settings[key]
}

// PullImage POST /images/pull - 拉取镜像（SSE 流式进度）
//
// 响应为 text/event-stream，每帧是一个 PullProgress JSON：
//
//	{"layer_id":"a1b2c3","status":"Downloading","current":1024,"total":2048,"percent":50}
//
// 最后一帧 status 为 "Pull complete" 或 "Status: Downloaded newer image"。
// 出错时发送 {"error":"..."} 帧后关闭。
func PullImage(c *gin.Context) {
	var req pullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Image) == "" {
		common.Fail(c, common.CodeParamError, "镜像名不能为空")
		return
	}

	// 解析镜像名，按需改写为加速源地址
	parsed, err := dockerpkg.ParseImageRef(req.Image)
	if err != nil {
		common.Fail(c, common.CodeParamError, err.Error())
		return
	}

	imageRef := parsed.String()
	if m := strings.TrimSpace(req.Mirror); m != "" && m != "@private" {
		rewritten, err := dockerpkg.ApplyMirror(parsed, m)
		if err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		imageRef = rewritten.String()
	}

	// 凭据：请求里带的优先；选私有仓库且未带凭据时，回落到设置里的账号
	username := strings.TrimSpace(req.Username)
	password := req.Password
	if username == "" && strings.TrimSpace(req.Mirror) == "@private" {
		username = strings.TrimSpace(settingValue(c, "docker_registry_user"))
		password = settingValue(c, "docker_registry_pass")
	}
	authHeader := dockerpkg.RegistryAuth(username, password)

	svc := containerSvc()
	if svc == nil {
		common.Fail(c, common.CodeServerError, "容器服务未初始化")
		return
	}

	// 流式响应
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		common.Fail(c, common.CodeServerError, "streaming not supported")
		return
	}

	// 解除本连接的写超时：http.Server.WriteTimeout 默认 60s，会从读请求头起
	// 覆盖整个响应生命周期，SSE 长连接（拉取大镜像可能几十分钟）会在 60s 被掐断。
	// 用 NewResponseController 把本连接的写截止时间清零，改由请求 ctx（客户端
	// 断开即取消）与心跳管理连接生命周期。必须在任何写操作之前调用。
	if rc := http.NewResponseController(c.Writer); rc != nil {
		_ = rc.SetWriteDeadline(time.Time{})
	}

	var mu sync.Mutex
	writeEvent := func(v interface{}) {
		mu.Lock()
		defer mu.Unlock()
		b, err := json.Marshal(v)
		if err != nil {
			return
		}
		fmt.Fprintf(c.Writer, "data: %s\n\n", b)
		flusher.Flush()
	}

	// 心跳：拉取大镜像时可能几十秒没有进度消息，需要定期发空帧
	// 保持连接不被中间层（nginx 等）断开
	stopBeat := make(chan struct{})
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				writeEvent(gin.H{"status": "pulling"})
			case <-stopBeat:
				return
			}
		}
	}()
	defer close(stopBeat)

	result, pullErr := svc.PullImageWithProgress(c.Request.Context(), imageRef, authHeader, func(p dockerpkg.PullProgress) {
		writeEvent(p)
	})

	if pullErr != nil {
		writeEvent(gin.H{"error": pullErr.Error()})
		return
	}
	writeEvent(gin.H{
		"status":    "Pull complete",
		"image_ref": result.ImageRef,
		"message":   fmt.Sprintf("镜像 %s 拉取完成", imageRef),
	})
}
