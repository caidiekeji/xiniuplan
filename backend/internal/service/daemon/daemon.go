package daemon

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// Service 进程守护（roadmap #11）
// 托管用户进程：崩溃自动拉起、面板启动时恢复、日志落盘
type Service struct {
	db      *gorm.DB
	runtime string // 守护进程日志根目录
	mu      sync.Mutex
	stopCh  chan struct{}
	running bool
	cmds    map[uint]*exec.Cmd // 运行中的进程句柄
}

// NewService 创建守护服务。runtimeDir 为日志目录（默认 ./runtime/daemon）
func NewService(db *gorm.DB, runtimeDir string) *Service {
	if runtimeDir == "" {
		runtimeDir = "./runtime/daemon"
	}
	return &Service{db: db, runtime: runtimeDir, cmds: map[uint]*exec.Cmd{}}
}

// StartService 启动监控循环与自恢复
func (s *Service) StartService() {
	if s.running {
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	_ = os.MkdirAll(s.runtime, 0755)
	// 恢复 enabled + auto_start 的进程
	s.RestoreAll()
	go s.watchLoop()
	slog.Info("[daemon] 进程守护服务已启动")
}

// Stop 停止监控（不停止被守护的进程，它们独立运行）
func (s *Service) Stop() {
	if s.running {
		close(s.stopCh)
		s.running = false
	}
}

func (s *Service) watchLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.checkAll()
		case <-s.stopCh:
			return
		}
	}
}

// checkAll 检查所有 enabled 且标记运行中的进程，死亡则自动拉起
func (s *Service) checkAll() {
	var list []model.DaemonProcess
	if err := s.db.Where("enabled = ?", true).Find(&list).Error; err != nil {
		return
	}
	for i := range list {
		p := &list[i]
		if p.PID <= 0 || !s.pidAlive(p.PID) {
			s.mu.Lock()
			_, inFlight := s.cmds[p.ID]
			s.mu.Unlock()
			if !inFlight {
				slog.Info("[daemon] 进程死亡，自动拉起", "name", p.Name, "old_pid", p.PID)
				_, _ = s.Launch(p.ID)
			}
		}
	}
}

// RestoreAll 面板启动时恢复 auto_start 的进程
func (s *Service) RestoreAll() {
	var list []model.DaemonProcess
	if err := s.db.Where("enabled = ? AND auto_start = ?", true, true).Find(&list).Error; err != nil {
		return
	}
	for i := range list {
		if list[i].PID <= 0 || !s.pidAlive(list[i].PID) {
			if _, err := s.Launch(list[i].ID); err != nil {
				slog.Error("[daemon] 恢复失败", "name", list[i].Name, "error", err)
			}
		}
	}
}

// pidAlive 检查进程是否存活
func (s *Service) pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// unix: signal 0 探活
	return p.Signal(pidSignal()) == nil
}

// ===== CRUD 与控制 =====

// List 进程列表（附带存活状态）
func (s *Service) List() ([]model.DaemonProcess, error) {
	list := make([]model.DaemonProcess, 0)
	if err := s.db.Order("created_at ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range list {
		if list[i].PID > 0 && s.pidAlive(list[i].PID) {
			list[i].LastExitCode = 0 // 运行中
		} else if list[i].PID > 0 {
			list[i].PID = 0 // 已死亡，清零标记
			s.db.Model(&model.DaemonProcess{}).Where("id = ?", list[i].ID).Update("pid", 0)
		}
	}
	return list, nil
}

// Create 创建守护任务并立即启动
func (s *Service) Create(p *model.DaemonProcess) error {
	if p.Name == "" || p.Command == "" {
		return fmt.Errorf("名称和命令不能为空")
	}
	if err := s.db.Create(p).Error; err != nil {
		return err
	}
	if p.Enabled {
		_, _ = s.Launch(p.ID)
	}
	return nil
}

