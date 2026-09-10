# SSH API

SSH 安全模块提供 SSH 配置管理、密钥管理、登录日志和 Fail2ban 防护功能。

## SSH 配置

### 获取 SSH 配置

`GET /api/v1/ssh/config`

读取当前 SSH 服务配置。

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "port": 22,
    "permit_root_login": true,
    "password_authentication": true,
    "pubkey_authentication": true,
    "max_auth_tries": 6,
    "login_grace_time": 120,
    "x11_forwarding": false,
    "allow_tcp_forwarding": true,
    "use_dns": false
  }
}
```

### 配置字段

| 字段 | 类型 | 说明 |
|------|------|------|
| port | int | SSH 监听端口 |
| permit_root_login | bool | 是否允许 root 登录 |
| password_authentication | bool | 是否允许密码认证 |
| pubkey_authentication | bool | 是否启用公钥认证 |
| max_auth_tries | int | 最大认证尝试次数 |
| login_grace_time | int | 登录超时时间（秒） |
| x11_forwarding | bool | 是否允许 X11 转发 |
| allow_tcp_forwarding | bool | 是否允许 TCP 转发 |
| use_dns | bool | 是否进行 DNS 反向解析 |

### 修改 SSH 配置

`PUT /api/v1/ssh/config`

### 请求体

字段同获取配置，所有字段可选（只传需要修改的字段）。

### 参数校验

- `port`：1-65535
- `max_auth_tries`：1-20

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "updated": true
  }
}
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 参数错误（端口超出范围、最大尝试次数超出范围等） |
| 50001 | 修改配置失败 |

> **警告**：修改 SSH 端口或禁用密码登录前，请确保已配置密钥登录并测试正常，且新端口已在防火墙开放。

## SSH 密钥管理

### 密钥列表

`GET /api/v1/ssh/keys`

列出 `~/.ssh/authorized_keys` 中的所有公钥。

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "fingerprint": "SHA256:abc123def456...",
      "type": "ssh-ed25519",
      "comment": "admin@example.com",
      "added_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

### 密钥对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| fingerprint | string | 公钥指纹（SHA256） |
| type | string | 密钥类型：ssh-rsa / ssh-ed25519 / ecdsa-sha2-nistp256 等 |
| comment | string | 公钥注释（通常是邮箱或主机名） |
| added_at | string | 添加时间 |

### 添加公钥

`POST /api/v1/ssh/keys`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| public_key | string | 是 | 公钥内容（完整的一行） |

### 请求示例

```json
{
  "public_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIB... admin@example.com"
}
```

### 响应示例

返回添加的密钥对象。

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 参数错误（公钥格式无效） |
| 50001 | 添加公钥失败 |

### 删除公钥

`DELETE /api/v1/ssh/keys/:fingerprint`

### 路径参数

| 参数 | 说明 |
|------|------|
| fingerprint | 公钥指纹（URL 编码） |

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

> 删除前请确认还有其他可用的登录方式。

## SSH 登录日志

`GET /api/v1/ssh/logs`

获取 SSH 登录日志（分页）。

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| pageSize | int | 20 | 每页条数 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "time": "2026-09-09T10:30:00Z",
        "username": "root",
        "ip": "192.168.1.100",
        "port": 54321,
        "status": "success",
        "auth_method": "publickey"
      },
      {
        "time": "2026-09-09T10:25:00Z",
        "username": "admin",
        "ip": "203.0.113.50",
        "port": 12345,
        "status": "failed",
        "auth_method": "password"
      }
    ],
    "total": 100,
    "page": 1,
    "pageSize": 20
  }
}
```

### 日志字段

| 字段 | 类型 | 说明 |
|------|------|------|
| time | string | 登录时间 |
| username | string | 用户名 |
| ip | string | 来源 IP |
| port | int | 来源端口 |
| status | string | 结果：success / failed |
| auth_method | string | 认证方式：password / publickey |

## Fail2ban 防护

### 获取 Fail2ban 状态

`GET /api/v1/ssh/fail2ban/status`

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "service_status": "running",
    "jail": "sshd",
    "currently_banned": 2,
    "total_banned": 15,
    "banned_ips": [
      {
        "ip": "203.0.113.50",
        "banned_at": "2026-09-09T10:00:00Z",
        "reason": "Too many authentication failures"
      },
      {
        "ip": "198.51.100.20",
        "banned_at": "2026-09-09T09:30:00Z",
        "reason": "Too many authentication failures"
      }
    ]
  }
}
```

### 修改 Fail2ban 配置

`PUT /api/v1/ssh/fail2ban/config`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| max_retry | int | 是 | 最大重试次数（1-20） |
| find_time | int | 是 | 检测时间窗口秒（60-86400） |
| ban_time | int | 是 | 封禁时长秒（60-31536000） |

### 请求示例

```json
{
  "max_retry": 5,
  "find_time": 600,
  "ban_time": 3600
}
```

### 参数校验

- `max_retry`：1-20
- `find_time`：60-86400（1 分钟到 1 天）
- `ban_time`：60-31536000（1 分钟到 1 年）

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "updated": true
  }
}
```

### 解封 IP

`POST /api/v1/ssh/fail2ban/unban`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| ip | string | 是 | 要解封的 IP 地址 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "unbanned": "203.0.113.50"
  }
}
```

### 手动封禁 IP

`POST /api/v1/ssh/fail2ban/ban`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| ip | string | 是 | 要封禁的 IP 地址 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "banned": "203.0.113.50"
  }
}
```

## 注意事项

1. **配置生效**：修改 SSH 配置后会自动重启 SSH 服务，现有连接可能断开
2. **密钥安全**：公钥添加后立即生效，删除公钥后该密钥无法再登录
3. **Fail2ban 依赖**：Fail2ban 功能需要系统已安装 fail2ban 服务
4. **日志来源**：登录日志从系统日志（/var/log/auth.log 或 /var/log/secure）解析
5. **操作风险**：禁用密码登录或删除所有密钥前，务必确保有其他访问方式
