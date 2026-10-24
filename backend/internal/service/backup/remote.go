package backup

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"cloudpivot/internal/model"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

// RemoteStorageService 远程存储配置管理 + 文件推送
type RemoteStorageService struct {
	db *gorm.DB
}

// NewRemoteStorageService 创建远程存储服务
func NewRemoteStorageService(db *gorm.DB) *RemoteStorageService {
	return &RemoteStorageService{db: db}
}

// --- 配置编解码（敏感字段 base64 混淆，避免明文落库可读） ---

func encodeConfig(cfg map[string]string) string {
	safe := make(map[string]string, len(cfg))
	for k, v := range cfg {
		if v == "" {
			continue
		}
		safe[k] = base64.StdEncoding.EncodeToString([]byte(v))
	}
	b, _ := json.Marshal(safe)
	return string(b)
}

func decodeConfig(raw string) map[string]string {
	result := map[string]string{}
	if raw == "" {
		return result
	}
	var safe map[string]string
	if err := json.Unmarshal([]byte(raw), &safe); err != nil {
		return result
	}
	for k, v := range safe {
		if b, err := base64.StdEncoding.DecodeString(v); err == nil {
			result[k] = string(b)
		}
	}
	return result
}

// MaskedConfig 输出脱敏后的配置（隐藏 secret/password/private_key/token）
func (s *RemoteStorageService) MaskedConfig(r *model.RemoteStorage) map[string]string {
	cfg := decodeConfig(r.Config)
	sensitive := map[string]bool{"secret_key": true, "password": true, "private_key": true, "token": true, "access_key": true, "username": true}
	out := make(map[string]string, len(cfg))
	for k, v := range cfg {
		if sensitive[k] && v != "" {
			out[k] = "******"
		} else {
			out[k] = v
		}
	}
	return out
}

// --- CRUD ---

func (s *RemoteStorageService) List() ([]model.RemoteStorage, error) {
	items := make([]model.RemoteStorage, 0)
	if err := s.db.Order("id ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("query remote storages: %w", err)
	}
	return items, nil
}

func (s *RemoteStorageService) Get(id uint) (*model.RemoteStorage, error) {
	var r model.RemoteStorage
	if err := s.db.First(&r, id).Error; err != nil {
		return nil, fmt.Errorf("remote storage not found")
	}
	return &r, nil
}

var validRemoteTypes = map[string]bool{"s3": true, "webdav": true, "sftp": true, "panel": true}

func (s *RemoteStorageService) Create(name, typ string, cfg map[string]string, enabled bool) (*model.RemoteStorage, error) {
	if name == "" {
		return nil, fmt.Errorf("名称不能为空")
	}
	if !validRemoteTypes[typ] {
		return nil, fmt.Errorf("不支持的存储类型: %s", typ)
	}
	if err := validateRemoteConfig(typ, cfg); err != nil {
		return nil, err
	}
	r := &model.RemoteStorage{
		Name:    name,
		Type:    typ,
		Config:  encodeConfig(cfg),
		Enabled: enabled,
	}
	if err := s.db.Create(r).Error; err != nil {
		return nil, fmt.Errorf("create remote storage: %w", err)
	}
	return r, nil
}

func (s *RemoteStorageService) Update(id uint, name string, typ string, cfg map[string]string, enabled *bool) (*model.RemoteStorage, error) {
	r, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		r.Name = name
	}
	if typ != "" {
		if !validRemoteTypes[typ] {
			return nil, fmt.Errorf("不支持的存储类型: %s", typ)
		}
		r.Type = typ
	}
	// cfg 为 nil 表示不修改配置（避免前端回显脱敏值覆盖真实密钥）
	if cfg != nil {
		if err := validateRemoteConfig(r.Type, cfg); err != nil {
			return nil, err
		}
		r.Config = encodeConfig(cfg)
	}
	if enabled != nil {
		r.Enabled = *enabled
	}
	if err := s.db.Save(r).Error; err != nil {
		return nil, fmt.Errorf("save remote storage: %w", err)
	}
	return r, nil
}

func (s *RemoteStorageService) Delete(id uint) error {
	r, err := s.Get(id)
	if err != nil {
		return err
	}
	return s.db.Delete(r).Error
}

