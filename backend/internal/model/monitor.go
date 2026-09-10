package model

// MonitorMetric 监控指标模型
type MonitorMetric struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	MetricType string  `gorm:"size:32;not null;index:idx_metric_ts" json:"metric_type"` // cpu | memory | disk | network
	Value      float64 `gorm:"not null" json:"value"`
	Timestamp  int64   `gorm:"not null;index:idx_metric_ts" json:"timestamp"` // unix seconds
}

func (MonitorMetric) TableName() string { return "monitor_metrics" }
