package model

import "time"

// RecoveryCode TOTP 恢复码
type RecoveryCode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	CodeHash  string    `gorm:"size:128;not null" json:"-"`
	Used      bool      `gorm:"default:false" json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

func (RecoveryCode) TableName() string { return "recovery_codes" }