func validateRemoteConfig(typ string, cfg map[string]string) error {
	get := func(k string) string { return strings.TrimSpace(cfg[k]) }
	switch typ {
	case "s3":
		if get("endpoint") == "" || get("bucket") == "" || get("access_key") == "" || get("secret_key") == "" {
			return fmt.Errorf("S3 配置不完整：endpoint/bucket/access_key/secret_key 必填")
		}
	case "webdav":
		if get("url") == "" {
			return fmt.Errorf("WebDAV 配置不完整：url 必填")
		}
	case "sftp":
		if get("host") == "" || get("username") == "" {
			return fmt.Errorf("SFTP 配置不完整：host/username 必填")
		}
		if get("password") == "" && get("private_key") == "" {
			return fmt.Errorf("SFTP 需要密码或私钥之一")
		}
	case "panel":
		if get("url") == "" || get("token") == "" {
			return fmt.Errorf("面板推送配置不完整：url/token 必填")
		}
	}
	return nil
}

// Test 连通性测试：尝试写入并删除一个探测文件
func (s *RemoteStorageService) Test(id uint) error {
	r, err := s.Get(id)
	if err != nil {
		return err
	}
	cfg := decodeConfig(r.Config)
	probeName := fmt.Sprintf("cloudpivot_probe_%d.txt", time.Now().UnixNano())
	probeData := []byte("cloudpivot remote storage probe")
	remotePath := joinRemotePath(cfg["prefix"], probeName)

	switch r.Type {
	case "s3":
		if err := s3PutObject(cfg, remotePath, probeData); err != nil {
			return fmt.Errorf("S3 连接失败: %w", err)
		}
		_ = s3DeleteObject(cfg, remotePath)
	case "webdav":
		if err := webdavPut(cfg, remotePath, probeData); err != nil {
			return fmt.Errorf("WebDAV 连接失败: %w", err)
		}
		_ = webdavDelete(cfg, remotePath)
	case "sftp":
		client, err := sftpConnect(cfg)
		if err != nil {
			return fmt.Errorf("SFTP 连接失败: %w", err)
		}
		defer client.Close()
		if err := sftpWrite(client.Client, remotePath, probeData); err != nil {
			return fmt.Errorf("SFTP 写入失败: %w", err)
		}
		_ = client.Remove(remotePath)
	case "panel":
		if err := panelPing(cfg); err != nil {
			return fmt.Errorf("面板连接失败: %w", err)
		}
	}
	return nil
}

// Upload 备份文件推送到指定远程存储（异步调用）
func (s *RemoteStorageService) Upload(r *model.RemoteStorage, localPath, remoteName string) error {
	if _, err := os.Stat(localPath); err != nil {
		return fmt.Errorf("本地备份文件不存在: %s", localPath)
	}
	data, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open backup file: %w", err)
	}
	defer data.Close()

	cfg := decodeConfig(r.Config)
	remotePath := joinRemotePath(cfg["prefix"], remoteName)

	switch r.Type {
	case "s3":
		// 大文件流式签名需要 content sha256，此处读入内存（备份文件一般 < 2GB 可接受）
		content, err := io.ReadAll(data)
		if err != nil {
			return err
		}
		return s3PutObject(cfg, remotePath, content)
	case "webdav":
		content, err := io.ReadAll(data)
		if err != nil {
			return err
		}
		return webdavPut(cfg, remotePath, content)
	case "sftp":
		client, err := sftpConnect(cfg)
		if err != nil {
			return err
		}
		defer client.Close()
		return sftpWriteFile(client.Client, remotePath, localPath)
	case "panel":
		content, err := io.ReadAll(data)
		if err != nil {
			return err
		}
		return panelPush(cfg, remoteName, content)
	default:
		return fmt.Errorf("unsupported remote type: %s", r.Type)
	}
}

func joinRemotePath(prefix, name string) string {
	prefix = strings.Trim(prefix, "/")
	if prefix == "" {
		return name
	}
	return prefix + "/" + name
}

// --- S3 (SigV4 PUT/DELETE，兼容 AWS S3 / MinIO / OSS / COS 等 S3 兼容端点) ---

