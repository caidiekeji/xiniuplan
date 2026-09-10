package database

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

const (
	defaultMySQLPortStart    = 3306
	defaultRedisPortStart    = 6379
	defaultPostgresPortStart = 5432

	defaultMySQLImage57 = "mysql:5.7"
	defaultMySQLImage80 = "mysql:8.0"

	defaultRedisImage6 = "redis:6-alpine"
	defaultRedisImage7 = "redis:7-alpine"

	defaultPostgresImage14 = "postgres:14"
	defaultPostgresImage15 = "postgres:15"
	defaultPostgresImage16 = "postgres:16"

	healthPollInterval = 2 * time.Second
	healthPollTimeout  = 30 * time.Second
)

// containerNamePrefix 根据类型返回容器名前缀
func containerNamePrefix(typ string) string {
	return "cloudpivot-" + typ + "-"
}

// InstanceService 数据库实例管理服务
type InstanceService struct {
	db         *gorm.DB
	executor   *common.CommandExecutor
	runtimeDir string
}

// NewInstanceService 创建实例管理服务
func NewInstanceService(db *gorm.DB, runtimeDir string) *InstanceService {
	exec := common.NewCommandExecutor(60 * time.Second)
	registerDockerTemplates(exec)
	return &InstanceService{
		db:         db,
		executor:   exec,
		runtimeDir: runtimeDir,
	}
}

// registerDockerTemplates 注册 Docker 命令白名单模板
func registerDockerTemplates(exec *common.CommandExecutor) {
	exec.Register(&common.CommandTemplate{
		Name: "docker_pull",
		Bin:  "docker",
		Args: []string{"pull", "{{image}}"},
	})

	// MySQL
	exec.Register(&common.CommandTemplate{
		Name: "docker_run_mysql",
		Bin:  "docker",
		Args: []string{
			"run", "-d",
			"--name", "{{name}}",
			"-p", "{{port}}:3306",
			"-e", "MYSQL_ROOT_PASSWORD={{password}}",
			"-v", "{{datadir}}:/var/lib/mysql",
			"--restart", "unless-stopped",
			"{{image}}",
		},
	})
	exec.Register(&common.CommandTemplate{
		Name: "docker_exec_mysql_ping",
		Bin:  "docker",
		Args: []string{
			"exec", "{{name}}",
			"mysqladmin", "ping", "-h", "127.0.0.1", "-uroot", "-p{{password}}",
		},
	})

	// Redis
	exec.Register(&common.CommandTemplate{
		Name: "docker_run_redis",
		Bin:  "docker",
		Args: []string{
			"run", "-d",
			"--name", "{{name}}",
			"-p", "{{port}}:6379",
			"-v", "{{datadir}}:/data",
			"--restart", "unless-stopped",
			"{{image}}",
			"redis-server",
			"--requirepass", "{{password}}",
			"--appendonly", "yes",
		},
	})
	exec.Register(&common.CommandTemplate{
		Name: "docker_exec_redis_ping",
		Bin:  "docker",
		Args: []string{
			"exec", "{{name}}",
			"redis-cli", "-a", "{{password}}", "ping",
		},
	})

	// PostgreSQL
	exec.Register(&common.CommandTemplate{
		Name: "docker_run_postgres",
		Bin:  "docker",
		Args: []string{
			"run", "-d",
			"--name", "{{name}}",
			"-p", "{{port}}:5432",
			"-e", "POSTGRES_PASSWORD={{password}}",
			"-v", "{{datadir}}:/var/lib/postgresql/data",
			"--restart", "unless-stopped",
			"{{image}}",
		},
	})
	exec.Register(&common.CommandTemplate{
		Name: "docker_exec_postgres_ping",
		Bin:  "docker",
		Args: []string{
			"exec", "{{name}}",
			"pg_isready", "-h", "127.0.0.1", "-U", "postgres",
		},
	})

	// 通用容器操作
	exec.Register(&common.CommandTemplate{
		Name: "docker_start",
		Bin:  "docker",
		Args: []string{"start", "{{name}}"},
	})
	exec.Register(&common.CommandTemplate{
		Name: "docker_stop",
		Bin:  "docker",
		Args: []string{"stop", "{{name}}"},
	})
	exec.Register(&common.CommandTemplate{
		Name: "docker_restart",
		Bin:  "docker",
		Args: []string{"restart", "{{name}}"},
	})
	exec.Register(&common.CommandTemplate{
		Name: "docker_rm",
		Bin:  "docker",
		Args: []string{"rm", "-f", "{{name}}"},
	})
	exec.Register(&common.CommandTemplate{
		Name: "docker_inspect_status",
		Bin:  "docker",
		Args: []string{"inspect", "-f", "{{.State.Status}}", "{{name}}"},
	})
}

