package deploy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// Service Git 部署（roadmap #9，PaaS 化第一步）
// webhook 接收 → 拉取仓库 → 按构建类型构建 → 发布
type Service struct {
	db       *gorm.DB
	workDir  string // 仓库克隆工作目录
	notifier Notifier
	mu       sync.Mutex // 同一 app 部署互斥
	appLocks map[uint]*sync.Mutex
}

// Notifier 部署结果通知接口
type Notifier interface {
	SendToAll(title, content string) error
}

// deployNameRegex 应用名称：会用作容器名与镜像 tag，只允许安全字符
var deployNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

// deployBranchRegex 分支名：会进入 git 命令，只允许安全字符
var deployBranchRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]{0,127}$`)

// NewService 创建部署服务
func NewService(db *gorm.DB, notifier Notifier) *Service {
	return &Service{
		db:       db,
		workDir:  "./runtime/deploy",
		notifier: notifier,
		appLocks: map[uint]*sync.Mutex{},
	}
}

func (s *Service) lockFor(appID uint) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.appLocks[appID] == nil {
		s.appLocks[appID] = &sync.Mutex{}
	}
	return s.appLocks[appID]
}

// ===== CRUD =====

func (s *Service) List() ([]model.DeployApp, error) {
	list := make([]model.DeployApp, 0)
	err := s.db.Order("created_at ASC").Find(&list).Error
	return list, err
}

func (s *Service) Create(app *model.DeployApp) error {
	if app.Name == "" || app.RepoURL == "" || app.Secret == "" {
		return fmt.Errorf("名称、仓库地址、Webhook 密钥不能为空")
	}
	if app.BuildType != "static" && app.BuildType != "node" && app.BuildType != "dockerfile" {
		return fmt.Errorf("构建类型必须为 static / node / dockerfile")
	}
	if app.BuildType == "node" && app.OutDir == "" {
		app.OutDir = "dist"
	}
	if (app.BuildType == "static" || app.BuildType == "node") && app.DestPath == "" {
		return fmt.Errorf("static/node 构建类型必须指定部署目标目录")
	}
	if app.BuildType == "dockerfile" && app.ContainerPort == 0 {
		return fmt.Errorf("dockerfile 构建类型必须指定容器端口")
	}
	if app.Branch == "" {
		app.Branch = "main"
	}
	// 路径与名称校验：这些值会进入 shell 命令与容器名，必须在入口就拦住
	// 危险字符，避免依赖下游转义兜底。
	if err := validateDeployPath(app.DestPath); err != nil {
		return err
	}
	if err := validateDeployPath(app.OutDir); err != nil {
		return err
	}
	if err := validateDeployName(app.Name); err != nil {
		return err
	}
	if err := validateDeployBranch(app.Branch); err != nil {
		return err
	}
	return s.db.Create(app).Error
}

// validateDeployPath 校验部署路径：只允许绝对路径且不含 shell 元字符。
//
// 该路径最终会出现在 `cp -r <path> <dest>` 这类 shell 命令中，
// 即便做了引号转义，也不该允许把分号、反引号等字符带进来。
func validateDeployPath(p string) error {
	if p == "" {
		return nil
	}
	if !strings.HasPrefix(p, "/") {
		return fmt.Errorf("部署路径必须是绝对路径: %s", p)
	}
	if strings.ContainsAny(p, "`$;&|<>\n\r'\"\\") {
		return fmt.Errorf("部署路径包含非法字符: %s", p)
	}
	return nil
}

// validateDeployName 校验应用名称：会用作容器名与镜像 tag
func validateDeployName(name string) error {
	if name == "" {
		return fmt.Errorf("名称不能为空")
	}
	if !deployNameRegex.MatchString(name) {
		return fmt.Errorf("应用名称只能包含字母、数字、下划线、连字符，且不能以分隔符开头: %s", name)
	}
	return nil
}

// validateDeployBranch 校验分支名：会进入 git 命令
func validateDeployBranch(branch string) error {
	if branch == "" {
		return nil
	}
	if !deployBranchRegex.MatchString(branch) {
		return fmt.Errorf("分支名只能包含字母、数字、斜杠、下划线、连字符、点: %s", branch)
	}
	return nil
}

