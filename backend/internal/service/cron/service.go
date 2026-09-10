package cron

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"
	backupsvc "cloudpivot/internal/service/backup"
	cronengine "cloudpivot/pkg/cron"

	"gorm.io/gorm"
)

// Service 计划任务服务
type Service struct {
	db        *gorm.DB
	engine    *cronengine.CronEngine
	backupSvc *backupsvc.Service
	executor  *common.CommandExecutor
}

// NewService 创建计划任务服务
func NewService(db *gorm.DB) *Service {
	exec := common.NewCommandExecutor(120 * time.Second)
	registerCronTemplates(exec)
	return &Service{
		db:        db,
		backupSvc: backupsvc.NewService(db),
		executor:  exec,
	}
}

// registerCronTemplates 注册计划任务命令白名单模板
func registerCronTemplates(exec *common.CommandExecutor) {
	exec.Register(&common.CommandTemplate{
		Name:        "docker_restart",
		Description: "Restart a docker container",
		Bin:         "docker",
		Args:        []string{"restart", "{{container}}"},
		AllowedArgs: map[string]bool{"container": true},
	})
}

// InitCronEngine 初始化并启动 Cron 引擎，从 DB 加载所有 enabled 任务
// 由 main.go 在集成时调用
func InitCronEngine(db *gorm.DB) *cronengine.CronEngine {
	svc := NewService(db)
	engine := cronengine.NewEngine()
	svc.engine = engine

	// 加载所有 enabled 任务
	var tasks []model.CronTask
	if err := db.Where("status = ?", "enabled").Find(&tasks).Error; err != nil {
		fmt.Printf("[cron-service] load enabled tasks failed: %v\n", err)
	}

	for i := range tasks {
		task := tasks[i]
		handler := svc.makeTaskHandler(&task)
		entryID, err := engine.AddTask(int(task.ID), task.Schedule, handler)
		if err != nil {
			fmt.Printf("[cron-service] register task #%d (%s) failed: %v\n", task.ID, task.Name, err)
			continue
		}
		task.EntryID = entryID
		// 更新下次执行时间
		if next := engine.NextRunTime(int(task.ID)); next != nil {
			task.NextRunAt = next
		}
		db.Save(&task)
	}

	engine.Start()
	fmt.Printf("[cron-service] cron engine started, %d tasks loaded\n", len(tasks))
	return engine
}

// makeTaskHandler 构建任务执行函数
func (s *Service) makeTaskHandler(task *model.CronTask) func() {
	return func() {
		s.runTask(task)
	}
}

// runTask 执行任务并记录日志
func (s *Service) runTask(task *model.CronTask) {
	now := time.Now()

	logEntry := &model.CronLog{
		TaskID:    task.ID,
		StartTime: now,
		Result:    "running",
	}
	s.db.Create(logEntry)

	// 更新任务最后执行时间
	task.LastRunAt = &now
	s.db.Model(&model.CronTask{}).Where("id = ?", task.ID).Update("last_run_at", now)

	var output string
	var execErr error

	switch task.Type {
	case "shell":
		output, execErr = s.execShell(task.Command)
	case "backup":
		output, execErr = s.execBackup(task)
	case "url":
		output, execErr = s.execURL(task.Target)
	case "app_restart":
		output, execErr = s.execAppRestart(task.Target)
	default:
		execErr = fmt.Errorf("unknown task type: %s", task.Type)
	}

	endTime := time.Now()
	result := "success"
	if execErr != nil {
		result = "failed"
		output = fmt.Sprintf("%s\nError: %v", output, execErr)
	}

	// 更新日志
	s.db.Model(logEntry).Updates(map[string]interface{}{
		"end_time": endTime,
		"output":   output,
		"result":   result,
	})

	// 更新任务下次执行时间
	if s.engine != nil {
		if next := s.engine.NextRunTime(int(task.ID)); next != nil {
			s.db.Model(&model.CronTask{}).Where("id = ?", task.ID).Update("next_run_at", next)
		}
	}
}

// execShell 执行 shell 命令
func (s *Service) execShell(command string) (string, error) {
	if command == "" {
		return "", errors.New("command is empty")
	}

	// 基本安全校验：禁止危险字符
	if err := common.ValidateInput(command); err != nil {
		return "", fmt.Errorf("command contains forbidden characters: %w", err)
	}

	// 分割命令为参数（简单空格分割）
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", errors.New("invalid command")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n[stderr]\n" + stderr.String()
	}

	if err != nil {
		return output, fmt.Errorf("shell command failed: %w", err)
	}
	return output, nil
}

