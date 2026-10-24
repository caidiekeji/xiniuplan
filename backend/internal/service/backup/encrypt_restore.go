package backup

import (
	"os"

	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// ResolveBackupFile 恢复前解析备份文件：加密备份解密到临时文件，
// 返回 (可读路径, 清理函数, 错误)。非加密备份原样返回、清理为空操作。
func ResolveBackupFile(db *gorm.DB, backup *model.Backup) (string, func(), error) {
	if !IsEncryptedFile(backup.FilePath) {
		return backup.FilePath, func() {}, nil
	}
	plain, _, err := DecryptIfNeeded(backup.FilePath)
	if err != nil {
		return "", func() {}, err
	}
	return plain, func() { _ = os.Remove(plain) }, nil
}
