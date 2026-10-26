package docker

import (
	"testing"
)

// TestParseImageRef 验证镜像引用解析
func TestParseImageRef(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		wantReg  string
		wantRepo string
		wantTag  string
		wantErr  bool
	}{
		{"官方镜像无标签", "nginx", "", "nginx", "latest", false},
		{"官方镜像带标签", "nginx:1.25", "", "nginx", "1.25", false},
		{"带命名空间", "library/nginx:1.25", "", "library/nginx", "1.25", false},
		{"私有仓库带端口", "localhost:5000/x:1", "localhost:5000", "x", "1", false},
		{"域名仓库多层路径", "harbor.example.com/a/b:1", "harbor.example.com", "a/b", "1", false},
		{"带 SHA 摘要不拆 tag", "nginx@sha256:abc", "", "nginx@sha256:abc", "", false},
		{"空名报错", "", "", "", "", true},
		{"空白字符报错", "ng inx", "", "", "", true},
		{"只有仓库无路径报错", "harbor.example.com/", "harbor.example.com", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := ParseImageRef(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ParseImageRef(%q) 应报错但通过了: %+v", tc.in, p)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseImageRef(%q) 报错: %v", tc.in, err)
			}
			if p.Registry != tc.wantReg {
				t.Errorf("Registry = %q, 期望 %q", p.Registry, tc.wantReg)
			}
			if p.Repository != tc.wantRepo {
				t.Errorf("Repository = %q, 期望 %q", p.Repository, tc.wantRepo)
			}
			if p.Tag != tc.wantTag {
				t.Errorf("Tag = %q, 期望 %q", p.Tag, tc.wantTag)
			}
		})
	}
}

// TestApplyMirror 验证加速源改写
func TestApplyMirror(t *testing.T) {
	t.Run("官方镜像拼加速源", func(t *testing.T) {
		p, err := ParseImageRef("nginx:1.25")
		if err != nil {
			t.Fatal(err)
		}
		mp, err := ApplyMirror(p, "docker.m.daocloud.io")
		if err != nil {
			t.Fatal(err)
		}
		want := "docker.m.daocloud.io/library/nginx:1.25"
		if got := mp.String(); got != want {
			t.Errorf("改写结果 = %s, 期望 %s", got, want)
		}
	})

	t.Run("加速源带尾斜杠", func(t *testing.T) {
		p, _ := ParseImageRef("redis")
		mp, err := ApplyMirror(p, "docker.1panel.live/")
		if err != nil {
			t.Fatal(err)
		}
		want := "docker.1panel.live/library/redis:latest"
		if got := mp.String(); got != want {
			t.Errorf("改写结果 = %s, 期望 %s", got, want)
		}
	})

	t.Run("私有仓库镜像不改写", func(t *testing.T) {
		p, _ := ParseImageRef("harbor.example.com/a/b:1")
		if _, err := ApplyMirror(p, "docker.m.daocloud.io"); err == nil {
			t.Error("已指定仓库的镜像应拒绝改写加速源")
		}
	})

	t.Run("带命名空间镜像保持路径", func(t *testing.T) {
		p, _ := ParseImageRef("bitnami/nginx:1.25")
		mp, err := ApplyMirror(p, "docker.m.daocloud.io")
		if err != nil {
			t.Fatal(err)
		}
		want := "docker.m.daocloud.io/bitnami/nginx:1.25"
		if got := mp.String(); got != want {
			t.Errorf("改写结果 = %s, 期望 %s", got, want)
		}
	})
}

// TestNormalizeDockerHubRepo 验证 library 前缀补全
func TestNormalizeDockerHubRepo(t *testing.T) {
	cases := []struct{ in, want string }{
		{"nginx", "library/nginx"},
		{"library/nginx", "library/nginx"},
		{"bitnami/nginx", "bitnami/nginx"},
	}
	for _, tc := range cases {
		p, _ := ParseImageRef(tc.in)
		if got := NormalizeDockerHubRepo(p); got != tc.want {
			t.Errorf("NormalizeDockerHubRepo(%q) = %s, 期望 %s", tc.in, got, tc.want)
		}
	}
}

// TestParseMirrorList 验证加速源列表解析
func TestParseMirrorList(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"逗号分隔", "a.io, b.io", 2},
		{"换行分隔", "a.io\nb.io\nc.io", 3},
		{"分号分隔", "a.io;b.io", 2},
		{"空串", "", 0},
		{"含空项", "a.io,, b.io", 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseMirrorList(tc.in)
			if len(got) != tc.want {
				t.Errorf("ParseMirrorList(%q) 得到 %d 项, 期望 %d: %v", tc.in, len(got), tc.want, got)
			}
		})
	}
}

// TestParsedImageStringRoundTrip 验证 String() 能还原为合法镜像引用
func TestParsedImageStringRoundTrip(t *testing.T) {
	cases := []string{
		"nginx:latest",
		"nginx:1.25",
		"library/nginx:1.25",
		"bitnami/nginx:1.25",
		"localhost:5000/x:1",
		"harbor.example.com/a/b:1",
		"nginx@sha256:abc",
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			p, err := ParseImageRef(in)
			if err != nil {
				t.Fatal(err)
			}
			if got := p.String(); got != in {
				t.Errorf("String() = %s, 期望还原为 %s", got, in)
			}
		})
	}
}

// TestRegistryAuthEncoding 验证认证头编码
func TestRegistryAuthEncoding(t *testing.T) {
	if got := RegistryAuth("", ""); got != "" {
		t.Errorf("空凭据应返回空串, 得到 %s", got)
	}
	got := RegistryAuth("user", "pass")
	// {"username":"user","password":"pass"} 的 base64
	want := "eyJ1c2VybmFtZSI6InVzZXIiLCJwYXNzd29yZCI6InBhc3MifQ=="
	if got != want {
		t.Errorf("RegistryAuth = %s, 期望 %s", got, want)
	}
}

// TestBase64Encode 验证 base64 实现（含补位）
func TestBase64Encode(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"f", "Zg=="},
		{"fo", "Zm8="},
		{"foo", "Zm9v"},
		{"foob", "Zm9vYg=="},
		{"fooba", "Zm9vYmE="},
		{"foobar", "Zm9vYmFy"},
	}
	for _, tc := range cases {
		if got := base64Encode(tc.in); got != tc.want {
			t.Errorf("base64Encode(%q) = %s, 期望 %s", tc.in, got, tc.want)
		}
	}
}
