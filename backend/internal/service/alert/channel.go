package alert

import (
	"fmt"
	"time"

	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// ChannelService 告警渠道服务
type ChannelService struct {
	db       *gorm.DB
	notifier *NotifierManager
}

// NewChannelService 创建告警渠道服务
func NewChannelService(db *gorm.DB) *ChannelService {
	return &ChannelService{
		db:       db,
		notifier: NewNotifierManager(db),
	}
}

// ListChannels 分页查询渠道列表
func (s *ChannelService) ListChannels(page, pageSize int) ([]model.AlertChannel, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	channels := make([]model.AlertChannel, 0)
	var total int64
	if err := s.db.Model(&model.AlertChannel{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count channels: %w", err)
	}
	if err := s.db.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&channels).Error; err != nil {
		return nil, 0, fmt.Errorf("query channels: %w", err)
	}
	return channels, total, nil
}

// GetChannel 获取单条渠道
func (s *ChannelService) GetChannel(id uint) (*model.AlertChannel, error) {
	var channel model.AlertChannel
	if err := s.db.First(&channel, id).Error; err != nil {
		return nil, fmt.Errorf("channel not found: %w", err)
	}
	return &channel, nil
}

// CreateChannel 创建渠道
func (s *ChannelService) CreateChannel(req dto.CreateAlertChannelRequest) (*model.AlertChannel, error) {
	now := time.Now()
	channel := model.AlertChannel{
		Name:      req.Name,
		Type:      req.Type,
		Config:    model.JSONMap(req.Config),
		Enabled:   req.Enabled,
		CreatedAt: now,
	}
	if err := s.db.Create(&channel).Error; err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}
	return &channel, nil
}

// UpdateChannel 更新渠道
func (s *ChannelService) UpdateChannel(id uint, req dto.UpdateAlertChannelRequest) (*model.AlertChannel, error) {
	channel, err := s.GetChannel(id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.Type != "" {
		channel.Type = req.Type
	}
	if req.Config != nil {
		channel.Config = model.JSONMap(req.Config)
	}
	if req.Enabled != nil {
		channel.Enabled = *req.Enabled
	}
	if err := s.db.Save(channel).Error; err != nil {
		return nil, fmt.Errorf("update channel: %w", err)
	}
	return channel, nil
}

// DeleteChannel 删除渠道
func (s *ChannelService) DeleteChannel(id uint) error {
	if err := s.db.Delete(&model.AlertChannel{}, id).Error; err != nil {
		return fmt.Errorf("delete channel: %w", err)
	}
	return nil
}

// TestChannel 测试渠道（发送测试消息）
func (s *ChannelService) TestChannel(id uint, req dto.TestChannelRequest) error {
	channel, err := s.GetChannel(id)
	if err != nil {
		return err
	}
	title := req.Title
	if title == "" {
		title = "告警渠道测试"
	}
	content := req.Content
	if content == "" {
		content = fmt.Sprintf("这是一条来自犀牛面板的测试消息，渠道类型: %s，时间: %s",
			channel.Type, time.Now().Format("2006-01-02 15:04:05"))
	}
	return s.notifier.Send(*channel, title, content)
}

// ToChannelResponse 将模型转为响应 DTO
func ToChannelResponse(channel *model.AlertChannel) dto.AlertChannelResponse {
	return dto.AlertChannelResponse{
		ID:        channel.ID,
		Name:      channel.Name,
		Type:      channel.Type,
		Config:    map[string]interface{}(channel.Config),
		Enabled:   channel.Enabled,
		CreatedAt: channel.CreatedAt.Format(time.RFC3339),
	}
}

// ToHistoryResponse 将模型转为响应 DTO
func ToHistoryResponse(h *model.AlertHistory) dto.AlertHistoryResponse {
	recoveredAt := ""
	if h.RecoveredAt != nil {
		recoveredAt = h.RecoveredAt.Format(time.RFC3339)
	}
	return dto.AlertHistoryResponse{
		ID:                 h.ID,
		RuleID:             h.RuleID,
		RuleName:           h.RuleName,
		MetricType:         h.MetricType,
		TriggerValue:       h.TriggerValue,
		Threshold:          h.Threshold,
		Status:             h.Status,
		TriggeredAt:        h.TriggeredAt.Format(time.RFC3339),
		RecoveredAt:        recoveredAt,
		NotificationStatus: h.NotificationStatus,
		NotificationError:  h.NotificationError,
		CreatedAt:          h.CreatedAt.Format(time.RFC3339),
	}
}
