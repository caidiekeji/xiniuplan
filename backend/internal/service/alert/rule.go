package alert

import (
	"fmt"
	"time"

	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// RuleService 告警规则服务
type RuleService struct {
	db     *gorm.DB
	engine *AlertEngine
}

// NewRuleService 创建告警规则服务
func NewRuleService(db *gorm.DB, engine *AlertEngine) *RuleService {
	return &RuleService{db: db, engine: engine}
}

// ListRules 分页查询规则列表
func (s *RuleService) ListRules(page, pageSize int) ([]model.AlertRule, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	rules := make([]model.AlertRule, 0)
	var total int64
	if err := s.db.Model(&model.AlertRule{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count rules: %w", err)
	}
	if err := s.db.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&rules).Error; err != nil {
		return nil, 0, fmt.Errorf("query rules: %w", err)
	}
	return rules, total, nil
}

// GetRule 获取单条规则
func (s *RuleService) GetRule(id uint) (*model.AlertRule, error) {
	var rule model.AlertRule
	if err := s.db.First(&rule, id).Error; err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
	}
	return &rule, nil
}

// CreateRule 创建规则
func (s *RuleService) CreateRule(req dto.CreateAlertRuleRequest) (*model.AlertRule, error) {
	now := time.Now()
	rule := model.AlertRule{
		Name:       req.Name,
		MetricType: req.MetricType,
		Condition:  req.Condition,
		Threshold:  req.Threshold,
		Duration:   req.Duration,
		Enabled:    req.Enabled,
		Channels:   model.StringSlice(req.Channels),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if rule.Duration <= 0 {
		rule.Duration = 1
	}
	if err := s.db.Create(&rule).Error; err != nil {
		return nil, fmt.Errorf("create rule: %w", err)
	}
	return &rule, nil
}

// UpdateRule 更新规则
func (s *RuleService) UpdateRule(id uint, req dto.UpdateAlertRuleRequest) (*model.AlertRule, error) {
	rule, err := s.GetRule(id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		rule.Name = req.Name
	}
	if req.MetricType != "" {
		rule.MetricType = req.MetricType
	}
	if req.Condition != "" {
		rule.Condition = req.Condition
	}
	if req.Threshold != 0 {
		rule.Threshold = req.Threshold
	}
	if req.Duration > 0 {
		rule.Duration = req.Duration
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.Channels != nil {
		rule.Channels = model.StringSlice(req.Channels)
	}
	rule.UpdatedAt = time.Now()
	if err := s.db.Save(rule).Error; err != nil {
		return nil, fmt.Errorf("update rule: %w", err)
	}
	return rule, nil
}

// DeleteRule 删除规则
func (s *RuleService) DeleteRule(id uint) error {
	if err := s.db.Delete(&model.AlertRule{}, id).Error; err != nil {
		return fmt.Errorf("delete rule: %w", err)
	}
	return nil
}

// EnableRule 启用规则
func (s *RuleService) EnableRule(id uint) error {
	result := s.db.Model(&model.AlertRule{}).Where("id = ?", id).
		Updates(map[string]interface{}{"enabled": true, "updated_at": time.Now()})
	if result.Error != nil {
		return fmt.Errorf("enable rule: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("rule not found")
	}
	return nil
}

// DisableRule 禁用规则
func (s *RuleService) DisableRule(id uint) error {
	result := s.db.Model(&model.AlertRule{}).Where("id = ?", id).
		Updates(map[string]interface{}{"enabled": false, "updated_at": time.Now()})
	if result.Error != nil {
		return fmt.Errorf("disable rule: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("rule not found")
	}
	return nil
}

// TestRule 手动测试规则（立即检查一次）
func (s *RuleService) TestRule(id uint) error {
	if s.engine == nil {
		return fmt.Errorf("alert engine not initialized")
	}
	return s.engine.CheckRule(id)
}

// ToRuleResponse 将模型转为响应 DTO
func ToRuleResponse(rule *model.AlertRule) dto.AlertRuleResponse {
	return dto.AlertRuleResponse{
		ID:         rule.ID,
		Name:       rule.Name,
		MetricType: rule.MetricType,
		Condition:  rule.Condition,
		Threshold:  rule.Threshold,
		Duration:   rule.Duration,
		Enabled:    rule.Enabled,
		Channels:   []string(rule.Channels),
		CreatedAt:  rule.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  rule.UpdatedAt.Format(time.RFC3339),
	}
}