// resolveImage 根据类型和版本确定镜像
func resolveImage(typ, version string) (string, error) {
	switch typ {
	case "mysql":
		switch version {
		case "", "8.0":
			return defaultMySQLImage80, nil
		case "5.7":
			return defaultMySQLImage57, nil
		default:
			return "", fmt.Errorf("不支持的 MySQL 版本: %s", version)
		}
	case "redis":
		switch version {
		case "", "7":
			return defaultRedisImage7, nil
		case "6":
			return defaultRedisImage6, nil
		default:
			return "", fmt.Errorf("不支持的 Redis 版本: %s", version)
		}
	case "postgres":
		switch version {
		case "", "16":
			return defaultPostgresImage16, nil
		case "15":
			return defaultPostgresImage15, nil
		case "14":
			return defaultPostgresImage14, nil
		default:
			return "", fmt.Errorf("不支持的 PostgreSQL 版本: %s", version)
		}
	default:
		return "", fmt.Errorf("不支持的数据库类型: %s", typ)
	}
}

// defaultPortStart 根据类型返回默认端口起始值
func defaultPortStart(typ string) int {
	switch typ {
	case "redis":
		return defaultRedisPortStart
	case "postgres":
		return defaultPostgresPortStart
	default:
		return defaultMySQLPortStart
	}
}

// runTemplateName 根据类型返回 docker run 模板名
func runTemplateName(typ string) string {
	return "docker_run_" + typ
}

// pingTemplateName 根据类型返回健康检查模板名
func pingTemplateName(typ string) string {
	return "docker_exec_" + typ + "_ping"
}

// CreateInstance 创建数据库实例
func (s *InstanceService) CreateInstance(req dto.CreateDBInstanceRequest) (*model.DatabaseInstance, error) {
	// 确定镜像
	image, err := resolveImage(req.Type, req.Version)
	if err != nil {
		return nil, err
	}

	// 分配端口
	port := req.Port
	if port == 0 {
		var err error
		port, err = s.allocatePort(defaultPortStart(req.Type))
		if err != nil {
			return nil, fmt.Errorf("分配端口失败: %w", err)
		}
	} else if !isPortAvailable(port) {
		return nil, fmt.Errorf("端口 %d 已被占用", port)
	}

	// 生成密码
	password := req.RootPassword
	if password == "" {
		var err error
		password, err = generateRandomPassword(16)
		if err != nil {
			return nil, fmt.Errorf("生成密码失败: %w", err)
		}
	}

	// 先创建记录获取 ID
	instance := &model.DatabaseInstance{
		Type:         req.Type,
		Version:      req.Version,
		Port:         port,
		RootPassword: password,
		Status:       "creating",
	}
	if err := s.db.Create(instance).Error; err != nil {
		return nil, fmt.Errorf("创建实例记录失败: %w", err)
	}

	containerName := fmt.Sprintf("%s%d", containerNamePrefix(req.Type), instance.ID)
	dataDir := filepath.Join(s.runtimeDir, req.Type, strconv.Itoa(int(instance.ID)), "data")

	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		s.markError(instance.ID, "创建数据目录失败: "+err.Error())
		return instance, nil
	}

	ctx := context.Background()

	// 拉取镜像（失败不阻断，可能本地已有）
	if _, err := s.executor.Execute(ctx, "docker_pull", map[string]string{"image": image}); err != nil {
		slog.Warn("[DBInstance] 拉取镜像失败（将尝试使用本地镜像）", "image", image, "error", err)
	}

	// 启动容器
	result, err := s.executor.Execute(ctx, runTemplateName(req.Type), map[string]string{
		"name":     containerName,
		"port":     strconv.Itoa(port),
		"password": password,
		"datadir":  dataDir,
		"image":    image,
	})
	if err != nil {
		s.markError(instance.ID, "启动容器失败: "+err.Error())
		return instance, nil
	}
	if result.ExitCode != 0 {
		s.markError(instance.ID, "启动容器失败: "+result.Stderr)
		return instance, nil
	}

	containerID := strings.TrimSpace(result.Stdout)

	// 更新容器 ID
	s.db.Model(instance).Updates(map[string]interface{}{
		"container_id": containerID,
		"status":       "starting",
	})

	// 等待健康
	go s.waitForHealth(instance.ID, req.Type, containerName, password)

	return instance, nil
}

