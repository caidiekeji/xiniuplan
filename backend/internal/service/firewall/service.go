package firewall

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// Service 防火墙服务
type Service struct {
	db       *gorm.DB
	executor *common.CommandExecutor
}

// NewService 创建防火墙服务
func NewService(db *gorm.DB) *Service {
	executor := common.NewCommandExecutor(30 * time.Second)
	registerCommandTemplates(executor)
	return &Service{
		db:       db,
		executor: executor,
	}
}

// registerCommandTemplates 注册系统命令白名单模板
func registerCommandTemplates(e *common.CommandExecutor) {
	// iptables 列出规则
	e.Register(&common.CommandTemplate{
		Name: "iptables_list",
		Bin:  "iptables",
		Args: []string{"-L", "-n"},
	})
	// iptables 添加端口规则
	e.Register(&common.CommandTemplate{
		Name: "iptables_add_port",
		Bin:  "iptables",
		Args: []string{"-A", "{{chain}}", "-p", "{{protocol}}", "--dport", "{{port}}", "-j", "{{target}}"},
	})
	// iptables 删除端口规则
	e.Register(&common.CommandTemplate{
		Name: "iptables_del_port",
		Bin:  "iptables",
		Args: []string{"-D", "{{chain}}", "-p", "{{protocol}}", "--dport", "{{port}}", "-j", "{{target}}"},
	})
	// iptables 添加 IP 规则
	e.Register(&common.CommandTemplate{
		Name: "iptables_add_ip",
		Bin:  "iptables",
		Args: []string{"-A", "{{chain}}", "-s", "{{ip}}", "-j", "{{target}}"},
	})
	// iptables 删除 IP 规则
	e.Register(&common.CommandTemplate{
		Name: "iptables_del_ip",
		Bin:  "iptables",
		Args: []string{"-D", "{{chain}}", "-s", "{{ip}}", "-j", "{{target}}"},
	})
	// firewalld 列出所有
	e.Register(&common.CommandTemplate{
		Name: "firewalld_list_all",
		Bin:  "firewall-cmd",
		Args: []string{"--list-all"},
	})
	// firewalld 添加端口
	e.Register(&common.CommandTemplate{
		Name: "firewalld_add_port",
		Bin:  "firewall-cmd",
		Args: []string{"--add-port={{port}}/{{protocol}}", "--permanent"},
	})
	// firewalld 删除端口
	e.Register(&common.CommandTemplate{
		Name: "firewalld_remove_port",
		Bin:  "firewall-cmd",
		Args: []string{"--remove-port={{port}}/{{protocol}}", "--permanent"},
	})
}

// ListRules 分页查询规则
func (s *Service) ListRules(page, pageSize int, ruleType string) ([]model.FirewallRule, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := s.db.Model(&model.FirewallRule{})
	if ruleType != "" {
		query = query.Where("type = ?", ruleType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count rules: %w", err)
	}

	rules := make([]model.FirewallRule, 0)
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&rules).Error; err != nil {
		return nil, 0, fmt.Errorf("query rules: %w", err)
	}

	return rules, total, nil
}

// AddRule 创建规则并执行系统命令
func (s *Service) AddRule(req dto.CreateRuleRequest) (*model.FirewallRule, error) {
	// 校验：port 类型必须有端口，ip 类型必须有 IP
	if req.Type == "port" && req.Port <= 0 {
		return nil, fmt.Errorf("端口类型规则必须指定端口号")
	}
	if req.Type == "ip" && req.IPAddress == "" {
		return nil, fmt.Errorf("IP 类型规则必须指定 IP 地址")
	}

	rule := &model.FirewallRule{
		Type:      req.Type,
		Direction: req.Direction,
		Protocol:  req.Protocol,
		Port:      req.Port,
		IPAddress: req.IPAddress,
		Action:    req.Action,
		Status:    "enabled",
		Remark:    req.Remark,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(rule).Error; err != nil {
		return nil, fmt.Errorf("create rule: %w", err)
	}

	// 执行系统命令（失败不影响数据库记录，返回警告）
	if err := s.applyRule(context.Background(), rule, true); err != nil {
		// 命令执行失败（如 Windows 环境无 iptables），记录但不回滚
		fmt.Printf("[firewall] 应用规则命令失败（规则已保存）: %v\n", err)
	}

	return rule, nil
}

// DeleteRule 删除规则并执行删除命令
func (s *Service) DeleteRule(id uint) error {
	var rule model.FirewallRule
	if err := s.db.First(&rule, id).Error; err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}

	// 执行删除命令
	if rule.Status == "enabled" {
		if err := s.applyRule(context.Background(), &rule, false); err != nil {
			fmt.Printf("[firewall] 删除规则命令失败: %v\n", err)
		}
	}

	if err := s.db.Delete(&rule).Error; err != nil {
		return fmt.Errorf("delete rule: %w", err)
	}
	return nil
}

