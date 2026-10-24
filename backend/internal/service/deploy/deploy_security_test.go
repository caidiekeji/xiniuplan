package deploy

import (
	"cloudpivot/internal/dto"
	"strings"
	"testing"
)

// TestShqEscapesSingleQuote 验证 shell 参数引用的正确性。
//
// 2026-10-24 事故：shq() 直接用 ' 包住原串，参数里带单引号时会逃逸出
// 引号造成命令注入。POSIX 单引号内不能出现单引号，正确做法是把 ' 写成 '\”。
func TestShqEscapesSingleQuote(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"普通路径", "/var/www/app", `'/var/www/app'`},
		{"含单引号", "/tmp/a'b", `'/tmp/a'\''b'`},
		{"仅单引号", "'", `''\'''`},
		{"含分号", "/tmp/a;touch /tmp/x", `'/tmp/a;touch /tmp/x'`},
		{"含反引号", "/tmp/`id`", "'/tmp/`id`'"},
		{"空串", "", "''"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shq(tc.in)
			if got != tc.want {
				t.Errorf("shq(%q) = %s, 期望 %s", tc.in, got, tc.want)
			}
		})
	}
}

// TestShqPreventsQuoteEscape 验证引号包裹后无法逃逸。
//
// 转义后的参数交给 shell 解析时，原始单引号必须仍然处于引号内部。
func TestShqPreventsQuoteEscape(t *testing.T) {
	attacks := []string{
		"/tmp/a' ; touch /tmp/pwned ; echo '",
		"'$(id)'",
		"'`id`'",
	}
	for _, atk := range attacks {
		got := shq(atk)
		// 转义后单引号总数必须是偶数（成对出现）
		if strings.Count(got, "'")%2 != 0 {
			t.Errorf("转义后单引号数量为奇数: 输入 %q -> %s", atk, got)
		}
		// 反斜杠总数必须是偶数
		if strings.Count(got, `\`)%2 != 0 {
			t.Errorf("转义后反斜杠数量为奇数: 输入 %q -> %s", atk, got)
		}
	}
}

// TestValidateDeployPath 验证部署路径白名单
func TestValidateDeployPath(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"合法绝对路径", "/var/www/app", false},
		{"空值允许", "", false},
		{"相对路径拒绝", "www/app", true},
		{"含分号拒绝", "/tmp/a;touch /tmp/x", true},
		{"含反引号拒绝", "/tmp/`id`", true},
		{"含单引号拒绝", "/tmp/a'b", true},
		{"含换行拒绝", "/tmp/a\nb", true},
		{"含美元符拒绝", "/tmp/$HOME", true},
		{"含管道拒绝", "/tmp/a|b", true},
		{"含重定向拒绝", "/tmp/a>b", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDeployPath(tc.in)
			if tc.wantErr != (err != nil) {
				t.Errorf("validateDeployPath(%q) 错误状态不符: %v", tc.in, err)
			}
		})
	}
}

// TestValidateDeployName 验证应用名称白名单
func TestValidateDeployName(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"合法名称", "my-app_1", false},
		{"空名拒绝", "", true},
		{"以连字符开头拒绝", "-app", true},
		{"含分号拒绝", "app;rm", true},
		{"含空格拒绝", "my app", true},
		{"含反引号拒绝", "app`x`", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDeployName(tc.in)
			if tc.wantErr != (err != nil) {
				t.Errorf("validateDeployName(%q) 错误状态不符: %v", tc.in, err)
			}
		})
	}
}

// TestValidateDeployBranch 验证分支名白名单
func TestValidateDeployBranch(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"合法分支", "main", false},
		{"带斜杠", "feature/new-thing", false},
		{"空值允许", "", false},
		{"含分号拒绝", "main;rm -rf /", true},
		{"含空格拒绝", "my branch", true},
		{"以斜杠开头拒绝", "/main", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDeployBranch(tc.in)
			if tc.wantErr != (err != nil) {
				t.Errorf("validateDeployBranch(%q) 错误状态不符: %v", tc.in, err)
			}
		})
	}
}

// TestUpdateDoesNotClearEnabledWhenOmitted 验证更新时未传 enabled 不会把应用禁用。
//
// 2026-10-24 事故：Update 无条件写 `"enabled": req.Enabled`，而 model 里的
// Enabled 是普通 bool，前端编辑表单不带这个字段 → 绑定后是零值 false
// → 用户只是改了名字，应用就被静默禁用。改用 *bool 后未传即为 nil，跳过更新。
func TestUpdateDoesNotClearEnabledWhenOmitted(t *testing.T) {
	// 未传 enabled：指针为 nil
	var omitted dto.DeployAppUpdateRequest
	if omitted.Enabled != nil {
		t.Fatal("未传 enabled 时指针应为 nil")
	}

	// 显式传 false：指针非 nil，应能正常表达「禁用」
	explicit := false
	var provided dto.DeployAppUpdateRequest
	provided.Enabled = &explicit
	if provided.Enabled == nil {
		t.Fatal("显式传入 enabled=false 时指针不应为 nil")
	}
	if *provided.Enabled != false {
		t.Error("显式传入的 false 应被保留")
	}
}
