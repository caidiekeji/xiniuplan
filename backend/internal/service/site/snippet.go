package site

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"cloudpivot/internal/model"
	"cloudpivot/pkg/nginx"

	"golang.org/x/crypto/bcrypt"
)

var snippetNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

// snippetDirOf 站点片段目录
func snippetDirOf(domain string) string {
	return confDirBase() + "/" + domain + ".d"
}

// ListSnippets 列出站点全部片段（W-006~009）
func (s *Service) ListSnippets(siteID uint) ([]map[string]string, error) {
	site, err := s.GetSite(siteID)
	if err != nil {
		return nil, err
	}
	dir := snippetDirOf(site.PrimaryDomain)
	os.MkdirAll(dir, 0755)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return []map[string]string{}, nil
	}
	items := make([]map[string]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".conf") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		items = append(items, map[string]string{
			"name":    strings.TrimSuffix(e.Name(), ".conf"),
			"content": string(content),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i]["name"] < items[j]["name"] })
	return items, nil
}

// SetSnippet 写入片段并重新生成主配置 + reload（片段靠 include 生效）
func (s *Service) SetSnippet(siteID uint, name, content string) error {
	site, err := s.GetSite(siteID)
	if err != nil {
		return err
	}
	if !snippetNameRegex.MatchString(name) {
		return fmt.Errorf("片段名只能包含字母数字下划线和连字符")
	}
	// 安全校验：片段内容会写入 conf.d 并被 nginx reload 加载
	if err := ValidateNginxContent(content); err != nil {
		return fmt.Errorf("片段内容校验失败: %w", err)
	}
	dir := snippetDirOf(site.PrimaryDomain)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create snippet dir: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".conf"), []byte(content), 0644); err != nil {
		return fmt.Errorf("write snippet: %w", err)
	}

	// 重新生成主配置（确保 include 指令存在）
	confContent, err := nginxGenerate(site)
	if err != nil {
		return err
	}
	if err := os.WriteFile(site.NginxConf, []byte(confContent), 0644); err != nil {
		return fmt.Errorf("write nginx conf: %w", err)
	}
	return s.reloadNginx(context.Background())
}

// DeleteSnippet 删除片段并 reload
func (s *Service) DeleteSnippet(siteID uint, name string) error {
	site, err := s.GetSite(siteID)
	if err != nil {
		return err
	}
	if !snippetNameRegex.MatchString(name) {
		return fmt.Errorf("无效的片段名")
	}
	path := filepath.Join(snippetDirOf(site.PrimaryDomain), name+".conf")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete snippet: %w", err)
	}
	return s.reloadNginx(context.Background())
}

// CreateBasicAuthUser 生成 BasicAuth 用户（bcrypt，nginx htpasswd 兼容）
// 返回 htpasswd 行 "user:hash"
func CreateBasicAuthUser(username, password string) (string, error) {
	if username == "" || strings.ContainsAny(username, ":\"'\\ \n") {
		return "", fmt.Errorf("用户名非法")
	}
	if len(password) < 6 {
		return "", fmt.Errorf("密码至少 6 位")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return username + ":" + string(hash), nil
}

// SetBasicAuthSnippet 写入 BasicAuth 片段 + htpasswd 文件
func (s *Service) SetBasicAuthSnippet(siteID uint, users []string) error {
	site, err := s.GetSite(siteID)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return fmt.Errorf("至少提供一个 htpasswd 用户行")
	}
	dir := snippetDirOf(site.PrimaryDomain)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	htpasswdPath := filepath.Join(dir, "htpasswd")
	if err := os.WriteFile(htpasswdPath, []byte(strings.Join(users, "\n")+"\n"), 0640); err != nil {
		return fmt.Errorf("write htpasswd: %w", err)
	}
	content, err := nginx.BuildBasicAuthSnippet("Restricted", htpasswdPath)
	if err != nil {
		return err
	}
	return s.SetSnippet(siteID, "basic_auth", content)
}

// AppendBasicAuthUser 追加 BasicAuth 用户（W-009）：读现有 htpasswd → 追加 bcrypt 行 → 重写片段
func (s *Service) AppendBasicAuthUser(siteID uint, username, password string) error {
	line, err := CreateBasicAuthUser(username, password)
	if err != nil {
		return err
	}
	site, err := s.GetSite(siteID)
	if err != nil {
		return err
	}
	htpasswdPath := filepath.Join(snippetDirOf(site.PrimaryDomain), "htpasswd")
	var users []string
	if data, err := os.ReadFile(htpasswdPath); err == nil {
		for _, l := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if l != "" {
				users = append(users, l)
			}
		}
	}
	users = append(users, line)
	return s.SetBasicAuthSnippet(siteID, users)
}

// confDirBase 返回 nginx conf.d 根目录（与 service.go nginxConfDir 一致，避免私有函数依赖）
func confDirBase() string {
	return nginxConfDir()
}

// nginxGenerate 包装 nginx.Generate（pkg），便于测试替换
var nginxGenerate = func(site *model.Site) (string, error) {
	return nginx.Generate(site)
}
