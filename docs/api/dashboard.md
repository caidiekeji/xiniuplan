# Dashboard API

Dashboard 模块提供系统资源概览、监控指标趋势和最近审计日志。

## 获取资源概览

`GET /api/v1/dashboard/overview`

获取服务器 CPU、内存、磁盘、网络等资源的实时概览数据。

### 请求头

```
Authorization: Bearer <access_token>
```

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "cpu": {
      "usage_percent": 23.5,
      "cores": 4,
      "model": "Intel(R) Xeon(R) CPU E5-2680 v4 @ 2.40GHz"
    },
    "memory": {
      "total": 8192,
      "used": 3200,
      "free": 4992,
      "usage_percent": 39.1
    },
    "disk": {
      "total": 102400,
      "used": 45000,
      "free": 57400,
      "usage_percent": 43.9,
      "mount_point": "/"
    },
    "network": {
      "in_bytes": 1048576,
      "out_bytes": 524288,
      "in_rate": 1024.0,
      "out_rate": 512.0
    },
    "load": {
      "load1": 0.45,
      "load5": 0.32,
      "load15": 0.28
    },
    "uptime": 864000,
    "hostname": "cloudpivot-server"
  }
}
```

### 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| cpu.usage_percent | float | CPU 使用率百分比 |
| cpu.cores | int | CPU 核心数 |
| cpu.model | string | CPU 型号 |
| memory.total | int | 总内存（MB） |
| memory.used | int | 已用内存（MB） |
| memory.free | int | 可用内存（MB） |
| memory.usage_percent | float | 内存使用率百分比 |
| disk.total | int | 磁盘总容量（MB） |
| disk.used | int | 已用磁盘（MB） |
| disk.free | int | 可用磁盘（MB） |
| disk.usage_percent | float | 磁盘使用率百分比 |
| network.in_bytes | int | 累计入站流量（字节） |
| network.out_bytes | int | 累计出站流量（字节） |
| network.in_rate | float | 入站速率（KB/s） |
| network.out_rate | float | 出站速率（KB/s） |
| load.load1 | float | 1 分钟平均负载 |
| load.load5 | float | 5 分钟平均负载 |
| load.load15 | float | 15 分钟平均负载 |
| uptime | int | 系统运行时间（秒） |
| hostname | string | 主机名 |

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40101 | 未认证 |
| 50001 | 获取资源概览失败 |

## 获取监控指标

`GET /api/v1/dashboard/metrics`

获取指定资源指标的历史趋势数据。

### 请求头

```
Authorization: Bearer <access_token>
```

### 查询参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| type | string | 否 | cpu | 指标类型：cpu / memory / disk / network |
| range | string | 否 | 24h | 时间范围：1h / 6h / 24h / 7d |

### 请求示例

```
GET /api/v1/dashboard/metrics?type=cpu&range=24h
```

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "type": "cpu",
    "range": "24h",
    "points": [
      {
        "timestamp": "2026-09-09T09:00:00Z",
        "value": 25.3
      },
      {
        "timestamp": "2026-09-09T10:00:00Z",
        "value": 30.1
      }
    ]
  }
}
```

### 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| type | string | 指标类型 |
| range | string | 时间范围 |
| points | array | 数据点列表 |
| points[].timestamp | string | 数据点时间（RFC3339） |
| points[].value | float | 指标值 |

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 无效的指标类型或时间范围 |
| 40101 | 未认证 |

## 获取最近审计日志

`GET /api/v1/dashboard/audit-logs`

获取最近的用户操作审计日志（用于仪表盘展示，返回最近若干条）。

### 请求头

```
Authorization: Bearer <access_token>
```

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "username": "admin",
      "action": "login",
      "resource": "auth",
      "ip": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "status": "success",
      "created_at": "2026-09-09T10:30:00Z"
    }
  ]
}
```

### 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 日志 ID |
| user_id | uint | 操作用户 ID |
| username | string | 操作用户名 |
| action | string | 操作类型 |
| resource | string | 操作资源 |
| ip | string | 操作者 IP |
| user_agent | string | 浏览器 User-Agent |
| status | string | 操作结果（success / failed） |
| created_at | string | 操作时间 |

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40101 | 未认证 |
| 50001 | 获取审计日志失败 |