// execBackup 执行备份任务
func (s *Service) execBackup(task *model.CronTask) (string, error) {
	backupType := task.Target
	if backupType == "" {
		backupType = "system"
	}

	req := dto.CreateBackupRequest{
		Type:   backupType,
		Target: task.Command, // command 字段存放具体目标（域名/数据库名）
		Remark: fmt.Sprintf("auto backup by cron task #%d", task.ID),
	}

	backup, err := s.backupSvc.CreateBackup(req)
	if err != nil {
		return "", fmt.Errorf("create backup failed: %w", err)
	}

	return fmt.Sprintf("backup task #%d created (type=%s, target=%s)", backup.ID, backup.Type, backup.Target), nil
}

// execURL 执行 HTTP GET 请求
func (s *Service) execURL(url string) (string, error) {
	if url == "" {
		return "", errors.New("url is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 最多读取 1MB
	if err != nil {
		return "", fmt.Errorf("read response failed: %w", err)
	}

	output := fmt.Sprintf("HTTP %d\n%s", resp.StatusCode, string(body))
	if resp.StatusCode >= 400 {
		return output, fmt.Errorf("http status %d", resp.StatusCode)
	}
	return output, nil
}

// execAppRestart 重启 Docker 容器
func (s *Service) execAppRestart(container string) (string, error) {
	if container == "" {
		return "", errors.New("container name is empty")
	}

	result, err := s.executor.Execute(context.Background(), "docker_restart", map[string]string{
		"container": container,
	})
	if err != nil {
		return "", fmt.Errorf("docker restart failed: %w", err)
	}
	if result.ExitCode != 0 {
		return result.Stderr, fmt.Errorf("docker restart exit code %d", result.ExitCode)
	}
	return result.Stdout, nil
}

// --- CRUD ---

// CreateTask 创建任务，enabled 则注册到 cron 引擎
func (s *Service) CreateTask(req dto.CreateCronRequest) (*model.CronTask, error) {
	task := &model.CronTask{
		Name:     req.Name,
		Type:     req.Type,
		Schedule: req.Schedule,
		Command:  req.Command,
		Target:   req.Target,
		Remark:   req.Remark,
		Status:   "enabled",
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("create cron task: %w", err)
	}

	// 注册到 cron 引擎
	if s.engine != nil {
		handler := s.makeTaskHandler(task)
		entryID, err := s.engine.AddTask(int(task.ID), task.Schedule, handler)
		if err != nil {
			task.Status = "disabled"
			s.db.Save(task)
			return nil, fmt.Errorf("register cron task: %w", err)
		}
		task.EntryID = entryID
		if next := s.engine.NextRunTime(int(task.ID)); next != nil {
			task.NextRunAt = next
		}
		s.db.Save(task)
	}

	return task, nil
}

// ListTasks 分页获取任务列表
func (s *Service) ListTasks(page, pageSize int) ([]model.CronTask, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	tasks := make([]model.CronTask, 0)
	var total int64

	if err := s.db.Model(&model.CronTask{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count cron tasks: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := s.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("query cron tasks: %w", err)
	}

	return tasks, total, nil
}

// GetTask 获取任务详情
func (s *Service) GetTask(id uint) (*model.CronTask, error) {
	var task model.CronTask
	if err := s.db.First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cron task not found")
		}
		return nil, fmt.Errorf("query cron task: %w", err)
	}
	return &task, nil
}

// UpdateTask 更新任务，处理 schedule 变更（重新注册）
func (s *Service) UpdateTask(id uint, req dto.UpdateCronRequest) (*model.CronTask, error) {
	task, err := s.GetTask(id)
	if err != nil {
		return nil, err
	}

	scheduleChanged := false

	if req.Name != nil {
		task.Name = *req.Name
	}
	if req.Type != nil {
		task.Type = *req.Type
	}
	if req.Schedule != nil {
		if *req.Schedule != task.Schedule {
			scheduleChanged = true
		}
		task.Schedule = *req.Schedule
	}
	if req.Command != nil {
		task.Command = *req.Command
	}
	if req.Target != nil {
		task.Target = *req.Target
	}
	if req.Remark != nil {
		task.Remark = *req.Remark
	}

	if err := s.db.Save(task).Error; err != nil {
		return nil, fmt.Errorf("update cron task: %w", err)
	}

	// 如果 schedule 变更且任务处于 enabled 状态，重新注册
	if scheduleChanged && task.Status == "enabled" && s.engine != nil {
		s.engine.RemoveTask(int(task.ID))
		handler := s.makeTaskHandler(task)
		entryID, err := s.engine.AddTask(int(task.ID), task.Schedule, handler)
		if err != nil {
			return nil, fmt.Errorf("re-register cron task: %w", err)
		}
		task.EntryID = entryID
		if next := s.engine.NextRunTime(int(task.ID)); next != nil {
			task.NextRunAt = next
		}
		s.db.Save(task)
	}

	return task, nil
}

// DeleteTask 从引擎移除并删除 DB 记录
func (s *Service) DeleteTask(id uint) error {
	task, err := s.GetTask(id)
	if err != nil {
		return err
	}

	if s.engine != nil {
		s.engine.RemoveTask(int(task.ID))
	}

	// 删除关联日志
	s.db.Where("task_id = ?", id).Delete(&model.CronLog{})

	if err := s.db.Delete(task).Error; err != nil {
		return fmt.Errorf("delete cron task: %w", err)
	}

	return nil
}

// EnableTask 启用任务
func (s *Service) EnableTask(id uint) (*model.CronTask, error) {
	task, err := s.GetTask(id)
	if err != nil {
		return nil, err
	}

	task.Status = "enabled"

	if s.engine != nil {
		handler := s.makeTaskHandler(task)
		entryID, err := s.engine.AddTask(int(task.ID), task.Schedule, handler)
		if err != nil {
			return nil, fmt.Errorf("register cron task: %w", err)
		}
		task.EntryID = entryID
		if next := s.engine.NextRunTime(int(task.ID)); next != nil {
			task.NextRunAt = next
		}
	}

	if err := s.db.Save(task).Error; err != nil {
		return nil, fmt.Errorf("enable cron task: %w", err)
	}

	return task, nil
}

// DisableTask 禁用任务
func (s *Service) DisableTask(id uint) (*model.CronTask, error) {
	task, err := s.GetTask(id)
	if err != nil {
		return nil, err
	}

	task.Status = "disabled"
	task.EntryID = 0
	task.NextRunAt = nil

	if s.engine != nil {
		s.engine.RemoveTask(int(task.ID))
	}

	if err := s.db.Save(task).Error; err != nil {
		return nil, fmt.Errorf("disable cron task: %w", err)
	}

	return task, nil
}

// TriggerTask 手动立即执行（异步）
func (s *Service) TriggerTask(id uint) error {
	task, err := s.GetTask(id)
	if err != nil {
		return err
	}

	go s.runTask(task)
	return nil
}

// ListLogs 分页获取执行日志
func (s *Service) ListLogs(taskID uint, page, pageSize int) ([]model.CronLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	query := s.db.Model(&model.CronLog{}).Where("task_id = ?", taskID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count cron logs: %w", err)
	}
	logs := make([]model.CronLog, 0)
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("query cron logs: %w", err)
	}

	return logs, total, nil
}

