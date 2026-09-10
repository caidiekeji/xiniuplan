package common

import (
	"context"
	"testing"
	"time"
)

// TestCommandExecutor_UnregisteredTemplate 测试未注册模板返回错误
func TestCommandExecutor_UnregisteredTemplate(t *testing.T) {
	exec := NewCommandExecutor(5 * time.Second)

	_, err := exec.Execute(context.Background(), "nonexistent_template", map[string]string{})
	if err == nil {
		t.Error("expected error for unregistered template, got nil")
	}
}

// TestCommandExecutor_MetaCharFiltering 测试元字符过滤
func TestCommandExecutor_MetaCharFiltering(t *testing.T) {
	exec := NewCommandExecutor(5 * time.Second)

	exec.Register(&CommandTemplate{
		Name: "echo_test",
		Bin:  "echo",
		Args: []string{"{{msg}}"},
	})

	// 包含 ; 的参数应被拒绝
	_, err := exec.Execute(context.Background(), "echo_test", map[string]string{"msg": "hello; rm -rf /"})
	if err == nil {
		t.Error("expected error for param containing ';'")
	}

	// 包含 | 的参数应被拒绝
	_, err = exec.Execute(context.Background(), "echo_test", map[string]string{"msg": "hello | cat /etc/passwd"})
	if err == nil {
		t.Error("expected error for param containing '|'")
	}

	// 包含 $ 的参数应被拒绝
	_, err = exec.Execute(context.Background(), "echo_test", map[string]string{"msg": "$(whoami)"})
	if err == nil {
		t.Error("expected error for param containing '$'")
	}

	// 包含反引号的参数应被拒绝
	_, err = exec.Execute(context.Background(), "echo_test", map[string]string{"msg": "`whoami`"})
	if err == nil {
		t.Error("expected error for param containing backtick")
	}

	// 正常参数应通过（在 Windows 上 echo 可能不可用，但参数校验应先通过）
	// 这里只校验参数校验逻辑，不实际执行
	err = ValidateInput("hello world")
	if err != nil {
		t.Errorf("normal input should pass validation: %v", err)
	}
}

// TestCommandExecutor_Timeout 测试命令超时
func TestCommandExecutor_Timeout(t *testing.T) {
	exec := NewCommandExecutor(1 * time.Second)

	exec.Register(&CommandTemplate{
		Name: "sleep_test",
		Bin:  "sleep",
		Args: []string{"10"},
	})

	start := time.Now()
	_, err := exec.Execute(context.Background(), "sleep_test", map[string]string{})
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected timeout error, got nil")
	}
	if elapsed > 3*time.Second {
		t.Errorf("timeout took too long: %v", elapsed)
	}
}

// TestCommandExecutor_MissingParam 测试缺少必需参数
func TestCommandExecutor_MissingParam(t *testing.T) {
	exec := NewCommandExecutor(5 * time.Second)

	exec.Register(&CommandTemplate{
		Name: "template_with_param",
		Bin:  "echo",
		Args: []string{"{{required_param}}"},
	})

	_, err := exec.Execute(context.Background(), "template_with_param", map[string]string{})
	if err == nil {
		t.Error("expected error for missing required param")
	}
}

// TestValidateInput 测试 ValidateInput 函数
func TestValidateInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"normal", "hello", false},
		{"with space", "hello world", false},
		{"with semicolon", "hello;", true},
		{"with pipe", "hello|world", true},
		{"with dollar", "$HOME", true},
		{"with backtick", "`cmd`", true},
		{"with parens", "(test)", true},
		{"with redirect", "file>out", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
