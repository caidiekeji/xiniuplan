package database

import (
	"errors"
	"strings"
	"testing"
)

// TestQuoteMySQLString 验证 MySQL 字符串字面量引用与转义。
//
// 2026-10-24 事故：CREATE USER 曾写成 `CREATE USER ?@? IDENTIFIED BY ?`，
// MySQL 驱动不认这种占位符，服务端直接报 Error 1064 语法错误；
// 改用反引号包裹后同样报错 —— 账号名在 MySQL 里是字符串字面量（'user'@'host'），
// 不是标识符。回归测试锁死「用单引号 + 完整转义」这条规则。
func TestQuoteMySQLString(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"普通用户名", "testuser", "'testuser'"},
		{"host 通配符", "%", "'%'"},
		{"内含单引号需转义", "o'brien", "'o''brien'"},
		{"内含反斜杠需转义", `back\slash`, `'back\\slash'`},
		{"内含单引号与反斜杠", `a'b\c`, `'a''b\\c'`},
		{"换行符需转义", "a\nb", `'a\nb'`},
		{"回车符需转义", "a\rb", `'a\rb'`},
		{"空字符串", "", "''"},
		{"中文原样保留", "中文密码", "'中文密码'"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := quoteMySQLString(tc.in)
			if got != tc.want {
				t.Errorf("quoteMySQLString(%q) = %s, 期望 %s", tc.in, got, tc.want)
			}
		})
	}
}

// TestQuoteMySQLStringRoundTrip 验证转义后能被 MySQL 正确解析回原值。
//
// 这是最 meaningful 的检查：转义后的字面量交给 MySQL 解析器还原，
// 必须与原始输入完全一致 —— 既证明密码能正确写入，
// 也证明攻击载荷无法逃逸出字符串字面量。
func TestQuoteMySQLStringRoundTrip(t *testing.T) {
	inputs := []string{
		"Pw@12345",
		"a'b\\c\"d;--",
		"p@ss w0rd!#$%^&*()",
		"中文密码测试",
		"tab\there",
		"newline\nhere",
		"'; DROP DATABASE app; --",
		"' OR '1'='1",
		"a\\'; DROP DATABASE app; --",
		"x'; --",
		"",
	}
	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			got := parseMySQLLiteral(quoteMySQLString(in))
			if got != in {
				t.Errorf("往返解析不一致:\n  输入: %q\n  解析: %q", in, got)
			}
		})
	}
}

