package common

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// 默认超时
const DefaultCommandTimeout = 30 * time.Second

// 禁止的 shell 元字符
var shellMetaRegex = regexp.MustCompile(`[;&|$` + "`" + `\\(){}[\]<>!#~*?]`)

// CommandTemplate 预定义命令模板
type CommandTemplate struct {
	Name        string   // 模板名称
	Description string   // 描述
	Bin         string   // 可执行文件路径
	Args        []string // 固定参数（不含用户输入）
	// 用户可填参数的位置索引，值为参数占位符名称
	// 例如 Args: ["-name", "{{name}}"] 中 {{name}} 可被替换
	AllowedArgs map[string]bool // 允许的用户参数名
	// Env 执行命令时设置的环境变量（用于传递敏感信息如密码，避免出现在进程参数中）
	Env map[string]string
}

// CommandResult 命令执行结果
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// CommandExecutor 安全命令执行器
type CommandExecutor struct {
	templates map[string]*CommandTemplate
	timeout   time.Duration
}

// NewCommandExecutor 创建命令执行器
func NewCommandExecutor(timeout time.Duration) *CommandExecutor {
	if timeout <= 0 {
		timeout = DefaultCommandTimeout
	}
	return &CommandExecutor{
		templates: make(map[string]*CommandTemplate),
		timeout:   timeout,
	}
}

// Register 注册命令模板
func (e *CommandExecutor) Register(t *CommandTemplate) {
	e.templates[t.Name] = t
}

// Execute 执行预定义命令，params 为用户输入参数（仅替换模板中的占位符）
func (e *CommandExecutor) Execute(ctx context.Context, templateName string, params map[string]string) (*CommandResult, error) {
	return e.ExecuteWithEnv(ctx, templateName, params, nil)
}

// ExecuteWithEnv 执行预定义命令，支持额外环境变量（用于传递敏感信息如密码）
func (e *CommandExecutor) ExecuteWithEnv(ctx context.Context, templateName string, params map[string]string, extraEnv map[string]string) (*CommandResult, error) {
	tpl, ok := e.templates[templateName]
	if !ok {
		return nil, fmt.Errorf("command template not registered: %s", templateName)
	}

	// 构建参数：替换占位符，校验用户输入
	args := make([]string, 0, len(tpl.Args))
	for _, arg := range tpl.Args {
		if strings.HasPrefix(arg, "{{") && strings.HasSuffix(arg, "}}") {
			key := strings.TrimSuffix(strings.TrimPrefix(arg, "{{"), "}}")
			val, exists := params[key]
			if !exists {
				return nil, fmt.Errorf("missing required param: %s", key)
			}
			// 安全校验：禁止 shell 元字符
			if shellMetaRegex.MatchString(val) {
				return nil, fmt.Errorf("param %s contains forbidden characters", key)
			}
			args = append(args, val)
		} else {
			args = append(args, arg)
		}
	}

	// 执行命令（不经过 shell，直接 exec，避免注入）
	// 无论调用方是否传入 ctx，都叠加 executor 自身的超时上限，
	// 保证 NewCommandExecutor(timeout) 设置的超时始终生效。
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, tpl.Bin, args...)
	// 设置环境变量（模板静态 + 调用方动态，用于传递敏感信息）
	env := os.Environ()
	for k, v := range tpl.Env {
		env = append(env, k+"="+v)
	}
	for k, v := range extraEnv {
		env = append(env, k+"="+v)
	}
	if len(env) > len(os.Environ()) {
		cmd.Env = env
	}
	start := time.Now()

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	exitCode := 0
	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("command timed out after %s: %w", e.timeout, ctx.Err())
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("execute command: %w", err)
		}
	}

	return &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Duration: time.Since(start),
	}, nil
}

// ValidateInput 校验单个用户输入是否安全
func ValidateInput(input string) error {
	if shellMetaRegex.MatchString(input) {
		return fmt.Errorf("input contains forbidden shell characters")
	}
	return nil
}