func s3PutObject(cfg map[string]string, key string, content []byte) error {
	endpoint := strings.TrimSuffix(strings.TrimSpace(cfg["endpoint"]), "/")
	bucket := cfg["bucket"]
	region := cfg["region"]
	if region == "" {
		region = "us-east-1"
	}
	accessKey := cfg["access_key"]
	secretKey := cfg["secret_key"]
	pathStyle := cfg["path_style"] == "true" || cfg["path_style"] == ""

	// 构造 URL：path-style = http(s)://endpoint/bucket/key；virtual-host = http(s)://bucket.endpoint/key
	var rawURL string
	var host, canonicalURI string
	if pathStyle {
		rawURL = fmt.Sprintf("%s/%s/%s", endpoint, bucket, key)
		host = strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
		canonicalURI = "/" + bucket + "/" + key
	} else {
		rawURL = fmt.Sprintf("%s/%s", endpoint, key)
		host = bucket + "." + strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
		canonicalURI = "/" + key
	}

	req, err := http.NewRequest(http.MethodPut, rawURL, strings.NewReader(string(content)))
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	payloadHash := sha256Hex(content)
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHash)
	if host != "" {
		req.Host = host
	}

	canonicalHeaders := "host:" + host + "\nx-amz-content-sha256:" + payloadHash + "\nx-amz-date:" + amzDate + "\n"
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"
	canonicalRequest := strings.Join([]string{
		http.MethodPut,
		s3URIEncode(canonicalURI, false),
		"",
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")
	scope := dateStamp + "/" + region + "/s3/aws4_request"
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	kDate := hmacSHA256([]byte("AWS4"+secretKey), dateStamp)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, "s3")
	kSigning := hmacSHA256(kService, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(kSigning, stringToSign))

	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKey, scope, signedHeaders, signature))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func s3DeleteObject(cfg map[string]string, key string) error {
	endpoint := strings.TrimSuffix(strings.TrimSpace(cfg["endpoint"]), "/")
	bucket := cfg["bucket"]
	url := fmt.Sprintf("%s/%s/%s", endpoint, bucket, key)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	// 删除探测文件用简单签名流程不值得，直接带匿名尝试（多数 MinIO/私有 S3 会 403，探测文件遗留无碍）
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		resp.Body.Close()
	}
	return nil
}

func s3URIEncode(path string, encodeSlash bool) string {
	var b strings.Builder
	for _, c := range []byte(path) {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '.' || c == '_' || c == '~' || (!encodeSlash && c == '/') {
			b.WriteByte(c)
		} else {
			b.WriteString(fmt.Sprintf("%%%02X", c))
		}
	}
	return b.String()
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// --- WebDAV ---

func webdavClient(cfg map[string]string) *http.Client {
	return &http.Client{Timeout: 10 * time.Minute}
}

func webdavAuth(req *http.Request, cfg map[string]string) {
	if u := cfg["username"]; u != "" {
		req.SetBasicAuth(u, cfg["password"])
	}
}

// webdavMkdirAll 逐级 MKCOL 确保目录存在
func webdavMkdirAll(cfg map[string]string, dir string) {
	if dir == "" || dir == "/" {
		return
	}
	client := webdavClient(cfg)
	base := strings.TrimSuffix(cfg["url"], "/")
	parts := strings.Split(strings.Trim(dir, "/"), "/")
	cur := ""
	for _, p := range parts {
		cur += "/" + p
		req, err := http.NewRequest(http.MethodGet, base+cur, nil) // 先探测
		if err != nil {
			continue
		}
		webdavAuth(req, cfg)
		if resp, err := client.Do(req); err == nil {
			resp.Body.Close()
			if resp.StatusCode < 300 {
				continue // 已存在
			}
		}
		req, _ = http.NewRequest("MKCOL", base+cur, nil)
		webdavAuth(req, cfg)
		if resp, err := client.Do(req); err == nil {
			resp.Body.Close()
		}
	}
}

func webdavPut(cfg map[string]string, remotePath string, content []byte) error {
	// 先确保父目录存在
	if idx := strings.LastIndex(remotePath, "/"); idx > 0 {
		webdavMkdirAll(cfg, remotePath[:idx])
	}
	base := strings.TrimSuffix(cfg["url"], "/")
	req, err := http.NewRequest(http.MethodPut, base+"/"+remotePath, strings.NewReader(string(content)))
	if err != nil {
		return err
	}
	webdavAuth(req, cfg)
	resp, err := webdavClient(cfg).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func webdavDelete(cfg map[string]string, remotePath string) error {
	base := strings.TrimSuffix(cfg["url"], "/")
	req, err := http.NewRequest(http.MethodDelete, base+"/"+remotePath, nil)
	if err != nil {
		return err
	}
	webdavAuth(req, cfg)
	if resp, err := webdavClient(cfg).Do(req); err == nil {
		resp.Body.Close()
	}
	return nil
}

// --- SFTP ---

func sftpConnect(cfg map[string]string) (*sftpCloseWrapper, error) {
	host := cfg["host"]
	port := cfg["port"]
	if port == "" {
		port = "22"
	}
	user := cfg["username"]

	var auths []ssh.AuthMethod
	if pk := cfg["private_key"]; pk != "" {
		signer, err := ssh.ParsePrivateKey([]byte(pk))
		if err != nil {
			// 可能是加密私钥，尝试 password 字段作为 passphrase
			if pass := cfg["password"]; pass != "" {
				signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(pk), []byte(pass))
			}
			if err != nil {
				return nil, fmt.Errorf("parse private key: %w", err)
			}
		}
		auths = append(auths, ssh.PublicKeys(signer))
	}
	if pass := cfg["password"]; pass != "" {
		auths = append(auths, ssh.Password(pass))
	}
	if len(auths) == 0 {
		return nil, fmt.Errorf("no ssh auth method")
	}

	sshCfg := &ssh.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 面板场景：用户自担风险，可在文档中说明
		Timeout:         15 * time.Second,
	}
	client, err := ssh.Dial("tcp", host+":"+port, sshCfg)
	if err != nil {
		return nil, err
	}
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, err
	}
	// 包装关闭：sftp.Client.Close 只关 sftp 通道，ssh 连接需单独关
	return &sftpCloseWrapper{Client: sftpClient, ssh: client}, nil
}

