package docker

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RegistryConfig 镜像仓库（拉取地址）配置
//
// 国内直连 Docker Hub 经常超时，需要可配加速源；私有仓库（Harbor / ACR）
// 需要地址与账号。这两类都影响「拉哪个镜像」而不是「连哪个 Docker」，
// 因此与连接配置（DOCKER_HOST）是两套独立的东西。
type RegistryConfig struct {
	// Mirrors 加速源列表，按顺序尝试，例如
	//   docker.m.daocloud.io, docker.1panel.live
	// 留空表示直连 Docker Hub
	Mirrors []string
	// Private 私有仓库地址（如 harbor.example.com），与账号配合使用
	Private string
	// Username / Password 私有仓库凭据
	Username string
	Password string
}

// RegistryOption 供前端下拉选择的仓库选项
type RegistryOption struct {
	// ID 选中值：空串 = 直连，@private = 私有仓库，其余为加速源地址
	ID string `json:"id"`
	// Name 展示名
	Name string `json:"name"`
}

// ParsedImage 解析后的镜像引用
type ParsedImage struct {
	// Registry 镜像所属仓库地址，空表示 Docker Hub
	//   例：harbor.example.com、docker.io
	Registry string
	// Repository 镜像名（不含仓库前缀）
	//   例：library/nginx、nginx（Docker Hub 官方镜像）
	Repository string
	// Tag 标签，空表示 latest
	Tag string
	// Original 用户原始输入（用于回显与错误信息）
	Original string
}

// ParseImageRef 解析镜像引用
//
// 规则（与 Docker CLI 一致）：
//
//	nginx                     → registry="" repo="nginx" tag="latest"
//	nginx:1.25                → registry="" repo="nginx" tag="1.25"
//	library/nginx:1.25        → registry="" repo="library/nginx" tag="1.25"
//	harbor.example.com/a/b:1  → registry="harbor.example.com" repo="a/b" tag="1"
//	localhost:5000/x:1        → registry="localhost:5000" repo="x" tag="1"
//
// 判据：第一段含 '.' 或 ':' 或等于 'localhost' 时视为仓库地址，
// 否则是 Docker Hub 上的官方镜像路径（可省略 library/ 前缀）。
func ParseImageRef(ref string) (*ParsedImage, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("镜像名不能为空")
	}
	if strings.ContainsAny(ref, " \t\n\r") {
		return nil, fmt.Errorf("镜像名不能包含空白字符: %s", ref)
	}

	p := &ParsedImage{Original: ref}

	// @sha256:... 摘要形式：摘要是「不可变的引用」，与 tag 二选一。
	// 直接整体保留（含 @sha256:xxx），不再拆 tag —— 否则冒号会被误判成
	// tag 分隔符，把摘要切碎。
	if idx := strings.Index(ref, "@"); idx >= 0 {
		p.Repository = ref
		p.Tag = "" // 摘要形式不带 tag
		return p, nil
	}

	// 分离 tag
	if idx := strings.LastIndex(ref, ":"); idx >= 0 && !strings.Contains(ref[idx+1:], "/") {
		p.Tag = ref[idx+1:]
		ref = ref[:idx]
	}

	// 分离 registry
	first := ref
	rest := ""
	if idx := strings.Index(ref, "/"); idx >= 0 {
		first = ref[:idx]
		rest = ref[idx+1:]
	}
	if isRegistryHost(first) {
		p.Registry = first
		p.Repository = rest
	} else {
		p.Repository = ref
	}

	if p.Repository == "" {
		return nil, fmt.Errorf("镜像名不完整: %s", p.Original)
	}
	if p.Tag == "" && !strings.Contains(p.Repository, "@") {
		p.Tag = "latest"
	}
	return p, nil
}

// isRegistryHost 判断第一段是否是仓库地址
func isRegistryHost(s string) bool {
	if s == "localhost" {
		return true
	}
	// 含点（域名/IP）或冒号（端口）都视为仓库地址
	return strings.ContainsAny(s, ".:")
}

