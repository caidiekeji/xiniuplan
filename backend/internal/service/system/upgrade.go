package system

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// 面板一键自升级（roadmap #17）：下载 → sha256 校验 → 原子替换 → 重启
// 升级源约定：update_check_url 返回 JSON { version, download_url, sha256 }（linux amd64 二进制）

// UpgradeResult 升级执行结果
type UpgradeResult struct {
	Downloaded int64  `json:"downloaded_bytes"`
	Sha256     string `json:"sha256"`
	Message    string `json:"message"`
}

// Upgrade 执行面板自升级：
// 1. 从更新源取版本清单 → 下载新二进制 → 校验 sha256
// 2. 原子替换当前二进制（Linux 允许替换运行中的可执行文件）
// 3. 调用 systemd 重启面板；非 systemd 环境下优雅退出（由进程守护/Restart=always 拉起）
func (s *Service) Upgrade() (*UpgradeResult, error) {
	settings, _ := s.GetSettings()
	checkURL := settings["update_check_url"]
	if checkURL == "" {
		if repo := settings["update_github_repo"]; repo != "" {
			checkURL = "https://api.github.com/repos/" + repo + "/releases/latest"
		}
	}
	if checkURL == "" {
		return nil, fmt.Errorf("未配置更新源")
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, checkURL, nil)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取更新清单失败: %w", err)
	}
	defer resp.Body.Close()

	// 兼容两种清单：标准 Release JSON（GitHub）与精简 JSON {version, download_url, sha256}
	var manifest struct {
		TagName string `json:"tag_name"`
		Version string `json:"version"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
		DownloadURL string `json:"download_url"`
		SHA256      string `json:"sha256"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("更新清单解析失败: %w", err)
	}

	version := manifest.Version
	if version == "" {
		version = manifest.TagName
	}
	downloadURL := manifest.DownloadURL
	wantSHA := manifest.SHA256
	if downloadURL == "" {
		// GitHub Release：找 linux amd64 资产
		for _, a := range manifest.Assets {
			name := a.Name
			if (contains(name, "linux") && contains(name, "amd64") || contains(name, "x86_64")) && contains(name, ".tar.gz") == false && ends(name, []string{".bin", ""}) {
				downloadURL = a.URL
				break
			}
		}
		if downloadURL == "" {
			for _, a := range manifest.Assets {
				if contains(a.Name, "linux") && contains(a.Name, "amd64") {
					downloadURL = a.URL
					break
				}
			}
		}
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("更新清单中没有可用的 linux amd64 资产")
	}
	if version != "" && version == currentVersion() {
		return &UpgradeResult{Message: "已是最新版本 " + version}, nil
	}

	// 下载新二进制
	dlClient := &http.Client{Timeout: 10 * time.Minute}
	dlReq, _ := http.NewRequest(http.MethodGet, downloadURL, nil)
	if ghToken := settings["update_github_token"]; ghToken != "" && contains(downloadURL, "github") {
		dlReq.Header.Set("Authorization", "Bearer "+ghToken)
	}
	dlResp, err := dlClient.Do(dlReq)
	if err != nil {
		return nil, fmt.Errorf("下载失败: %w", err)
	}
	defer dlResp.Body.Close()
	if dlResp.StatusCode != 200 {
		return nil, fmt.Errorf("下载失败: HTTP %d", dlResp.StatusCode)
	}

	// 可能为 tar.gz——直接落盘再按扩展名处理
	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("panel-upgrade-%d.bin", time.Now().Unix()))
	out, err := os.Create(tmpPath)
	if err != nil {
		return nil, err
	}
	size, err := io.Copy(out, dlResp.Body)
	out.Close()
	if err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("下载写入失败: %w", err)
	}

	// sha256 校验（清单提供时）
	if wantSHA != "" {
		f, _ := os.Open(tmpPath)
		h := sha256.New()
		io.Copy(h, f)
		f.Close()
		got := hex.EncodeToString(h.Sum(nil))
		if got != wantSHA {
			os.Remove(tmpPath)
			return nil, fmt.Errorf("sha256 校验失败（期望 %s，实际 %s）——已中止升级", wantSHA, got)
		}
	}

	// 若为 tar.gz：解包提取二进制
	if ends(downloadURL, []string{".tar.gz", ".tgz"}) {
		extracted, err := extractBinaryFromTarGz(tmpPath)
		if err != nil {
			os.Remove(tmpPath)
			return nil, err
		}
		os.Remove(tmpPath)
		tmpPath = extracted
	}

	// 自身可执行文件路径
	selfPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	selfPath = resolveSymlink(selfPath)

	// 备份旧版本（失败回滚用）
	bakPath := selfPath + ".old"
	os.Remove(bakPath)
	if err := copyFile(selfPath, bakPath); err != nil {
		return nil, fmt.Errorf("备份旧版本失败: %w", err)
	}

	// 原子替换：mv 到同目录
	if err := copyFile(tmpPath, selfPath+".new"); err != nil {
		return nil, err
	}
	os.Chmod(selfPath+".new", 0755)
	if err := os.Rename(selfPath+".new", selfPath); err != nil {
		return nil, fmt.Errorf("替换二进制失败: %w", err)
	}
	os.Remove(tmpPath)

	// 尝试 systemd 重启（面板运行于 systemd 环境）；失败则优雅退出由 Restart=always 拉起
	if err := restartSelf(); err != nil {
		// 退出进程，依赖守护机制自动拉起（Restart=always / RestartSec=5）
		go func() {
			time.Sleep(500 * time.Millisecond)
			os.Exit(0)
		}()
		return &UpgradeResult{
			Downloaded: size, Sha256: wantSHA,
			Message: fmt.Sprintf("已升级到 %s，面板正在重启…若 10 秒后未恢复，请手动执行: mv %s %s && systemctl restart rhinopanel", version, bakPath, selfPath),
		}, nil
	}

	return &UpgradeResult{
		Downloaded: size, Sha256: wantSHA,
		Message: fmt.Sprintf("已升级到 %s 并重启完成。旧版本备份于 %s", version, bakPath),
	}, nil
}