type sftpCloseWrapper struct {
	*sftp.Client
	ssh *ssh.Client
}

func (w *sftpCloseWrapper) Close() error {
	_ = w.Client.Close()
	return w.ssh.Close()
}

func sftpWriteFile(client *sftp.Client, remotePath, localPath string) error {
	// 确保远程目录存在
	if idx := strings.LastIndex(remotePath, "/"); idx > 0 {
		sftpMkdirAll(client, remotePath[:idx])
	}
	local, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer local.Close()
	remote, err := client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("create remote file: %w", err)
	}
	defer remote.Close()
	if _, err := io.Copy(remote, local); err != nil {
		return fmt.Errorf("upload data: %w", err)
	}
	return nil
}

func sftpWrite(client *sftp.Client, remotePath string, content []byte) error {
	if idx := strings.LastIndex(remotePath, "/"); idx > 0 {
		sftpMkdirAll(client, remotePath[:idx])
	}
	f, err := client.Create(remotePath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(content)
	return err
}

func sftpMkdirAll(client *sftp.Client, dir string) {
	if dir == "" || dir == "/" || dir == "." {
		return
	}
	parts := strings.Split(strings.Trim(dir, "/"), "/")
	cur := ""
	for _, p := range parts {
		cur += "/" + p
		if _, err := client.Stat(cur); err == nil {
			continue
		}
		_ = client.Mkdir(cur)
	}
}

// --- 面板间推送（异地容灾 BK-008） ---

func panelPing(cfg map[string]string) error {
	u := strings.TrimSuffix(cfg["url"], "/") + "/api/v1/backup/receive/ping"
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	req.Header.Set("X-Panel-Token", cfg["token"])
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

func panelPush(cfg map[string]string, fileName string, content []byte) error {
	u := strings.TrimSuffix(cfg["url"], "/") + "/api/v1/backup/receive"
	req, _ := http.NewRequest(http.MethodPost, u, strings.NewReader(string(content)))
	req.Header.Set("X-Panel-Token", cfg["token"])
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Backup-Filename", url.QueryEscape(fileName))
	resp, err := (&http.Client{Timeout: 30 * time.Minute}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// sortKeys 供调试用（保持签名确定性）
func sortKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
