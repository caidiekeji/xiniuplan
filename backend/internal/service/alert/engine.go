package alert

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

const (
	checkInterval   = 60 * time.Second // 检查间隔
	debounceMinutes = 5                // 同一规则防抖分钟数
)

// AlertEngine 告警引擎，后台定时检查规则并触发/恢复告警
type AlertEngine struct {
	db       *gorm.DB
	notifier *NotifierManager
	stopCh   chan struct{}
	mu       sync.Mutex
	running  bool
}

// NewAlertEngine 创建告警引擎
func NewAlertEngine(db *gorm.DB) *AlertEngine {
	return &AlertEngine{
		db:       db,
		notifier: NewNotifierManager(db),
		stopCh:   make(chan struct{}),
	}
}

// InitAlertEngine 导出初始化函数，供 main.go 调用
func InitAlertEngine(db *gorm.DB) *AlertEngine {
	return NewAlertEngine(db)
}

// Start 启动后台检查 goroutine
func (e *AlertEngine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.running {
		return
	}
	e.running = true
	e.stopCh = make(chan struct{})
	go e.run()
	slog.Info("[alert] 告警引擎已启动", "interval", checkInterval)
}

// Stop 停止后台检查 goroutine
func (e *AlertEngine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.running {
		return
	}
	e.running = false
	close(e.stopCh)
	slog.Info("[alert] 告警引擎已停止")
}

func (e *AlertEngine) run() {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	// 启动后立即检查一次
	e.checkAll()

	for {
		select {
		case <-ticker.C:
			e.checkAll()
		case <-e.stopCh:
			return
		}
	}
}

// checkAll 检查所有启用的规则
func (e *AlertEngine) checkAll() {
	var rules []model.AlertRule
	if err := e.db.Where("enabled = ?", true).Find(&rules).Error; err != nil {
		slog.Error("[alert] 查询规则失败", "error", err)
		return
	}
	for i := range rules {
		e.checkRule(&rules[i])
	}
}

// CheckRule 手动检查单条规则（供 API 测试调用）
func (e *AlertEngine) CheckRule(ruleID uint) error {
	var rule model.AlertRule
	if err := e.db.First(&rule, ruleID).Error; err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}
	e.checkRule(&rule)
	return nil
}

// checkRule 检查单条规则，处理触发与恢复
func (e *AlertEngine) checkRule(rule *model.AlertRule) {
	// 查询最近 duration 分钟的指标数据
	metrics, latestValue, err := e.queryMetrics(rule)
	if err != nil {
		slog.Error("[alert] 规则查询指标失败", "rule", rule.Name, "error", err)
		return
	}
	if len(metrics) == 0 {
		// 无数据不触发
		return
	}

	// 判断是否持续满足条件
	conditionMet := e.evaluateCondition(rule, metrics)

	// 查询该规则当前是否有未恢复的告警
	var activeHistory model.AlertHistory
	hasActive := false
	if err := e.db.Where("rule_id = ? AND status = ?", rule.ID, "triggered").
		Order("triggered_at DESC").First(&activeHistory).Error; err == nil {
		hasActive = true
	}

	if conditionMet && !hasActive {
		// 触发新告警（带防抖）
		if e.isDebounced(rule.ID) {
			return
		}
		e.triggerAlert(rule, latestValue)
	} else if !conditionMet && hasActive {
		// 恢复告警
		e.recoverAlert(rule, &activeHistory)
	}
}

// queryMetrics 查询规则对应指标在 duration 窗口内的数据
func (e *AlertEngine) queryMetrics(rule *model.AlertRule) ([]model.MonitorMetric, float64, error) {
	durationMin := rule.Duration
	if durationMin <= 0 {
		durationMin = 1
	}
	cutoff := time.Now().Add(-time.Duration(durationMin) * time.Minute).Unix()

	// network 类型需要查 sent/recv 两个子类型
	metricTypes := []string{rule.MetricType}
	if rule.MetricType == "network" {
		metricTypes = []string{"network_sent", "network_recv"}
	}

	metrics := make([]model.MonitorMetric, 0)
	if err := e.db.Where("metric_type IN ? AND timestamp >= ?", metricTypes, cutoff).
		Order("timestamp ASC").Find(&metrics).Error; err != nil {
		return nil, 0, fmt.Errorf("query metrics: %w", err)
	}

	// 取最新值
	var latestValue float64
	if len(metrics) > 0 {
		latestValue = metrics[len(metrics)-1].Value
	}
	return metrics, latestValue, nil
}