// Update 更新配置（不中断运行中的进程，下次重启生效）
func (s *Service) Update(id uint, req *model.DaemonProcess) error {
	return s.db.Model(&model.DaemonProcess{}).Where("id = ?", id).Updates(map[string]any{
		"name": req.Name, "command": req.Command, "work_dir": req.WorkDir,
		"enabled": req.Enabled, "auto_start": req.AutoStart,
	}).Error
}

// Delete 删除守护任务（先停止进程）
func (s *Service) Delete(id uint) error {
	var p model.DaemonProcess
	if err := s.db.First(&p, id).Error; err != nil {
		return err
	}
	_ = s.StopProcess(id)
	s.mu.Lock()
	delete(s.cmds, id)
	s.mu.Unlock()
	return s.db.Delete(&model.DaemonProcess{}, id).Error
}

// Launch 启动（或拉起）守护进程
func (s *Service) Launch(id uint) (*model.DaemonProcess, error) {
	var p model.DaemonProcess
	if err := s.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	if p.PID > 0 && s.pidAlive(p.PID) {
		return &p, nil // 已在运行
	}
	if !p.Enabled {
		return &p, fmt.Errorf("进程 %s 未启用", p.Name)
	}

	logFile := filepath.Join(s.runtime, p.Name+".log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}
	defer f.Close()

	cmd := exec.Command("/bin/sh", "-c", p.Command)
	if p.WorkDir != "" {
		cmd.Dir = p.WorkDir
	}
	cmd.Stdout = bufio.NewWriter(f)
	cmd.Stderr = bufio.NewWriter(f)
	setDetach(cmd)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动失败: %w", err)
	}

	s.mu.Lock()
	s.cmds[p.ID] = cmd
	s.mu.Unlock()

	// 异步等待退出：记录退出码，触发 watchLoop 自动拉起
	go func(pid int) {
		_ = cmd.Wait()
		s.mu.Lock()
		delete(s.cmds, p.ID)
		s.mu.Unlock()
		s.db.Model(&model.DaemonProcess{}).Where("id = ?", p.ID).Updates(map[string]any{
			"pid": 0, "last_exit_code": cmd.ProcessState.ExitCode(),
		})
	}(cmd.Process.Pid)

	now := time.Now()
	updates := map[string]any{
		"pid": cmd.Process.Pid, "last_started": now,
		"restarts": p.Restarts + 1, "last_exit_code": 0,
	}
	if err := s.db.Model(&p).Updates(updates).Error; err != nil {
		return nil, err
	}
	p.PID = cmd.Process.Pid
	p.LastStarted = &now
	slog.Info("[daemon] 进程已启动", "name", p.Name, "pid", cmd.Process.Pid)
	return &p, nil
}

// StopProcess 停止守护进程（禁用自动拉起由调用方先置 enabled=false）
func (s *Service) StopProcess(id uint) error {
	var p model.DaemonProcess
	if err := s.db.First(&p, id).Error; err != nil {
		return err
	}
	if p.PID > 0 {
		// 先杀进程组（Setsid 后子进程同组），再兜底杀主进程
		exec.Command("kill", "-TERM", "-"+strconv.Itoa(p.PID)).Run()
		time.Sleep(500 * time.Millisecond)
		if s.pidAlive(p.PID) {
			exec.Command("kill", "-KILL", "-"+strconv.Itoa(p.PID)).Run()
		}
	}
	s.mu.Lock()
	delete(s.cmds, id)
	s.mu.Unlock()
	return s.db.Model(&p).Update("pid", 0).Error
}

// Logs 读取守护进程日志尾部
func (s *Service) Logs(id uint, lines int) (string, error) {
	var p model.DaemonProcess
	if err := s.db.First(&p, id).Error; err != nil {
		return "", err
	}
	if lines <= 0 || lines > 2000 {
		lines = 200
	}
	logFile := filepath.Join(s.runtime, p.Name+".log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		return "", nil // 无日志视为空
	}
	all := string(data)
	linesArr := splitLines(all)
	if len(linesArr) > lines {
		linesArr = linesArr[len(linesArr)-lines:]
	}
	out := ""
	for _, l := range linesArr {
		out += l + "\n"
	}
	return out, nil
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