// waitForHealth 轮询等待容器健康
func (s *InstanceService) waitForHealth(instanceID uint, typ, containerName, password string) {
	deadline := time.Now().Add(healthPollTimeout)
	pingTpl := pingTemplateName(typ)

	for time.Now().Before(deadline) {
		result, err := s.executor.Execute(context.Background(), pingTpl, map[string]string{
			"name":     containerName,
			"password": password,
		})
		if err == nil && result.ExitCode == 0 {
			alive := false
			switch typ {
			case "mysql":
				alive = strings.Contains(result.Stdout, "alive")
			case "redis":
				alive = strings.Contains(result.Stdout, "PONG")
			case "postgres":
				alive = strings.Contains(result.Stdout, "accepting connections")
			default:
				alive = true
			}
			if alive {
				s.db.Model(&model.DatabaseInstance{}).Where("id = ?", instanceID).Update("status", "running")
				slog.Info("[DBInstance] 实例已就绪", "id", instanceID, "type", typ)
				return
			}
		}
		time.Sleep(healthPollInterval)
	}
	slog.Warn("[DBInstance] 实例健康检查超时", "id", instanceID, "type", typ)
	s.db.Model(&model.DatabaseInstance{}).Where("id = ?", instanceID).Update("status", "error")
}

// markError 标记实例为错误状态
func (s *InstanceService) markError(id uint, msg string) {
	slog.Error("[DBInstance] 实例错误", "id", id, "error", msg)
	s.db.Model(&model.DatabaseInstance{}).Where("id = ?", id).Update("status", "error")
}