// parseMySQLLiteral 模拟 MySQL 对单引号字符串字面量的解析（反斜杠转义开启）。
//
// 用于验证转义函数的正确性：MySQL 服务端会用同样的规则解析字面量，
// 只有往返一致才说明密码被正确写入、且注入载荷无法逃逸。
func parseMySQLLiteral(literal string) string {
	if len(literal) < 2 || literal[0] != '\'' || literal[len(literal)-1] != '\'' {
		return "<格式错误>"
	}
	inner := literal[1 : len(literal)-1]
	var b strings.Builder
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		switch c {
		case '\\':
			// 反斜杠转义
			i++
			if i >= len(inner) {
				b.WriteByte('\\')
				break
			}
			switch inner[i] {
			case '0':
				b.WriteByte(0)
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 'Z':
				b.WriteByte(0x1a)
			default:
				b.WriteByte(inner[i])
			}
		case '\'':
			// '' 表示转义后的单引号
			if i+1 < len(inner) && inner[i+1] == '\'' {
				b.WriteByte('\'')
				i++
				continue
			}
			// 孤立的单引号：MySQL 会报错，这里保留原样以便测试发现
			b.WriteString("<未闭合单引号>")
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// TestQuoteMySQLStringNoBackslash 验证 NO_BACKSLASH_ESCAPES 模式下的转义。
//
// 该模式下反斜杠不是转义符，只能用 ” 翻倍转义单引号；
// 若误用默认模式（含 \\ 转义），含反斜杠的密码会被写成两个反斜杠。
func TestQuoteMySQLStringNoBackslash(t *testing.T) {
	inputs := []string{
		"Pw@12345",
		"a'b",
		`back\slash`,
		"中文密码",
		"'; DROP DATABASE app; --",
	}
	for _, in := range inputs {
		got := quoteMySQLStringNoBackslash(in)
		if parsed := parseMySQLLiteralNoBackslash(got); parsed != in {
			t.Errorf("往返解析不一致:\n  输入: %q\n  字面量: %s\n  解析: %q", in, got, parsed)
		}
		// 反斜杠必须原样保留（该模式下它不是转义符）
		if strings.Contains(in, `\`) && !strings.Contains(got, `\`) {
			t.Errorf("反斜杠被错误转义: 输入 %q -> %s", in, got)
		}
	}
}

// parseMySQLLiteralNoBackslash 模拟 NO_BACKSLASH_ESCAPES 模式下的解析
func parseMySQLLiteralNoBackslash(literal string) string {
	if len(literal) < 2 || literal[0] != '\'' || literal[len(literal)-1] != '\'' {
		return "<格式错误>"
	}
	inner := literal[1 : len(literal)-1]
	var b strings.Builder
	for i := 0; i < len(inner); i++ {
		if inner[i] == '\'' {
			if i+1 < len(inner) && inner[i+1] == '\'' {
				b.WriteByte('\'')
				i++
				continue
			}
			b.WriteString("<未闭合单引号>")
			continue
		}
		b.WriteByte(inner[i])
	}
	return b.String()
}

// TestQuoteMySQLStringNotIdentifier 确认账号名走字符串而非反引号标识符。
// 反引号版本是导致 Error 1064 的直接原因。
func TestQuoteMySQLStringNotIdentifier(t *testing.T) {
	got := quoteMySQLString("testuser")
	if strings.HasPrefix(got, "`") {
		t.Errorf("账号名不应使用反引号标识符引用，得到 %s", got)
	}
	if !strings.HasPrefix(got, "'") {
		t.Errorf("账号名应使用单引号字符串引用，得到 %s", got)
	}
}

// TestCreateUserRejectsInvalidInput 验证创建用户在连库之前就拒绝非法输入。
//
// 无效输入必须被拦截在连接数据库之前 —— 早期版本直接拼 SQL，
// 非法用户名会一路送到 MySQL 服务端。
func TestCreateUserRejectsInvalidInput(t *testing.T) {
	svc := NewUserService(nil)

	// 非法用户名
	if err := svc.CreateUser(1, "bad user", "pw", "%"); err == nil {
		t.Error("非法用户名应被拒绝")
	} else if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("应标记为 ErrInvalidRequest，实际: %v", err)
	}

	// 非法 host
	if err := svc.CreateUser(1, "gooduser", "pw", "bad host"); err == nil {
		t.Error("非法 host 应被拒绝")
	} else if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("应标记为 ErrInvalidRequest，实际: %v", err)
	}

	// 空密码
	if err := svc.CreateUser(1, "gooduser", "", "%"); err == nil {
		t.Error("空密码应被拒绝")
	} else if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("应标记为 ErrInvalidRequest，实际: %v", err)
	}
}

// TestGrantPrivilegesRejectsUnsupportedPrivilege 验证权限白名单拦截
func TestGrantPrivilegesRejectsUnsupportedPrivilege(t *testing.T) {
	svc := NewUserService(nil)

	err := svc.GrantPrivileges(1, "testuser", "%", "app", []string{"DROP"})
	if err == nil {
		t.Fatal("DROP 权限应被白名单拦截")
	}
	if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("应标记为 ErrInvalidRequest，实际: %v", err)
	}
}

// TestValidateErrorsAreClientErrors 验证校验类错误统一标记为客户端错误，
// API 层才能把它们翻译成 400 而不是 500。
func TestValidateErrorsAreClientErrors(t *testing.T) {
	if err := validateMySQLIdentifier("bad name"); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("标识符校验错误应标记 ErrInvalidRequest，实际: %v", err)
	}
	if err := validateMySQLHost("bad host"); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("host 校验错误应标记 ErrInvalidRequest，实际: %v", err)
	}
	if err := NewUnsupportedTypeError("redis", "该操作仅支持 MySQL 实例"); !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("类型不支持错误应标记 ErrUnsupportedType，实际: %v", err)
	}
}
