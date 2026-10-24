package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Service AI 运维助手（roadmap #14）：日志诊断/故障分析（OpenAI 兼容接口）
type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

type DiagnoseRequest struct {
	Type     string `json:"type"`     // logs | metrics | config
	Content  string `json:"content"`  // 待分析的原始内容（日志片段/指标快照）
	Question string `json:"question"` // 用户附加问题（可选）
}

type DiagnoseResult struct {
	Analysis string `json:"analysis"`
	Model    string `json:"model"`
	Latency  int64  `json:"latency_ms"`
}

func (s *Service) getSettings() (apiURL, apiKey, model string, err error) {
	rows, err2 := s.db.Table("settings").Where("`key` IN ?", []string{"ai_api_url", "ai_api_key", "ai_model"}).Rows()
	if err2 != nil {
		return "", "", "", err2
	}
	defer rows.Close()
	vals := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err == nil {
			vals[k] = v
		}
	}
	apiURL = vals["ai_api_url"]
	apiKey = vals["ai_api_key"]
	model = vals["ai_model"]
	if apiURL == "" || model == "" {
		err = fmt.Errorf("AI 助手未配置：请先在系统设置中填写 AI 接口地址与模型名（OpenAI 兼容接口）")
	}
	return
}

const diagnoseSystemPrompt = `你是一名资深 Linux 服务器运维专家。用户是服务器管理面板的使用者，会提供日志片段、监控指标或配置内容。
请直接给出：
1. 问题诊断（发现什么异常、根本原因推测）
2. 风险等级（高/中/低/无异常）
3. 具体修复步骤（可执行的命令或操作）
用简体中文回答，简洁、结论先行，不要客套。`

// Diagnose 调用 LLM 分析内容
func (s *Service) Diagnose(req *DiagnoseRequest) (*DiagnoseResult, error) {
	apiURL, apiKey, model, err := s.getSettings()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Content) == "" {
		return nil, fmt.Errorf("分析内容不能为空")
	}
	// 内容截断（防 token 爆炸）
	content := req.Content
	if len(content) > 32000 {
		content = content[:32000] + "\n...[内容过长已截断]"
	}
	userMsg := fmt.Sprintf("类型: %s\n%s\n\n待分析内容:\n%s", req.Type, req.Question, content)

	start := time.Now()
	body, _ := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": diagnoseSystemPrompt},
			{"role": "user", "content": userMsg},
		},
		"temperature": 0.2,
	})
	url := strings.TrimRight(apiURL, "/")
	if !strings.HasSuffix(url, "/chat/completions") {
		url += "/chat/completions"
	}
	client := &http.Client{Timeout: 120 * time.Second}
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("AI 接口请求失败: %w", err)
	}
	defer resp.Body.Close()

	var lr struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, fmt.Errorf("AI 响应解析失败(HTTP %d): %w", resp.StatusCode, err)
	}
	if lr.Error != nil {
		return nil, fmt.Errorf("AI 接口错误: %s", lr.Error.Message)
	}
	if len(lr.Choices) == 0 {
		return nil, fmt.Errorf("AI 未返回分析结果")
	}
	return &DiagnoseResult{
		Analysis: lr.Choices[0].Message.Content,
		Model:    model,
		Latency:  time.Since(start).Milliseconds(),
	}, nil
}
