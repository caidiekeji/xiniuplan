# 防火墙 API

防火墙模块提供服务器端口规则管理和安全建议功能。

## 规则列表

`GET /api/v1/firewall/rules`

获取防火墙规则列表（分页）。

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| pageSize | int | 10 | 每页条数 |
| type | string | - | 按类型筛选：inbound / outbound |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "type": "inbound",
        "protocol": "tcp",
        "port": "443",
        "source": "",
        "action": "accept",
        "enabled": true,
        "description": "HTTPS 访问",
        "created_at": "2026-01-01T00:00:00Z"
      },
      {
        "id": 2,
        "type": "inbound",
        "protocol": "tcp",
        "port": "22",
        "source": "192.168.1.0/24",
        "action": "accept",
        "enabled": true,
        "description": "SSH（仅内网）",
        "created_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 2,
    "page": 1,
    "pageSize": 10
  }
}
```

### 规则对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 规则 ID |
| type | string | 类型：inbound / outbound |
| protocol | string | 协议：tcp / udp / icmp / all |
| port | string | 端口号或范围（如 "80", "3000-4000"） |
| source | string | 源 IP/CIDR（空表示任意） |
| action | string | 动作：accept / drop / reject |
| enabled | bool | 是否启用 |
| description | string | 规则描述 |
| created_at | string | 创建时间 |

## 添加规则

`POST /api/v1/firewall/rules`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | 类型：inbound / outbound |
| protocol | string | 是 | 协议：tcp / udp / icmp / all |
| port | string | 是 | 端口号或范围 |
| source | string | 否 | 源 IP/CIDR（默认空=任意） |
| action | string | 是 | 动作：accept / drop / reject |
| description | string | 否 | 规则描述 |

### 请求示例

```json
{
  "type": "inbound",
  "protocol": "tcp",
  "port": "8080",
  "source": "",
  "action": "accept",
  "description": "应用端口"
}
```

### 响应示例

返回创建的规则对象。

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 参数错误 |
| 50001 | 添加规则失败 |

## 删除规则

`DELETE /api/v1/firewall/rules/:id`

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "deleted": true
  }
}
```

## 切换规则状态

`POST /api/v1/firewall/rules/:id/toggle`

启用或禁用规则。

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| enabled | bool | 是 | 目标状态 |

### 请求示例

```json
{
  "enabled": false
}
```

### 响应示例

返回更新后的规则对象。

## 安全建议

`GET /api/v1/firewall/suggestions`

自动检查服务器安全配置并给出加固建议。

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": "ssh_default_port",
      "title": "SSH 使用默认端口",
      "level": "warning",
      "description": "SSH 服务使用默认 22 端口，容易受到自动化扫描和暴力破解攻击",
      "suggestion": "建议修改为非标准端口（如 2222），并在防火墙中开放新端口",
      "status": "detected"
    },
    {
      "id": "password_login",
      "title": "SSH 密码登录已启用",
      "level": "critical",
      "description": "SSH 允许密码登录，存在被暴力破解的风险",
      "suggestion": "建议配置 SSH 密钥登录后禁用密码登录",
      "status": "detected"
    }
  ]
}
```

### 建议对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 建议项唯一标识 |
| title | string | 建议标题 |
| level | string | 级别：info / warning / critical |
| description | string | 问题描述 |
| suggestion | string | 修复建议 |
| status | string | 状态：detected / passed |

### 检查项

安全建议检查包括：

- SSH 是否使用默认 22 端口
- SSH 是否允许密码登录
- SSH 是否允许 root 直接登录
- 是否有不必要的端口开放
- 防火墙服务是否启用
- 面板端口是否对外开放
- Fail2ban 是否启用

## 注意事项

1. **操作风险**：错误的防火墙规则可能导致无法远程连接服务器，操作前确保有备用访问方式
2. **系统依赖**：防火墙功能依赖系统防火墙服务（iptables/firewalld/ufw），需确保服务正常运行
3. **规则生效**：添加/删除/切换规则后立即生效，无需重启服务
4. **规则顺序**：防火墙规则按顺序匹配，注意规则的先后顺序
5. **面板端口**：切勿封禁面板端口（默认 8888）和 SSH 端口，否则可能无法访问面板