// ===== 工具 =====

func currentVersion() string { return panelVersion }

func contains(s, sub string) bool { return strings.Contains(s, sub) }

func ends(s string, suffixes []string) bool {
	for _, suf := range suffixes {
		if suf != "" && len(s) >= len(suf) && s[len(s)-len(suf):] == suf {
			return true
		}
	}
	return false
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func resolveSymlink(p string) string {
	if real, err := filepath.EvalSymlinks(p); err == nil {
		return real
	}
	return p
}

// extractBinaryFromTarGz 从 tar.gz 中提取第一个可执行文件
// restartSelf 通过 systemd 重启面板（jie 用户需配置 sudoers 免密或 NOPASSWD）
func restartSelf() error {
	cmd := exec.Command("sudo", "-n", "systemctl", "restart", "rhinopanel")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("systemd 重启失败(%s): %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func execCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func extractBinaryFromTarGz(tarPath string) (string, error) {
	// 简化实现：使用系统 tar 命令解压到临时目录
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("panel-extract-%d", time.Now().Unix()))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", err
	}
	cmd := execCommand("tar", "-xzf", tarPath, "-C", tmpDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("解压失败: %w: %s", err, string(out))
	}
	// 找 rhinopanel 或 cloudpivot 命名的文件
	for _, name := range []string{"rhinopanel", "cloudpivot"} {
		p := filepath.Join(tmpDir, name)
		if _, err := os.Stat(p); err == nil {
			os.Chmod(p, 0755)
			return p, nil
		}
	}
	// 退而求其次：取目录下最大的文件
	var best string
	var bestSize int64
	filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && info.Size() > bestSize && info.Mode()&0111 != 0 {
			best, bestSize = path, info.Size()
		}
		return nil
	})
	if best == "" {
		return "", fmt.Errorf("压缩包中未找到可执行文件")
	}
	return best, nil
}
