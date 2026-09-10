# 系统设置 API

系统设置模块提供面板全局配置管理、SMTP 邮件配置、版本信息、系统清理和详细系统信息功能。

## 系统设置

### 获取设置

`GET /api/v1/system/settings`

获取面板所有系统配置。

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "settings": {
      "site_name": "云枢面板",
      "site_url": "http://localhost:8888",
      "language": "zh-CN",
      "timezone": "Asia/Shanghai",
      "session_timeout": "120",
      "login_max_retries": "5",
      "password_min_length": "8",
      "monitor_interval": "60",
      "data_retention_days": "30",
      "backup_path": "/data/backups",
      "backup_retention_days": "30",
      "alert_enabled": "true"
    }
  }
}
```

### 设置项说明

| 配置类别 | 配置项 | 说明 |
|----------|--------|------|
| 基本设置 | site_name | 面板名称 |
| 基本设置 | site_url | 面板访问地址 |
| 基本设置 | language | 默认语言 |
| 基本设置 | timezone | 系统时区 |
| 安全设置 | session_timeout | 会话超时时间（分钟） |
| 安全设置 | login_max_retries | 最大登录重试次数 |
| 安全设置 | password_min_length | 密码最小长度 |
| 监控设置 | monitor_interval | 监控采集间隔（秒） |
| 监控设置 | data_retention_days | 监控数据保留天数 |
| 备份设置 | backup_path | 备份文件存储路径 |
| 备份设置 | backup_retention_days | 备份保留天数 |
| 通知设置 | alert_enabled | 是否启用告警通知 |

### 更新设置

`PUT /api/v1/system/settings`

需要 `setting:update` 权限（仅超级管理员）。

### 请求体

```json
{
  "settings": {
    "site_name": "我的面板",
    "session_timeout": "60",
    "monitor_interval": "30"
  }
}
```

`settings` 为键值对对象，只传需要修改的配置项。

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

## SMTP 邮件配置

### 获取 SMTP 配置

`GET /api/v1/system/smtp`

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "host": "smtp.example.com",
    "port": "465",
    "username": "alert@example.com",
    "password": "********",
    "from_address": "alert@example.com",
    "from_name": "云枢面板告警",
    "use_ssl": "true"
  }
}
```

> 密码字段返回脱敏值（`********`），不会返回明文。

### 更新 SMTP 配置

`PUT /api/v1/system/smtp`

### 请求体

请求体为键值对 map（`map[string]string`），包含以下字段：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| host | string | 是 | SMTP 服务器地址 |
| port | string | 是 | 端口（25/465/587） |
| username | string | 是 | SMTP 用户名 |
| password | string | 是 | SMTP 密码/授权码 |
| from_address | string | 是 | 发件人邮箱 |
| from_name | string | 否 | 发件人名称 |
| use_ssl | string | 否 | 是否使用 SSL（"true"/"false"） |

### 请求示例

```json
{
  "host": "smtp.qq.com",
  "port": "465",
  "username": "alert@example.com",
  "password": "your_auth_code",
  "from_address": "alert@example.com",
  "from_name": "云枢面板",
  "use_ssl": "true"
}
```

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

### 发送测试邮件

`POST /api/v1/system/smtp/test`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| to | string | 是 | 收件人邮箱 |

### 请求示例

```json
{
  "to": "admin@example.com"
}
```

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "sent": true
  }
}
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 参数错误（收件人缺失） |
| 50001 | 发送失败（SMTP 配置错误、网络不通等） |

## 版本信息

### 获取当前版本

`GET /api/v1/system/version`

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "version": "1.0.0",
    "build_time": "2026-09-01T00:00:00Z",
    "git_commit": "abc123def456",
    "go_version": "go1.22.0"
  }
}
```

### 检查更新

`POST /api/v1/system/version/check`

连接更新服务器检查最新版本。

### 响应示例（有新版本）

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "current_version": "1.0.0",
    "latest_version": "1.1.0",
    "has_update": true,
    "release_notes": "1. 新增应用商店功能\n2. 优化监控性能\n3. 修复已知问题",
    "download_url": "https://example.com/download/cloudpivot-1.1.0.tar.gz"
  }
}
```

### 响应示例（已是最新）

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "current_version": "1.0.0",
    "latest_version": "1.0.0",
    "has_update": false
  }
}
```

## 系统清理

`POST /api/v1/system/clean`

清理系统临时文件和缓存。

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| clean_type | string | 是 | 清理类型：temp / logs / docker / cache / old_backups |

### 请求示例

```json
{
  "clean_type": "logs"
}
```

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "clean_type": "logs",
    "result": {
      "files_cleaned": 15,
      "space_freed": 104857600
    }
  }
}
```

