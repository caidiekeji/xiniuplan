package cron

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// CronEngine 封装 robfig/cron/v3 的定时任务引擎
type CronEngine struct {
	cron    *cron.Cron
	mu      sync.RWMutex
	entries map[int]cron.EntryID // task db id -> cron entry id
}

// NewEngine 创建定时任务引擎
func NewEngine() *CronEngine {
	c := cron.New(cron.WithParser(cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))
	return &CronEngine{
		cron:    c,
		entries: make(map[int]cron.EntryID),
	}
}

// Start 启动引擎
func (e *CronEngine) Start() {
	e.cron.Start()
}

// Stop 停止引擎，等待正在执行的任务完成
func (e *CronEngine) Stop() {
	ctx := e.cron.Stop()
	<-ctx.Done()
}

// AddTask 注册任务到引擎，返回 cron entry id
func (e *CronEngine) AddTask(taskID int, schedule string, handler func()) (int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 如果已存在，先移除
	if entryID, ok := e.entries[taskID]; ok {
		e.cron.Remove(entryID)
		delete(e.entries, taskID)
	}

	entryID, err := e.cron.AddFunc(schedule, handler)
	if err != nil {
		return 0, fmt.Errorf("add cron task: %w", err)
	}

	e.entries[taskID] = entryID
	return int(entryID), nil
}

// RemoveTask 从引擎移除任务
func (e *CronEngine) RemoveTask(taskID int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if entryID, ok := e.entries[taskID]; ok {
		e.cron.Remove(entryID)
		delete(e.entries, taskID)
	}
}

// NextRunTime 获取任务下次执行时间
func (e *CronEngine) NextRunTime(taskID int) *time.Time {
	e.mu.RLock()
	defer e.mu.RUnlock()

	entryID, ok := e.entries[taskID]
	if !ok {
		return nil
	}

	entry := e.cron.Entry(entryID)
	if entry.Schedule == nil {
		return nil
	}
	return &entry.Next
}

// HasTask 检查任务是否已注册
func (e *CronEngine) HasTask(taskID int) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, ok := e.entries[taskID]
	return ok
}