func (s *Service) Update(id uint, req *dto.DeployAppUpdateRequest) error {
	// 与 Create 同样的入口校验：更新同样会写入危险路径
	if err := validateDeployPath(req.DestPath); err != nil {
		return err
	}
	if err := validateDeployPath(req.OutDir); err != nil {
		return err
	}
	if req.Name != "" {
		if err := validateDeployName(req.Name); err != nil {
			return err
		}
	}
	if req.Branch != "" {
		if err := validateDeployBranch(req.Branch); err != nil {
			return err
		}
	}
	updates := map[string]any{
		"name": req.Name, "repo_url": req.RepoURL, "branch": req.Branch,
		"build_type": req.BuildType, "out_dir": req.OutDir, "dest_path": req.DestPath,
		"container_port": req.ContainerPort,
	}
	// enabled 仅在前端显式传入时才更新。
	// 编辑表单没有这个开关，若无条件写入，用户改个名字应用就被静默禁用了。
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Secret != "" {
		updates["secret"] = req.Secret
	}
	return s.db.Model(&model.DeployApp{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) Delete(id uint) error {
	s.db.Where("app_id = ?", id).Delete(&model.DeployLog{})
	return s.db.Delete(&model.DeployApp{}, id).Error
}

func (s *Service) Logs(appID uint, limit int) ([]model.DeployLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	list := make([]model.DeployLog, 0)
	err := s.db.Where("app_id = ?", appID).Order("created_at DESC").Limit(limit).Find(&list).Error
	return list, err
}

// VerifyWebhook 校验 webhook 签名（HMAC-SHA256 over raw body）或纯 token 模式
// sign 为空时退化为 secret 直接比对（兼容 Gitea 简单 webhook）
func (s *Service) VerifyWebhook(app *model.DeployApp, body []byte, sign string) bool {
	if sign == "" {
		return false
	}
	if len(sign) == 64 { // hex sha256 hmac
		mac := hmac.New(sha256.New, []byte(app.Secret))
		mac.Write(body)
		expect := hex.EncodeToString(mac.Sum(nil))
		return hmac.Equal([]byte(expect), []byte(sign))
	}
	return sign == app.Secret
}

// Deploy 执行部署（同步执行，调用方可以 goroutine 包装）
func (s *Service) Deploy(appID uint) (output string, err error) {
	var app model.DeployApp
	if err = s.db.First(&app, appID).Error; err != nil {
		return "", fmt.Errorf("应用不存在")
	}

	lock := s.lockFor(appID)
	if !lock.TryLock() {
		return "", fmt.Errorf("该应用正在部署中，请稍后")
	}
	defer lock.Unlock()

	now := time.Now()
	s.db.Model(&app).Updates(map[string]any{"last_status": "deploying", "last_deploy_at": now})

	output, err = s.doDeploy(&app)

	status := "success"
	if err != nil {
		status = "failed"
		output += "\n[ERROR] " + err.Error()
	}
	s.db.Model(&app).Update("last_status", status)
	s.db.Create(&model.DeployLog{AppID: app.ID, Status: status, Output: output, CreatedAt: time.Now()})

	if s.notifier != nil {
		title := fmt.Sprintf("部署%s: %s", map[bool]string{true: "成功", false: "失败"}[err == nil], app.Name)
		_ = s.notifier.SendToAll(title, output[:min(len(output), 1500)])
	}
	return output, err
}

// doDeploy 部署主流程：clone → build → publish
func (s *Service) doDeploy(app *model.DeployApp) (string, error) {
	_ = os.MkdirAll(s.workDir, 0755)
	appDir := filepath.Join(s.workDir, fmt.Sprintf("app-%d", app.ID))

	// 1. 拉取/更新代码
	var out string
	if _, err := os.Stat(filepath.Join(appDir, ".git")); err == nil {
		if o, e := runIn(appDir, "git", "fetch", "--all"); e != nil {
			out += o
			return out, fmt.Errorf("git fetch 失败: %w", e)
		} else {
			out += o
		}
		if o, e := runIn(appDir, "git", "reset", "--hard", "origin/"+app.Branch); e != nil {
			out += o
			return out, fmt.Errorf("git reset 失败: %w", e)
		} else {
			out += o
		}
		if o, e := runIn(appDir, "git", "clean", "-fd"); e != nil {
			out += o
			return out, fmt.Errorf("git clean 失败: %w", e)
		} else {
			out += o
		}
	} else {
		_ = os.RemoveAll(appDir)
		if o, e := runIn("", "git", "clone", "--depth", "1", "-b", app.Branch, app.RepoURL, appDir); e != nil {
			out += o
			return out, fmt.Errorf("git clone 失败: %w", e)
		} else {
			out += o
		}
	}
	if !exists(filepath.Join(appDir, ".git")) {
		return out, fmt.Errorf("代码拉取失败")
	}

	// 2. 按构建类型构建并发布
	switch app.BuildType {
	case "static":
		if o, e := runIn("", "mkdir", "-p", app.DestPath); e != nil {
			out += o
			return out, fmt.Errorf("创建目标目录失败: %w", e)
		} else {
			out += o
		}
		// 不用 `|| true`：cp 失败必须让部署失败，否则会「发布失败但报成功」
		if o, e := runIn("", "sh", "-c",
			fmt.Sprintf("cp -r %s/. %s/ 2>/dev/null", shq(appDir), shq(app.DestPath))); e != nil {
			out += o
			return out, fmt.Errorf("静态文件发布失败: %w", e)
		} else {
			out += o
		}
		out += "[OK] 静态文件已发布到 " + app.DestPath

	case "node":
		if o, e := runIn(appDir, "sh", "-c", "npm ci --registry=https://registry.npmmirror.com || npm install --registry=https://registry.npmmirror.com"); e != nil {
			out += o
			return out, fmt.Errorf("依赖安装失败: %w", e)
		} else {
			out += o
		}
		if o, e := runIn(appDir, "npm", "run", "build"); e != nil {
			out += o
			return out, fmt.Errorf("构建失败: %w", e)
		} else {
			out += o
		}
		buildOut := filepath.Join(appDir, app.OutDir)
		if !exists(buildOut) {
			return out, fmt.Errorf("构建产物目录 %s 不存在", app.OutDir)
		}
		if o, e := runIn("", "mkdir", "-p", app.DestPath); e != nil {
			out += o
			return out, fmt.Errorf("创建目标目录失败: %w", e)
		} else {
			out += o
		}
		if o, e := runIn("", "sh", "-c",
			fmt.Sprintf("cp -r %s/. %s/ 2>/dev/null", shq(buildOut), shq(app.DestPath))); e != nil {
			out += o
			return out, fmt.Errorf("构建产物发布失败: %w", e)
		} else {
			out += o
		}
		out += "[OK] 构建产物已发布到 " + app.DestPath

	case "dockerfile":
		image := fmt.Sprintf("deployapp-%s:latest", app.Name)
		if o, e := runIn(appDir, "docker", "build", "-t", image, "."); e != nil {
			out += o
			return out, fmt.Errorf("镜像构建失败: %w", e)
		} else {
			out += o
		}
		// 替换旧容器：旧容器不存在是正常情况，不算失败
		out += runIgnoreError("", "sh", "-c",
			fmt.Sprintf("docker rm -f %s 2>/dev/null; true", shq("deployapp-"+app.Name)))
		if o, e := runIn("", "docker", "run", "-d", "--name", "deployapp-"+app.Name,
			"--restart", "unless-stopped",
			"-p", fmt.Sprintf("%d:%d", app.ContainerPort, app.ContainerPort), image); e != nil {
			out += o
			return out, fmt.Errorf("容器启动失败: %w", e)
		} else {
			out += o
		}
		out += fmt.Sprintf("[OK] 容器已启动: deployapp-%s (端口 %d)", app.Name, app.ContainerPort)
	}

	slog.Info("[deploy] 部署完成", "app", app.Name)
	return out, nil
}

// ===== 工具函数 =====

// runIn 执行命令并返回输出。失败时返回 error（输出仍然带回，便于排查）。
//
// 早期版本只返回输出字符串、丢弃 error，导致构建失败时上层仍收到 nil，
// 部署被记为「成功」—— 属于「失败报成功」的根因。
func runIn(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	b, err := cmd.CombinedOutput()
	out := fmt.Sprintf("\n$ %s %v\n%s", name, args, string(b))
	if err != nil {
		out += fmt.Sprintf(" [exit: %v]", err)
		return out, err
	}
	return out, nil
}

// runIgnoreError 执行命令并忽略退出码（用于「失败也可接受」的场景，
// 如删除可能不存在的旧容器），输出仍会带回。
func runIgnoreError(dir, name string, args ...string) string {
	out, _ := runIn(dir, name, args...)
	return out
}

// shq 用单引号包裹 shell 参数。
//
// POSIX 单引号内不能出现单引号，正确做法是把 ' 写成 '\”（结束引号、
// 转义的单引号、再开引号）。早期版本直接用 ' 包住原串，参数里带单引号时
// 会逃逸出引号造成命令注入。
func shq(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
