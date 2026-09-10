package audit

import (
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// Service 审计日志服务
type Service struct {
	db *gorm.DB
}

// NewService 创建审计日志服务
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ListAuditLogs 分页查询操作日志，支持筛选
func (s *Service) ListAuditLogs(page, pageSize int, userID uint, action, result string) ([]model.AuditLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	query := s.db.Model(&model.AuditLog{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if result != "" {
		query = query.Where("result = ?", result)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}

	logs := make([]model.AuditLog, 0)
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("query audit logs: %w", err)
	}

	return logs, total, nil
}

// ListLoginLogs 分页查询登录日志（action=login）
func (s *Service) ListLoginLogs(page, pageSize int) ([]model.AuditLog, int64, error) {
	return s.ListAuditLogs(page, pageSize, 0, "login", "")
}

// ExportAuditLogs 导出指定时间范围的审计日志为 CSV 格式字符串
func (s *Service) ExportAuditLogs(startDate, endDate string) (string, error) {
	query := s.db.Model(&model.AuditLog{})

	if startDate != "" {
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return "", fmt.Errorf("invalid start date: %w", err)
		}
		query = query.Where("created_at >= ?", start)
	}
	if endDate != "" {
		end, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return "", fmt.Errorf("invalid end date: %w", err)
		}
		// 包含结束日期全天
		query = query.Where("created_at < ?", end.AddDate(0, 0, 1))
	}

	var logs []model.AuditLog
	if err := query.Order("created_at ASC").Find(&logs).Error; err != nil {
		return "", fmt.Errorf("query audit logs for export: %w", err)
	}

	// 构建 CSV
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// 写入表头
	header := []string{"ID", "用户ID", "操作", "资源", "IP", "UserAgent", "结果", "创建时间"}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("write csv header: %w", err)
	}

	// 写入数据行
	for _, log := range logs {
		row := []string{
			fmt.Sprintf("%d", log.ID),
			fmt.Sprintf("%d", log.UserID),
			log.Action,
			log.Resource,
			log.IP,
			log.UserAgent,
			log.Result,
			log.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("write csv row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("flush csv: %w", err)
	}

	return buf.String(), nil
}

// ToResponse 转换模型为审计日志响应 DTO
func ToResponse(log *model.AuditLog) dto.AuditLogItem {
	return dto.AuditLogItem{
		ID:        log.ID,
		UserID:    log.UserID,
		Action:    log.Action,
		Resource:  log.Resource,
		IP:        log.IP,
		UserAgent: log.UserAgent,
		Result:    log.Result,
		CreatedAt: log.CreatedAt.Format(time.RFC3339),
	}
}
