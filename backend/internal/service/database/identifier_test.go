package database

import (
	"strings"
	"testing"
)

// TestValidateMySQLIdentifier_Valid 测试合法标识符通过校验
func TestValidateMySQLIdentifier_Valid(t *testing.T) {
	validNames := []string{
		"mydb",
		"my_db",
		"_database",
		"db123",
		"a",
		"db$pecial",
		"UPPER_CASE",
	}

	for _, name := range validNames {
		if err := validateMySQLIdentifier(name); err != nil {
			t.Errorf("valid identifier %q should pass: %v", name, err)
		}
	}
}

// TestValidateMySQLIdentifier_Empty 测试空字符串报错
func TestValidateMySQLIdentifier_Empty(t *testing.T) {
	if err := validateMySQLIdentifier(""); err == nil {
		t.Error("empty identifier should return error")
	}
}

// TestValidateMySQLIdentifier_SpecialChars 测试含特殊字符报错
func TestValidateMySQLIdentifier_SpecialChars(t *testing.T) {
	invalidNames := []string{
		"my db",  // 空格
		"my-db",  // 连字符
		"my.db",  // 点号
		"my'db",  // 单引号
		`my"db`,  // 双引号
		"my;db",  // 分号
		"my/db",  // 斜杠
		"my\\db", // 反斜杠
		"my`db",  // 反引号
	}

	for _, name := range invalidNames {
		if err := validateMySQLIdentifier(name); err == nil {
			t.Errorf("invalid identifier %q should return error", name)
		}
	}
}

// TestValidateMySQLIdentifier_StartsWithDigit 测试以数字开头报错
func TestValidateMySQLIdentifier_StartsWithDigit(t *testing.T) {
	if err := validateMySQLIdentifier("123db"); err == nil {
		t.Error("identifier starting with digit should return error")
	}
}

// TestValidateMySQLIdentifier_TooLong 测试超长标识符报错
func TestValidateMySQLIdentifier_TooLong(t *testing.T) {
	// MySQL 标识符最大 64 字符，65 个字符应失败
	longName := strings.Repeat("a", 65)
	if err := validateMySQLIdentifier(longName); err == nil {
		t.Error("identifier longer than 64 chars should return error")
	}

	// 64 个字符应通过
	validName := strings.Repeat("a", 64)
	if err := validateMySQLIdentifier(validName); err != nil {
		t.Errorf("64-char identifier should pass: %v", err)
	}
}

// TestValidateMySQLHost 测试 MySQL host 校验
func TestValidateMySQLHost(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{"empty is allowed (default %)", "", false},
		{"localhost", "localhost", false},
		{"percent wildcard", "%", false},
		{"IP address", "192.168.1.100", false},
		{"domain", "db.example.com", false},
		{"with dash", "my-host.local", false},
		{"with semicolon", "localhost; DROP TABLE", true},
		{"with space", "local host", true},
		{"with quote", "local'host", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMySQLHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMySQLHost(%q) error = %v, wantErr %v", tt.host, err, tt.wantErr)
			}
		})
	}
}

// TestQuoteMySQLIdentifier 测试标识符引用
func TestQuoteMySQLIdentifier(t *testing.T) {
	// 普通标识符
	result := quoteMySQLIdentifier("mydb")
	if result != "`mydb`" {
		t.Errorf("expected `mydb`, got %s", result)
	}

	// 含反引号的标识符应转义
	result = quoteMySQLIdentifier("my`db")
	if result != "`my``db`" {
		t.Errorf("expected `my``db`, got %s", result)
	}

	// 空字符串
	result = quoteMySQLIdentifier("")
	if result != "``" {
		t.Errorf("expected ``, got %s", result)
	}
}