// evaluateCondition 判断窗口内所有数据点是否持续满足条件
func (e *AlertEngine) evaluateCondition(rule *model.AlertRule, metrics []model.MonitorMetric) bool {
	if len(metrics) == 0 {
		return false
	}
	for _, m := range metrics {
		if !matchCondition(rule.Condition, m.Value, rule.Threshold) {
			return false
		}
	}
	return true
}

// matchCondition 判断单个值是否满足条件
func matchCondition(condition string, value, threshold float64) bool {
	switch condition {
	case "gt":
		return value > threshold
	case "lt":
		return value < threshold
	case "eq":
		return value == threshold
	default:
		return false
	}
}

// isDebounced 检查规则是否在防抖期内（5分钟内已触发过则跳过）
func (e *AlertEngine) isDebounced(ruleID uint) bool {
	cutoff := time.Now().Add(-debounceMinutes * time.Minute)
	var count int64
	e.db.Model(&model.AlertHistory{}).
		Where("rule_id = ? AND triggered_at >= ?", ruleID, cutoff).
		Count(&count)
	return count > 0
}

// triggerAlert 触发告警，写入历史并发送通知
func (e *AlertEngine) triggerAlert(rule *model.AlertRule, triggerValue float64) {
	now := time.Now()
	history := model.AlertHistory{
		RuleID:             rule.ID,
		RuleName:           rule.Name,
		MetricType:         rule.MetricType,
		TriggerValue:       triggerValue,
		Threshold:          rule.Threshold,
		Status:             "triggered",
		TriggeredAt:        now,
		NotificationStatus: "pending",
		CreatedAt:          now,
	}
	if err := e.db.Create(&history).Error; err != nil {
		slog.Error("[alert] 写入告警历史失败", "error", err)
		return
	}

	// 发送通知
	title := fmt.Sprintf("[告警触发] %s", rule.Name)
	content := fmt.Sprintf("指标: %s\n条件: %s %.2f\n当前值: %.2f\n持续时间: %d 分钟\n触发时间: %s",
		rule.MetricType, rule.Condition, rule.Threshold, triggerValue, rule.Duration, now.Format("2006-01-02 15:04:05"))

	notifyErr := e.notifier.SendByIDs(rule.Channels, title, content)
	e.updateNotificationStatus(history.ID, notifyErr)

	slog.Info("[alert] 告警触发", "rule", rule.Name, "value", triggerValue, "threshold", rule.Threshold)
}

// recoverAlert 恢复告警，更新历史并发送恢复通知
func (e *AlertEngine) recoverAlert(rule *model.AlertRule, history *model.AlertHistory) {
	now := time.Now()
	history.Status = "recovered"
	history.RecoveredAt = &now
	if err := e.db.Save(history).Error; err != nil {
		slog.Error("[alert] 更新告警恢复状态失败", "error", err)
		return
	}

	// 发送恢复通知
	title := fmt.Sprintf("[告警恢复] %s", rule.Name)
	content := fmt.Sprintf("指标: %s\n阈值: %.2f\n触发值: %.2f\n触发时间: %s\n恢复时间: %s",
		rule.MetricType, rule.Threshold, history.TriggerValue,
		history.TriggeredAt.Format("2006-01-02 15:04:05"),
		now.Format("2006-01-02 15:04:05"))

	notifyErr := e.notifier.SendByIDs(rule.Channels, title, content)
	if notifyErr != nil {
		slog.Error("[alert] 恢复通知发送失败", "error", notifyErr)
	}

	slog.Info("[alert] 告警恢复", "rule", rule.Name)
}

// updateNotificationStatus 更新告警历史的通知状态
func (e *AlertEngine) updateNotificationStatus(historyID uint, notifyErr error) {
	updates := map[string]interface{}{}
	if notifyErr != nil {
		updates["notification_status"] = "failed"
		updates["notification_error"] = notifyErr.Error()
	} else {
		updates["notification_status"] = "sent"
		updates["notification_error"] = ""
	}
	e.db.Model(&model.AlertHistory{}).Where("id = ?", historyID).Updates(updates)
}