// ToggleRule 启用/禁用规则
func (s *Service) ToggleRule(id uint, enabled bool) (*model.FirewallRule, error) {
	var rule model.FirewallRule
	if err := s.db.First(&rule, id).Error; err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
	}

	newStatus := "disabled"
	if enabled {
		newStatus = "enabled"
	}

	rule.Status = newStatus
	if err := s.db.Save(&rule).Error; err != nil {
		return nil, fmt.Errorf("update rule status: %w", err)
	}

	// 执行对应命令
	if err := s.applyRule(context.Background(), &rule, enabled); err != nil {
		fmt.Printf("[firewall] 切换规则命令失败: %v\n", err)
	}

	return &rule, nil
}

// applyRule 执行 iptables/firewalld 命令
// add=true 表示添加规则，add=false 表示删除规则
func (s *Service) applyRule(ctx context.Context, rule *model.FirewallRule, add bool) error {
	chain := "INPUT"
	if rule.Direction == "outbound" {
		chain = "OUTPUT"
	}

	target := "ACCEPT"
	if rule.Action == "drop" {
		target = "DROP"
	}

	protocol := rule.Protocol
	if protocol == "all" {
		protocol = "all"
	}

	var err error
	if rule.Type == "port" {
		portStr := strconv.Itoa(rule.Port)
		if add {
			_, err = s.executor.Execute(ctx, "iptables_add_port", map[string]string{
				"chain":    chain,
				"protocol": protocol,
				"port":     portStr,
				"target":   target,
			})
		} else {
			_, err = s.executor.Execute(ctx, "iptables_del_port", map[string]string{
				"chain":    chain,
				"protocol": protocol,
				"port":     portStr,
				"target":   target,
			})
		}
	} else if rule.Type == "ip" {
		if add {
			_, err = s.executor.Execute(ctx, "iptables_add_ip", map[string]string{
				"chain":  chain,
				"ip":     rule.IPAddress,
				"target": target,
			})
		} else {
			_, err = s.executor.Execute(ctx, "iptables_del_ip", map[string]string{
				"chain":  chain,
				"ip":     rule.IPAddress,
				"target": target,
			})
		}
	}

	return err
}

// GetSecuritySuggestions 检测高危端口是否对外开放，返回建议列表
func (s *Service) GetSecuritySuggestions() ([]dto.SecuritySuggestion, error) {
	// 高危端口定义
	highRiskPorts := []struct {
		Port    int
		Service string
		Level   string
	}{
		{3306, "MySQL", "high"},
		{6379, "Redis", "high"},
		{27017, "MongoDB", "high"},
		{21, "FTP", "high"},
		{23, "Telnet", "high"},
		{1433, "MSSQL", "high"},
		{5432, "PostgreSQL", "medium"},
		{11211, "Memcached", "medium"},
		{9200, "Elasticsearch", "medium"},
	}

	suggestions := make([]dto.SecuritySuggestion, 0)

	// 查询数据库中已存在的 accept 规则，判断高危端口是否被放行
	for _, hp := range highRiskPorts {
		var count int64
		s.db.Model(&model.FirewallRule{}).
			Where("type = ? AND port = ? AND action = ? AND status = ?", "port", hp.Port, "accept", "enabled").
			Count(&count)

		if count > 0 {
			suggestion := dto.SecuritySuggestion{
				Port:       hp.Port,
				Service:    hp.Service,
				RiskLevel:  hp.Level,
				Suggestion: fmt.Sprintf("%s 端口 %d 已对外开放，建议限制来源 IP 或关闭非必要访问", hp.Service, hp.Port),
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	// 如果没有检测到任何规则，给出通用建议
	if len(suggestions) == 0 {
		suggestions = append(suggestions, dto.SecuritySuggestion{
			Port:       0,
			Service:    "system",
			RiskLevel:  "low",
			Suggestion: "未检测到高危端口放行规则，系统防火墙配置较为安全",
		})
	}

	return suggestions, nil
}

// ToResponse 转换模型为响应 DTO
func ToResponse(rule *model.FirewallRule) dto.RuleResponse {
	return dto.RuleResponse{
		ID:        rule.ID,
		Type:      rule.Type,
		Direction: rule.Direction,
		Protocol:  rule.Protocol,
		Port:      rule.Port,
		IPAddress: rule.IPAddress,
		Action:    rule.Action,
		Status:    rule.Status,
		Remark:    rule.Remark,
		CreatedAt: rule.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