// SetEngine 设置 cron 引擎（用于外部初始化后注入）
func (s *Service) SetEngine(engine *cronengine.CronEngine) {
	s.engine = engine
}

// --- 响应转换 ---

// ToTaskResponse 将 model.CronTask 转换为 dto.CronTaskResponse
func ToTaskResponse(t *model.CronTask) dto.CronTaskResponse {
	resp := dto.CronTaskResponse{
		ID:        t.ID,
		Name:      t.Name,
		Type:      t.Type,
		Schedule:  t.Schedule,
		Command:   t.Command,
		Target:    t.Target,
		Remark:    t.Remark,
		Status:    t.Status,
		CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if t.LastRunAt != nil {
		resp.LastRunAt = t.LastRunAt.Format("2006-01-02 15:04:05")
	}
	if t.NextRunAt != nil {
		resp.NextRunAt = t.NextRunAt.Format("2006-01-02 15:04:05")
		resp.NextRunTime = t.NextRunAt.Format("2006-01-02 15:04:05")
	}

	return resp
}

// ToLogResponse 将 model.CronLog 转换为 dto.CronLogResponse
func ToLogResponse(log *model.CronLog, taskName string) dto.CronLogResponse {
	resp := dto.CronLogResponse{
		ID:        log.ID,
		TaskID:    log.TaskID,
		TaskName:  taskName,
		StartTime: log.StartTime.Format("2006-01-02 15:04:05"),
		Output:    log.Output,
		Result:    log.Result,
		CreatedAt: log.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if log.EndTime != nil {
		resp.EndTime = log.EndTime.Format("2006-01-02 15:04:05")
		duration := log.EndTime.Sub(log.StartTime)
		resp.Duration = duration.String()
	}

	return resp
}
