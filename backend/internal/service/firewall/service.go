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
	// --- 端口转发 NAT 规则（S-004，iptables -t nat） ---
	// PREROUTING DNAT：外部访问本机端口 → 目标地址
	e.Register(&common.CommandTemplate{
		Name: "iptables_add_forward",
		Bin:  "iptables",
		Args: []string{"-t", "nat", "-A", "PREROUTING", "-p", "{{protocol}}", "--dport", "{{port}}", "-j", "DNAT", "--to-destination", "{{dest}}"},
	})
	e.Register(&common.CommandTemplate{
		Name: "iptables_del_forward",
		Bin:  "iptables",
		Args: []string{"-t", "nat", "-D", "PREROUTING", "-p", "{{protocol}}", "--dport", "{{port}}", "-j", "DNAT", "--to-destination", "{{dest}}"},
	})
	// OUTPUT DNAT：本机自身访问的转发
	e.Register(&common.CommandTemplate{
		Name: "iptables_add_forward_local",
		Bin:  "iptables",
		Args: []string{"-t", "nat", "-A", "OUTPUT", "-p", "{{protocol}}", "--dport", "{{port}}", "-j", "DNAT", "--to-destination", "{{dest}}"},
	})
	e.Register(&common.CommandTemplate{
		Name: "iptables_del_forward_local",
		Bin:  "iptables",
		Args: []string{"-t", "nat", "-D", "OUTPUT", "-p", "{{protocol}}", "--dport", "{{port}}", "-j", "DNAT", "--to-destination", "{{dest}}"},
	})
	// POSTROUTING SNAT/MASQUERADE：保证回包可达
	e.Register(&common.CommandTemplate{
		Name: "iptables_add_forward_masq",
		Bin:  "iptables",
		Args: []string{"-t", "nat", "-A", "POSTROUTING", "-p", "{{protocol}}", "-d", "{{dest_ip}}", "--dport", "{{dest_port}}", "-j", "MASQUERADE"},
	})
	e.Register(&common.CommandTemplate{
		Name: "iptables_del_forward_masq",
		Bin:  "iptables",
		Args: []string{"-t", "nat", "-D", "POSTROUTING", "-p", "{{protocol}}", "-d", "{{dest_ip}}", "--dport", "{{dest_port}}", "-j", "MASQUERADE"},
	})
	// FORWARD 链放行转发流量
	e.Register(&common.CommandTemplate{
		Name: "iptables_add_forward_accept",
		Bin:  "iptables",
		Args: []string{"-A", "FORWARD", "-p", "{{protocol}}", "-d", "{{dest_ip}}", "--dport", "{{dest_port}}", "-j", "ACCEPT"},
	})
	e.Register(&common.CommandTemplate{
		Name: "iptables_del_forward_accept",
		Bin:  "iptables",
		Args: []string{"-D", "FORWARD", "-p", "{{protocol}}", "-d", "{{dest_ip}}", "--dport", "{{dest_port}}", "-j", "ACCEPT"},
	})
	// 开启内核转发
	e.Register(&common.CommandTemplate{
		Name: "sysctl_ip_forward",
		Bin:  "sysctl",
		Args: []string{"-w", "net.ipv4.ip_forward=1"},
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
	// 校验：port 类型必须有端口，ip 类型必须有 IP，forward 类型必须有目标地址
	if req.Type == "port" && req.Port <= 0 {
		return nil, fmt.Errorf("端口类型规则必须指定端口号")
	}
	if req.Type == "ip" && req.IPAddress == "" {
		return nil, fmt.Errorf("IP 类型规则必须指定 IP 地址")
	}
	if req.Type == "forward" {
		if req.Port <= 0 {
			return nil, fmt.Errorf("端口转发规则必须指定源端口号")
		}
		if req.DestIP == "" {
			return nil, fmt.Errorf("端口转发规则必须指定目标地址")
		}
		if req.DestPort <= 0 {
			req.DestPort = req.Port // 未指定目标端口则保持一致
		}
		if req.Protocol != "tcp" && req.Protocol != "udp" {
			return nil, fmt.Errorf("端口转发仅支持 tcp/udp 协议")
		}
	}

	rule := &model.FirewallRule{
		Type:      req.Type,
		Direction: req.Direction,
		Protocol:  req.Protocol,
		Port:      req.Port,
		IPAddress: req.IPAddress,
		DestIP:    req.DestIP,
		DestPort:  req.DestPort,
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
	} else if rule.Type == "forward" {
		err = s.applyForwardRule(ctx, rule, add)
	}

	return err
}

// applyForwardRule 应用端口转发 NAT 规则（DNAT + MASQUERADE + FORWARD 放行）
func (s *Service) applyForwardRule(ctx context.Context, rule *model.FirewallRule, add bool) error {
	dest := fmt.Sprintf("%s:%d", rule.DestIP, rule.DestPort)
	params := map[string]string{
		"protocol":  rule.Protocol,
		"port":      strconv.Itoa(rule.Port),
		"dest":      dest,
		"dest_ip":   rule.DestIP,
		"dest_port": strconv.Itoa(rule.DestPort),
	}

	verb := "add"
	if !add {
		verb = "del"
	}

	// 1. 确保内核转发已开启（添加时）
	if add {
		s.executor.Execute(ctx, "sysctl_ip_forward", nil)
	}

	// 2. DNAT 规则（PREROUTING 供外部访问；本机访问走 OUTPUT）
	_, err := s.executor.Execute(ctx, "iptables_"+verb+"_forward", params)
	if err != nil {
		return err
	}
	s.executor.Execute(ctx, "iptables_"+verb+"_forward_local", params)
	// 3. MASQUERADE + FORWARD 放行（删除时同步清理）
	s.executor.Execute(ctx, "iptables_"+verb+"_forward_masq", params)
	s.executor.Execute(ctx, "iptables_"+verb+"_forward_accept", params)
	return nil
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
		DestIP:    rule.DestIP,
		DestPort:  rule.DestPort,
		Action:    rule.Action,
		Status:    rule.Status,
		Remark:    rule.Remark,
		CreatedAt: rule.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
