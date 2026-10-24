package backup

import (
	"os"

	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// ApplyEncryption 按设置对备份文件加密（roadmap #4）
// 设置 backup_encrypt=true 时：加密 FilePath 并更新记录为 .enc。
// 幂等：已加密文件直接跳过。
func ApplyEncryption(db *gorm.DB, backup *model.Backup) error {
	var setting model.Setting
	if err := db.Where("`key` = ?", "backup_encrypt").First(&setting).Error; err != nil {
		return nil // 未设置该键时默认不加密
	}
	if setting.Value != "true" && setting.Value != "1" {
		return nil
	}
	if backup.FilePath == "" {
		return nil
	}
	if IsEncryptedFile(backup.FilePath) {
		backup.Encrypted = true
		return nil
	}
	encPath := backup.FilePath + ".enc"
	if err := EncryptFile(backup.FilePath, encPath); err != nil {
		return err
	}
	backup.FilePath = encPath
	backup.Encrypted = true
	if info, err := os.Stat(encPath); err == nil {
		backup.Size = info.Size()
	}
	return nil
}

// DecryptIfNeeded 恢复前解密（非加密文件原样返回）
func DecryptIfNeeded(path string) (string, bool, error) {
	if !IsEncryptedFile(path) {
		return path, false, nil
	}
	plain := path + ".dec"
	if err := DecryptFile(path, plain); err != nil {
		return "", true, err
	}
	return plain, true, nil
}
