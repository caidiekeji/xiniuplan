package site

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cloudpivot/internal/model"
)

// --- 网站防篡改 ---
// 基线：对站点根目录内的可执行/模板文件（.php/.html/.htm/.js）计算 SHA-256，
// 排除 vendor/node_modules/.git 等依赖目录；校验时重算比对，差异即告警。
// 目录兜底：无法定位站点根目录时以 NginxConf 同级的 public/ www/ 为候选。

var tamperSkipDirs = map[string]bool{
	"vendor": true, "node_modules": true, ".git": true,
	".recycle": true, "cache": true, "runtime": true, "temp": true,
}

var tamperExts = map[string]bool{
	".php": true, ".html": true, ".htm": true, ".js": true, ".json": true,
}

// siteRoot 返回站点根目录（优先站点自身路径，回退配置目录的 public/）
func (s *Service) siteRoot(site *model.Site) string {
	candidates := []string{}
	if site.RootDir != "" {
		candidates = append(candidates, site.RootDir)
	}
	if site.NginxConf != "" {
		base := filepath.Dir(site.NginxConf)
		candidates = append(candidates, filepath.Join(base, "public"), filepath.Join(base, "www"))
	}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return ""
}

// BuildTamperBaseline 重建站点防篡改基线，返回纳入保护的文件数
func (s *Service) BuildTamperBaseline(siteID uint) (int, error) {
	var site model.Site
	if err := s.db.First(&site, siteID).Error; err != nil {
		return 0, fmt.Errorf("站点不存在: %w", err)
	}
	root := s.siteRoot(&site)
	if root == "" {
		return 0, fmt.Errorf("未找到站点根目录")
	}
	// 清旧基线再重建
	s.db.Where("site_id = ?", siteID).Delete(&model.SiteTamperBaseline{})

	count := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过不可读文件
		}
		if info.IsDir() {
			if tamperSkipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !tamperExts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		if info.Size() > 8*1024*1024 { // 超过 8MB 跳过（如打包产物）
			return nil
		}
		hash, hErr := fileHash(path)
		if hErr != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		s.db.Create(&model.SiteTamperBaseline{
			SiteID: siteID,
			Path:   rel,
			Hash:   hash,
		})
		count++
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("扫描站点目录: %w", err)
	}
	return count, nil
}

// TamperDiff 单文件差异
type TamperDiff struct {
	Path    string `json:"path"`
	Status  string `json:"status"` // modified | added | deleted
	OldHash string `json:"old_hash,omitempty"`
	NewHash string `json:"new_hash,omitempty"`
}

// VerifyTamper 校验站点文件是否被篡改，返回差异列表
func (s *Service) VerifyTamper(siteID uint) ([]TamperDiff, error) {
	var site model.Site
	if err := s.db.First(&site, siteID).Error; err != nil {
		return nil, fmt.Errorf("站点不存在: %w", err)
	}
	var baseline []model.SiteTamperBaseline
	s.db.Where("site_id = ?", siteID).Find(&baseline)
	if len(baseline) == 0 {
		return nil, fmt.Errorf("尚未建立基线，请先执行「建立基线」")
	}
	base := map[string]string{}
	for _, b := range baseline {
		base[b.Path] = b.Hash
	}
	root := s.siteRoot(&site)
	if root == "" {
		return nil, fmt.Errorf("未找到站点根目录")
	}

	var diffs []TamperDiff
	// 1) 基线文件：缺失或哈希不符
	for _, b := range baseline {
		full := filepath.Join(root, b.Path)
		hash, err := fileHash(full)
		if err != nil {
			diffs = append(diffs, TamperDiff{Path: b.Path, Status: "deleted", OldHash: b.Hash})
			continue
		}
		if hash != b.Hash {
			diffs = append(diffs, TamperDiff{Path: b.Path, Status: "modified", OldHash: b.Hash, NewHash: hash})
		}
	}
	// 2) 新增的受保护类型文件
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() && tamperSkipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !tamperExts[strings.ToLower(filepath.Ext(path))] || info.Size() > 8*1024*1024 {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if _, ok := base[rel]; !ok {
			diffs = append(diffs, TamperDiff{Path: rel, Status: "added"})
		}
		return nil
	})
	return diffs, nil
}

func fileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