// ListInstances 分页查询实例列表
func (s *InstanceService) ListInstances(page, pageSize int) ([]model.DatabaseInstance, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	instances := make([]model.DatabaseInstance, 0)
	var total int64

	if err := s.db.Model(&model.DatabaseInstance{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := s.db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&instances).Error; err != nil {
		return nil, 0, err
	}
	return instances, total, nil
}

// GetInstance 获取实例详情
func (s *InstanceService) GetInstance(id uint) (*model.DatabaseInstance, error) {
	var instance model.DatabaseInstance
	if err := s.db.First(&instance, id).Error; err != nil {
		return nil, fmt.Errorf("实例不存在: %d", id)
	}
	return &instance, nil
}

// StartInstance 启动实例
func (s *InstanceService) StartInstance(id uint) error {
	instance, err := s.GetInstance(id)
	if err != nil {
		return err
	}
	containerName := fmt.Sprintf("%s%d", containerNamePrefix(instance.Type), instance.ID)
	result, err := s.executor.Execute(context.Background(), "docker_start", map[string]string{"name": containerName})
	if err != nil {
		return fmt.Errorf("启动容器失败: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("启动容器失败: %s", result.Stderr)
	}
	s.db.Model(instance).Update("status", "running")
	return nil
}

// StopInstance 停止实例
func (s *InstanceService) StopInstance(id uint) error {
	instance, err := s.GetInstance(id)
	if err != nil {
		return err
	}
	containerName := fmt.Sprintf("%s%d", containerNamePrefix(instance.Type), instance.ID)
	result, err := s.executor.Execute(context.Background(), "docker_stop", map[string]string{"name": containerName})
	if err != nil {
		return fmt.Errorf("停止容器失败: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("停止容器失败: %s", result.Stderr)
	}
	s.db.Model(instance).Update("status", "stopped")
	return nil
}

// RestartInstance 重启实例
func (s *InstanceService) RestartInstance(id uint) error {
	instance, err := s.GetInstance(id)
	if err != nil {
		return err
	}
	containerName := fmt.Sprintf("%s%d", containerNamePrefix(instance.Type), instance.ID)
	result, err := s.executor.Execute(context.Background(), "docker_restart", map[string]string{"name": containerName})
	if err != nil {
		return fmt.Errorf("重启容器失败: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("重启容器失败: %s", result.Stderr)
	}
	s.db.Model(instance).Update("status", "running")
	return nil
}

// DeleteInstance 删除实例
func (s *InstanceService) DeleteInstance(id uint, cleanupData bool) error {
	instance, err := s.GetInstance(id)
	if err != nil {
		return err
	}

	// 外部导入的实例只删除记录，不操作 Docker
	if !instance.IsExternal {
		containerName := fmt.Sprintf("%s%d", containerNamePrefix(instance.Type), instance.ID)

		// 强制删除容器
		result, err := s.executor.Execute(context.Background(), "docker_rm", map[string]string{"name": containerName})
		if err != nil {
			slog.Warn("[DBInstance] 删除容器失败", "container", containerName, "error", err)
		} else if result.ExitCode != 0 {
			slog.Warn("[DBInstance] 删除容器失败", "container", containerName, "stderr", result.Stderr)
		}

		// 清理数据目录
		if cleanupData {
			dataDir := filepath.Join(s.runtimeDir, instance.Type, strconv.Itoa(int(instance.ID)))
			os.RemoveAll(dataDir)
		}
	}

	// 删除数据库记录
	return s.db.Delete(instance).Error
}

// GetInstanceStatus 获取实例状态
func (s *InstanceService) GetInstanceStatus(id uint) (string, error) {
	instance, err := s.GetInstance(id)
	if err != nil {
		return "", err
	}

	// 外部导入的实例：直接做 TCP 连接检测
	if instance.IsExternal {
		host := instance.Host
		if host == "" {
			host = "127.0.0.1"
		}
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, instance.Port), 3*time.Second)
		if err != nil {
			return "stopped", nil
		}
		conn.Close()
		return "running", nil
	}

	// Docker 实例：查询容器状态
	containerName := fmt.Sprintf("%s%d", containerNamePrefix(instance.Type), instance.ID)
	result, err := s.executor.Execute(context.Background(), "docker_inspect_status", map[string]string{"name": containerName})
	if err != nil {
		return "unknown", fmt.Errorf("获取容器状态失败: %w", err)
	}
	if result.ExitCode != 0 {
		return "not_found", nil
	}
	return strings.TrimSpace(result.Stdout), nil
}

// allocatePort 从 startPort 开始查找可用端口
func (s *InstanceService) allocatePort(startPort int) (int, error) {
	// 获取已占用端口
	var usedPorts []int
	s.db.Model(&model.DatabaseInstance{}).Pluck("port", &usedPorts)
	usedSet := make(map[int]bool)
	for _, p := range usedPorts {
		usedSet[p] = true
	}

	for port := startPort; port < startPort+1000; port++ {
		if usedSet[port] {
			continue
		}
		if isPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("未找到可用端口（范围 %d-%d）", startPort, startPort+999)
}

// isPortAvailable 检测端口是否可用
func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// generateRandomPassword 生成指定长度的随机密码
func generateRandomPassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// ImportInstance 导入已有的数据库实例（不创建Docker容器）
func (s *InstanceService) ImportInstance(req dto.ImportDBInstanceRequest) (*model.DatabaseInstance, error) {
	// 验证类型
	if req.Type != "mysql" && req.Type != "postgres" && req.Type != "redis" {
		return nil, fmt.Errorf("不支持的数据库类型: %s", req.Type)
	}

	// 检测连接
	if !testDBConnection(req.Type, req.Host, req.Port, req.Username, req.Password) {
		return nil, fmt.Errorf("无法连接到 %s:%d，请检查地址、端口和凭据", req.Host, req.Port)
	}

	// 自动探测版本
	version := req.Version
	if version == "" {
		version = detectDBVersion(req.Type, req.Host, req.Port, req.Username, req.Password)
	}

	instance := &model.DatabaseInstance{
		Type:         req.Type,
		Version:      version,
		Host:         req.Host,
		Port:         req.Port,
		Username:     req.Username,
		RootPassword: req.Password,
		IsExternal:   true,
		Status:       "running",
	}
	if err := s.db.Create(instance).Error; err != nil {
		return nil, fmt.Errorf("创建实例记录失败: %w", err)
	}
	return instance, nil
}

// DiscoverInstances 自动扫描本机已安装的数据库服务
func (s *InstanceService) DiscoverInstances() ([]dto.DiscoveredDB, error) {
	result := make([]dto.DiscoveredDB, 0)

	// 扫描常见端口
	targets := []struct {
		Type string
		Port int
	}{
		{"mysql", 3306},
		{"postgres", 5432},
		{"redis", 6379},
	}

	for _, t := range targets {
		// 检查端口是否在监听
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", t.Port), 2*time.Second)
		if err != nil {
			continue
		}
		conn.Close()

		// 尝试探测版本（无凭据，可能失败）
		version := detectDBVersion(t.Type, "127.0.0.1", t.Port, "", "")

		result = append(result, dto.DiscoveredDB{
			Type:    t.Type,
			Host:    "127.0.0.1",
			Port:    t.Port,
			Version: version,
			Running: true,
		})
	}

	return result, nil
}

// testDBConnection 测试数据库连接
func testDBConnection(typ, host string, port int, username, password string) bool {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// detectDBVersion 探测数据库版本（尽力而为，失败返回空）
func detectDBVersion(typ, host string, port int, username, password string) string {
	// 简化版：只做端口连通性检测，版本通过命令行工具获取
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cmd string
	switch typ {
	case "mysql":
		cmd = fmt.Sprintf("mysql -h%s -P%d -u%s -p%s -e 'SELECT VERSION();' 2>/dev/null | tail -1", host, port, username, password)
	case "postgres":
		cmd = fmt.Sprintf("psql -h%s -p%d -U%s -t -c 'SHOW server_version;' 2>/dev/null", host, port, username)
	case "redis":
		cmd = fmt.Sprintf("redis-cli -h %s -p %d INFO server 2>/dev/null | grep redis_version | cut -d: -f2", host, port)
	default:
		return ""
	}

	exec := common.NewCommandExecutor(5 * time.Second)
	result, err := exec.Execute(ctx, "shell", map[string]string{"command": cmd})
	if err != nil || result.ExitCode != 0 {
		return ""
	}
	return strings.TrimSpace(result.Stdout)
}
