package database

import (
	"regexp"
	"strings"
)

// mysqlIdentifierRegex MySQL 标识符合法字符：字母、数字、下划线、$
// 长度 1-64（MySQL 标识符最大长度）
var mysqlIdentifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_$]{0,63}$`)

// validateMySQLIdentifier 校验 MySQL 标识符（数据库名、表名、用户名）
// 防止通过标识符进行 SQL 注入
func validateMySQLIdentifier(name string) error {
	if name == "" {
		return NewInvalidRequestError("标识符不能为空")
	}
	if !mysqlIdentifierRegex.MatchString(name) {
		return NewInvalidRequestError("无效的标识符: %s（仅允许字母、数字、下划线，且不能以数字开头）", name)
	}
	return nil
}

// quoteMySQLIdentifier 用反引号引用 MySQL 标识符，并转义内部反引号
func quoteMySQLIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

// quoteMySQLString 用单引号引用 MySQL 字符串字面量，并转义特殊字符。
//
// 用于 CREATE USER 这类语法：MySQL 的账号名写作 'user'@'host'，属于字符串
// 字面量而非标识符 —— 用反引号包裹会直接报 Error 1064 语法错误。
//
// 为什么密码必须内联而不能用 ? 占位符：MySQL 不支持为 CREATE USER 做服务端
// 预编译（实测 go-sql-driver 走 server-side prepare 时会把 ? 原样发给服务端，
// 报 Error 1064 语法错误，创建用户 100% 失败）。开启 interpolateParams 可以让
// 驱动做客户端插值，但它会让所有查询失去服务端预编译，得不偿失。
// 因此这里自行转义后内联 —— 转义规则与驱动 escapeStringBackslash 一致，
// 已实测：含单引号/反斜杠/中文/制表符/换行的密码均能正确写入并可登录。
//
// 调用前必须先过 validateMySQLIdentifier / validateMySQLHost 白名单校验。
//
// 注意：本函数按「反斜杠转义开启」模式转义（MySQL 默认 sql_mode）。
// 若服务器开启了 NO_BACKSLASH_ESCAPES，需改用 quoteMySQLStringNoBackslash。
func quoteMySQLString(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('\'')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`''`)
		case 0:
			b.WriteString(`\0`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case 0x1a:
			b.WriteString(`\Z`)
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte('\'')
	return b.String()
}

// quoteMySQLStringNoBackslash 引用 MySQL 字符串字面量，仅转义单引号。
//
// 用于服务器开启 NO_BACKSLASH_ESCAPES 的场景：此时反斜杠不是转义符，
// 写成 \\ 会被当成两个真实反斜杠，导致密码错误。此时只能用 ” 翻倍。
func quoteMySQLStringNoBackslash(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('\'')
	for i := 0; i < len(s); i++ {
		if c := s[i]; c == '\'' {
			b.WriteString(`''`)
		} else {
			b.WriteByte(c)
		}
	}
	b.WriteByte('\'')
	return b.String()
}

// validateMySQLHost 校验 MySQL 用户 host 字段
// 允许 %、localhost、IP 地址、域名
var mysqlHostRegex = regexp.MustCompile(`^[%a-zA-Z0-9._-]+$`)

func validateMySQLHost(host string) error {
	if host == "" {
		return nil // 空值使用默认 %
	}
	if !mysqlHostRegex.MatchString(host) {
		return NewInvalidRequestError("无效的 host: %s", host)
	}
	return nil
}

// charsetRegex 校验 MySQL 字符集名称
var charsetRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// collationRegex 校验 MySQL 排序规则名称
var collationRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
