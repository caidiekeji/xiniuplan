package database

import (
	"fmt"
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
		return fmt.Errorf("标识符不能为空")
	}
	if !mysqlIdentifierRegex.MatchString(name) {
		return fmt.Errorf("无效的标识符: %s（仅允许字母、数字、下划线，且不能以数字开头）", name)
	}
	return nil
}

// quoteMySQLIdentifier 用反引号引用 MySQL 标识符，并转义内部反引号
func quoteMySQLIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

// validateMySQLHost 校验 MySQL 用户 host 字段
// 允许 %、localhost、IP 地址、域名
var mysqlHostRegex = regexp.MustCompile(`^[%a-zA-Z0-9._-]+$`)

func validateMySQLHost(host string) error {
	if host == "" {
		return nil // 空值使用默认 %
	}
	if !mysqlHostRegex.MatchString(host) {
		return fmt.Errorf("无效的 host: %s", host)
	}
	return nil
}

// charsetRegex 校验 MySQL 字符集名称
var charsetRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// collationRegex 校验 MySQL 排序规则名称
var collationRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
