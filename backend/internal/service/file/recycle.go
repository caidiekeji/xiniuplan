package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// --- 文件回收站 ---
// Delete 时文件移入 <root>/.recycle/<ts>_<name> 而非直接删除，
// 7 天后由 PurgeExpired 自动清理。DeleteForever 保持硬删语义。

const recycleDirName = ".recycle"
const recycleRetentionDays = 7

func (m *Manager) recycleDir() string {
	return filepath.Join(m.Root(), recycleDirName)
}

// ToRecycle 批量移入回收站（纯文件系统；DB 记录由 API 层写入）。
// 回收站自身内部的文件直接硬删，避免递归。
func (m *Manager) ToRecycle(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no paths provided")
	}
	recycleDir := m.recycleDir()
	if err := os.MkdirAll(recycleDir, 0750); err != nil {
		return err
	}
	for _, p := range paths {
		abs, err := m.safePath(p)
		if err != nil {
			return err
		}
		if filepath.Dir(abs) == recycleDir || p == recycleDirName {
			if err := os.RemoveAll(abs); err != nil {
				return err
			}
			continue
		}
		stamp := time.Now().Format("20060102-150405")
		dst := filepath.Join(recycleDir, stamp+"_"+filepath.Base(abs))
		if err := os.Rename(abs, dst); err != nil {
			if cpErr := copyPath(abs, dst); cpErr != nil {
				return cpErr
			}
			if rmErr := os.RemoveAll(abs); rmErr != nil {
				return rmErr
			}
		}
	}
	return nil
}

// RecordRecycle 写入回收站 DB 记录（API 层调用）
func (m *Manager) RecordRecycle(paths []string, db *gorm.DB) {
	for _, p := range paths {
		abs, err := m.safePath(p)
		if err != nil {
			continue
		}
		info, err := os.Lstat(filepath.Join(m.recycleDir(), latestMatch(m.recycleDir(), filepath.Base(abs))))
		if err != nil || info == nil {
			continue
		}
		db.Create(&model.RecycleItem{
			OriginalPath: m.toVirtual(abs),
			RecyclePath:  filepath.Join(m.recycleDir(), latestMatch(m.recycleDir(), filepath.Base(abs))),
			Size:         info.Size(),
			IsDir:        info.IsDir(),
			DeletedAt:    time.Now(),
		})
	}
}

// latestMatch 在回收站内找指定基名最新的时间戳副本
func latestMatch(dir, base string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	best := ""
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, "_"+base) && name > best {
			best = name
		}
	}
	return best
}

// ListRecycle 回收站列表
func (m *Manager) ListRecycle(db *gorm.DB) ([]model.RecycleItem, error) {
	var items []model.RecycleItem
	err := db.Order("deleted_at DESC").Find(&items).Error
	return items, err
}

// RestoreFromRecycle 从回收站还原
func (m *Manager) RestoreFromRecycle(id uint, db *gorm.DB) error {
	var item model.RecycleItem
	if err := db.First(&item, id).Error; err != nil {
		return err
	}
	abs, err := m.safePath(item.OriginalPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		return err
	}
	if err := os.Rename(item.RecyclePath, abs); err != nil {
		if cpErr := copyPath(item.RecyclePath, abs); cpErr != nil {
			return cpErr
		}
		if rmErr := os.RemoveAll(item.RecyclePath); rmErr != nil {
			return rmErr
		}
	}
	return db.Delete(&model.RecycleItem{}, id).Error
}

// PurgeRecycle 彻底删除回收站项
func (m *Manager) PurgeRecycle(id uint, db *gorm.DB) error {
	var item model.RecycleItem
	if err := db.First(&item, id).Error; err != nil {
		return err
	}
	if err := os.RemoveAll(item.RecyclePath); err != nil {
		return err
	}
	return db.Delete(&model.RecycleItem{}, id).Error
}

// PurgeExpired 清理超过保留期的回收站项
func (m *Manager) PurgeExpired(db *gorm.DB) (int, error) {
	cutoff := time.Now().AddDate(0, 0, -recycleRetentionDays)
	var items []model.RecycleItem
	if err := db.Where("deleted_at < ?", cutoff).Find(&items).Error; err != nil {
		return 0, err
	}
	for _, item := range items {
		_ = os.RemoveAll(item.RecyclePath)
		db.Delete(&model.RecycleItem{}, item.ID)
	}
	return len(items), nil
}

func copyPath(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dst, info.Mode()); err != nil {
			return err
		}
		for _, e := range entries {
			if err := copyPath(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode())
}
