package backup

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// --- 备份加密（AES-256-GCM） ---
// 密钥文件 data/backup.key（32 字节 hex 编码，首次自动生成，0640 权限）。
// 加密后的备份文件后缀为 .enc，恢复时自动识别并解密。

const backupKeyPath = "data/backup.key"

// loadBackupKey 读取（或初始化）备份加密密钥
func loadBackupKey() ([]byte, error) {
	data, err := os.ReadFile(backupKeyPath)
	if err == nil && len(data) == 64 {
		raw, decErr := hex.DecodeString(string(data[:64]))
		if decErr == nil {
			return raw, nil
		}
	}
	// 生成新密钥
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate backup key: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(backupKeyPath), 0755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(backupKeyPath, []byte(hex.EncodeToString(key)), 0600); err != nil {
		return nil, fmt.Errorf("write backup key: %w", err)
	}
	return key, nil
}

// EncryptFile 用 AES-256-GCM 加密 src 到 dst（先写临时文件再原子替换，避免半截密文）
func EncryptFile(src, dst string) error {
	key, err := loadBackupKey()
	if err != nil {
		return err
	}
	plain, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read plaintext: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	// 头部: magic "RPENC1" + nonce，便于恢复时识别与提取
	header := append([]byte("RPENC1"), nonce...)
	out := append(header, gcm.Seal(nil, nonce, plain, nil)...)

	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, out, 0600); err != nil {
		return fmt.Errorf("write ciphertext: %w", err)
	}
	if err := os.Rename(tmp, dst); err != nil {
		return fmt.Errorf("rename ciphertext: %w", err)
	}
	_ = os.Remove(src) // 明文备份不再保留
	return nil
}

// DecryptFile 解密 src 到 dst
func DecryptFile(src, dst string) error {
	key, err := loadBackupKey()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read ciphertext: %w", err)
	}
	if len(data) < 6+12 {
		return fmt.Errorf("ciphertext too short")
	}
	if string(data[:6]) != "RPENC1" {
		return fmt.Errorf("not an encrypted backup")
	}
	nonce := data[6:18]
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	plain, err := gcm.Open(nil, nonce, data[18:], nil)
	if err != nil {
		return fmt.Errorf("decrypt failed (密钥不匹配或文件损坏): %w", err)
	}
	return os.WriteFile(dst, plain, 0600)
}

// KeyFingerprint 返回密钥指纹（SHA-256 前 8 字节），供前端展示确认用
func KeyFingerprint() string {
	key, err := loadBackupKey()
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(key)
	return hex.EncodeToString(sum[:8])
}

// IsEncryptedFile 判断是否为加密备份文件
func IsEncryptedFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	head := make([]byte, 6)
	if _, err := io.ReadFull(f, head); err != nil {
		return false
	}
	return string(head) == "RPENC1"
}
