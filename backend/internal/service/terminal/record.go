package terminal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// --- 会话录制（T-006）与录制回放 ---

const recordDir = "./data/terminal_records"

// EnableRecording 为会话开启录制，返回录制文件路径
func (s *TerminalSession) EnableRecording() (string, error) {
	if err := os.MkdirAll(recordDir, 0750); err != nil {
		return "", fmt.Errorf("create record dir: %w", err)
	}
	path := filepath.Join(recordDir, fmt.Sprintf("%s_%s.log", s.ID, time.Now().Format("20060102_150405")))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return "", fmt.Errorf("create record file: %w", err)
	}
	s.mu.Lock()
	s.RecordFile = f
	s.mu.Unlock()
	return path, nil
}

// StopRecording 停止录制并关闭文件
func (s *TerminalSession) StopRecording() {
	s.mu.Lock()
	f := s.RecordFile
	s.RecordFile = nil
	s.mu.Unlock()
	if f != nil {
		_ = f.Close()
	}
}

// RecordInfo 录制文件信息
type RecordInfo struct {
	Name string    `json:"name"`
	Size int64     `json:"size"`
	Time time.Time `json:"time"`
}

// ListRecords 列出全部终端录制文件
func ListRecords() ([]RecordInfo, error) {
	if err := os.MkdirAll(recordDir, 0750); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(recordDir)
	if err != nil {
		return []RecordInfo{}, nil
	}
	records := make([]RecordInfo, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".log") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		records = append(records, RecordInfo{
			Name: e.Name(),
			Size: info.Size(),
			Time: info.ModTime(),
		})
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Time.After(records[j].Time) })
	return records, nil
}

// ReadRecord 读取录制内容（回放）。文件名经清洗防穿越。
func ReadRecord(name string) (string, error) {
	name = filepath.Base(name)
	if !strings.HasSuffix(name, ".log") || strings.Contains(name, "..") {
		return "", fmt.Errorf("无效的录制文件名")
	}
	data, err := os.ReadFile(filepath.Join(recordDir, name))
	if err != nil {
		return "", fmt.Errorf("录制文件不存在")
	}
	return string(data), nil
}

// DeleteRecord 删除录制文件
func DeleteRecord(name string) error {
	name = filepath.Base(name)
	if !strings.HasSuffix(name, ".log") || strings.Contains(name, "..") {
		return fmt.Errorf("无效的录制文件名")
	}
	return os.Remove(filepath.Join(recordDir, name))
}

// --- 快捷命令面板（T-004） ---

// QuickCommand 快捷命令项
type QuickCommand struct {
	Label   string `json:"label"`
	Command string `json:"command"`
	Desc    string `json:"desc"`
}

// QuickCommands 内置常用命令
var QuickCommands = []QuickCommand{
	{Label: "系统概览", Command: "top -bn1 | head -20", Desc: "查看 CPU / 内存占用最高的进程"},
	{Label: "磁盘用量", Command: "df -h", Desc: "查看各挂载点磁盘使用情况"},
	{Label: "内存详情", Command: "free -h", Desc: "查看物理内存与交换分区"},
	{Label: "网络连接", Command: "ss -tunlp", Desc: "查看监听端口与连接"},
	{Label: "最近登录", Command: "last -20", Desc: "查看最近 20 条登录记录"},
	{Label: "系统日志", Command: "journalctl -n 50 --no-pager", Desc: "查看最近 50 条系统日志"},
	{Label: "Docker 概览", Command: "docker ps -a --format 'table {{.Names}}\\t{{.Status}}\\t{{.Ports}}'", Desc: "查看全部容器状态"},
	{Label: "Docker 资源", Command: "docker stats --no-stream", Desc: "查看容器实时资源占用"},
	{Label: "查找大文件", Command: "find / -type f -size +100M -exec ls -lh {} \\; 2>/dev/null | head -20", Desc: "查找超过 100MB 的大文件"},
	{Label: "清理页缓存", Command: "sync && echo 3 > /proc/sys/vm/drop_caches", Desc: "释放页缓存（需 root）"},
}