### 清理类型说明

| 类型 | 标识 | 清理内容 |
|------|------|----------|
| 临时文件 | temp | 系统临时目录过期文件 |
| 日志文件 | logs | 系统和应用日志轮转旧文件 |
| Docker 缓存 | docker | 未使用的镜像、容器、构建缓存 |
| 面板缓存 | cache | 面板运行临时缓存 |
| 旧备份 | old_backups | 超过保留期限的备份文件 |

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 参数错误（clean_type 缺失或无效） |
| 50001 | 清理失败 |

## 系统信息

`GET /api/v1/system/info`

获取服务器详细系统信息。

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "hostname": "cloudpivot-server",
    "os": "Ubuntu 22.04.3 LTS",
    "kernel": "5.15.0-91-generic",
    "arch": "x86_64",
    "uptime": 864000,
    "cpu": {
      "model": "Intel(R) Xeon(R) CPU E5-2680 v4 @ 2.40GHz",
      "cores": 4,
      "threads": 8
    },
    "memory": {
      "total": 8192,
      "used": 3200,
      "free": 4992
    },
    "disk": {
      "total": 102400,
      "used": 45000,
      "free": 57400,
      "mount_point": "/"
    },
    "load": {
      "load1": 0.45,
      "load5": 0.32,
      "load15": 0.28
    },
    "runtime": {
      "panel_uptime": 3600,
      "go_version": "go1.22.0",
      "goroutines": 50,
      "memory_alloc": 52428800,
      "db_path": "./cloudpivot.db",
      "db_size": 10485760
    }
  }
}
```

### 响应字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| hostname | string | 主机名 |
| os | string | 操作系统名称和版本 |
| kernel | string | 内核版本 |
| arch | string | 系统架构 |
| uptime | int | 系统运行时间（秒） |
| cpu.model | string | CPU 型号 |
| cpu.cores | int | 物理核心数 |
| cpu.threads | int | 逻辑线程数 |
| memory.total | int | 总内存（MB） |
| memory.used | int | 已用内存（MB） |
| memory.free | int | 可用内存（MB） |
| disk.total | int | 磁盘总容量（MB） |
| disk.used | int | 已用磁盘（MB） |
| disk.free | int | 可用磁盘（MB） |
| load.load1/5/15 | float | 1/5/15 分钟平均负载 |
| runtime.panel_uptime | int | 面板运行时间（秒） |
| runtime.go_version | string | Go 版本 |
| runtime.goroutines | int | Goroutine 数量 |
| runtime.memory_alloc | int | 面板内存分配（字节） |
| runtime.db_path | string | 数据库文件路径 |
| runtime.db_size | int | 数据库文件大小（字节） |

## DNS 云厂商管理

DNS 管理接口由 `RegisterDNSRoutes` 注册，提供云厂商 DNS 管理功能：

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/dns/providers` | DNS 厂商列表 |
| POST | `/api/v1/dns/providers` | 添加 DNS 厂商配置 |
| PUT | `/api/v1/dns/providers/:id` | 更新厂商配置 |
| DELETE | `/api/v1/dns/providers/:id` | 删除厂商配置 |
| GET | `/api/v1/dns/providers/:id/domains` | 获取域名列表 |
| GET | `/api/v1/dns/providers/:id/records` | 获取 DNS 记录 |
| POST | `/api/v1/dns/records` | 添加 DNS 记录 |
| PUT | `/api/v1/dns/records/:id` | 更新 DNS 记录 |
| DELETE | `/api/v1/dns/records/:id` | 删除 DNS 记录 |

支持的 DNS 厂商包括：阿里云（Aliyun）、腾讯云（DNSPod）、Cloudflare 等。

## 注意事项

1. **权限要求**：修改系统设置需要 `setting:update` 权限，默认仅超级管理员拥有
2. **设置生效**：大部分设置修改后立即生效，部分设置（如监控采集间隔）可能需要重启服务
3. **SMTP 安全**：SMTP 密码加密存储，API 返回时脱敏
4. **更新检查**：检查更新需要服务器能访问外部网络，如无法连接请检查网络和防火墙
5. **清理风险**：系统清理操作不可撤销，执行前确认不需要相关文件
6. **数据库备份**：修改系统设置前建议先备份数据库文件（cloudpivot.db）