// NormalizeDockerHubRepo 补全 Docker Hub 官方镜像的 library/ 前缀
//
// Docker Hub 官方镜像（nginx、redis）实际路径是 library/nginx、library/redis。
// 不补也能拉（服务端会补），但补全后镜像列表里显示的名字与输入一致，
// 避免用户以为拉错了镜像。
func NormalizeDockerHubRepo(p *ParsedImage) string {
	if p.Registry != "" {
		return p.Repository
	}
	if !strings.Contains(p.Repository, "/") {
		return "library/" + p.Repository
	}
	// 含 / 但第一段是用户名空间（如 bitnami/nginx），保持原样
	return p.Repository
}

// ApplyMirror 把镜像引用改写到加速源
//
// 例：nginx:1.25 + mirror docker.m.daocloud.io
//
//	→ docker.m.daocloud.io/library/nginx:1.25
//
// 只对 Docker Hub 上的镜像生效 —— 私有仓库镜像（已含 registry）不改写，
// 改写只会让它拉不到。
func ApplyMirror(p *ParsedImage, mirror string) (*ParsedImage, error) {
	if p.Registry != "" {
		return nil, fmt.Errorf("镜像 %s 已指定仓库 %s，不使用加速源", p.Original, p.Registry)
	}
	// 摘要形式（@sha256:...）的镜像也走原地址：多数加速源不支持摘要拉取，
	// 强行改写会得到 mirror/...@sha256:... 这种加速源无法解析的引用。
	if strings.Contains(p.Repository, "@") {
		return nil, fmt.Errorf("镜像 %s 为摘要形式，不使用加速源", p.Original)
	}
	mirror = strings.TrimSpace(mirror)
	if mirror == "" {
		return nil, fmt.Errorf("加速源地址为空")
	}
	mirror = strings.TrimSuffix(mirror, "/")

	np := &ParsedImage{
		Registry:   mirror,
		Repository: NormalizeDockerHubRepo(p),
		Tag:        p.Tag,
		Original:   p.Original,
	}
	return np, nil
}

// String 还原为镜像引用字符串
func (p *ParsedImage) String() string {
	// @sha256 摘要形式：repository 已含完整引用，不能再拼 ":tag"
	if strings.Contains(p.Repository, "@") {
		s := p.Repository
		if p.Registry != "" {
			s = p.Registry + "/" + s
		}
		return s
	}
	s := p.Repository
	if p.Registry != "" {
		s = p.Registry + "/" + p.Repository
	}
	return s + ":" + p.Tag
}

// RegistryAuth 生成 Docker SDK 需要的 RegistryAuth 头（base64 JSON）
//
// 用结构体而非 map：map 的 JSON 序列化 key 顺序随机，同一份凭据每次编码
// 结果都不同，无法缓存与比对。
func RegistryAuth(username, password string) string {
	if username == "" && password == "" {
		return ""
	}
	type authPayload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	b, err := json.Marshal(authPayload{Username: username, Password: password})
	if err != nil {
		return ""
	}
	return base64Encode(string(b))
}

// base64Encode 标准 base64 编码
func base64Encode(s string) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var out strings.Builder
	data := []byte(s)
	for i := 0; i < len(data); i += 3 {
		var n uint32
		remain := len(data) - i
		n = uint32(data[i]) << 16
		if remain > 1 {
			n |= uint32(data[i+1]) << 8
		}
		if remain > 2 {
			n |= uint32(data[i+2])
		}
		out.WriteByte(alphabet[(n>>18)&0x3F])
		out.WriteByte(alphabet[(n>>12)&0x3F])
		if remain > 1 {
			out.WriteByte(alphabet[(n>>6)&0x3F])
		} else {
			out.WriteByte('=')
		}
		if remain > 2 {
			out.WriteByte(alphabet[n&0x3F])
		} else {
			out.WriteByte('=')
		}
	}
	return out.String()
}

// ParseMirrorList 解析加速源列表（逗号或换行分隔）
func ParseMirrorList(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '\n' || r == ';'
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
