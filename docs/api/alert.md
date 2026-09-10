# 告警 API

告警通知模块提供告警规则管理、通知渠道管理和告警历史记录功能。

## 告警规则

### 规则列表

`GET /api/v1/alert/rules`

获取告警规则列表（分页）。

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| pageSize | int | 10 | 每页条数 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "name": "CPU 使用率过高",
        "metric": "cpu",
        "condition": ">",
        "threshold": 80,
        "duration": 300,
        "level": "warning",
        "enabled": true,
        "channel_ids": [1, 2],
        "description": "CPU 使用率持续超过 80% 时告警",
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-09-01T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```

### 规则对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 规则 ID |
| name | string | 规则名称 |
| metric | string | 监控指标：cpu / memory / disk / network_in / network_out / load |
| condition | string | 条件：> / < / >= / <= / == |
| threshold | float | 阈值 |
| duration | int | 持续时间（秒） |
| level | string | 级别：info / warning / critical |
| enabled | bool | 是否启用 |
| channel_ids | array | 关联的通知渠道 ID 列表 |
| description | string | 规则描述 |
| created_at | string | 创建时间 |
| updated_at | string | 更新时间 |

### 创建规则

`POST /api/v1/alert/rules`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 规则名称 |
| metric | string | 是 | 监控指标 |
| condition | string | 是 | 条件 |
| threshold | float | 是 | 阈值 |
| duration | int | 否 | 持续时间（秒，默认 60） |
| level | string | 否 | 级别（默认 warning） |
| channel_ids | array | 否 | 通知渠道 ID 列表 |
| description | string | 否 | 描述 |

### 响应示例

返回创建的规则对象（enabled 默认为 true）。

### 获取规则详情

`GET /api/v1/alert/rules/:id`

返回单个规则对象。

### 更新规则

`PUT /api/v1/alert/rules/:id`

字段同创建规则，所有字段可选。

### 删除规则

`DELETE /api/v1/alert/rules/:id`

### 启用规则

`POST /api/v1/alert/rules/:id/enable`

### 禁用规则

`POST /api/v1/alert/rules/:id/disable`

### 测试规则

`POST /api/v1/alert/rules/:id/test`

手动触发一次规则评估，用于测试通知渠道是否正常。

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "tested": true
  }
}
```

## 通知渠道

### 渠道列表

`GET /api/v1/alert/channels`

获取通知渠道列表（分页）。

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "name": "管理员邮件",
        "type": "email",
        "config": {
          "to": "admin@example.com"
        },
        "enabled": true,
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z"
      },
      {
        "id": 2,
        "name": "企业微信告警",
        "type": "webhook",
        "config": {
          "url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx",
          "method": "POST",
          "headers": { "Content-Type": "application/json" },
          "template": "{\"msgtype\":\"text\",\"text\":{\"content\":\"{message}\"}}"
        },
        "enabled": true,
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 2,
    "page": 1,
    "pageSize": 10
  }
}
```

### 渠道对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 渠道 ID |
| name | string | 渠道名称 |
| type | string | 类型：email / webhook |
| config | object | 渠道配置（结构因类型而异） |
| enabled | bool | 是否启用 |
| created_at | string | 创建时间 |
| updated_at | string | 更新时间 |

### 邮件渠道配置

| 字段 | 类型 | 说明 |
|------|------|------|
| to | string | 收件人邮箱（多个用逗号分隔） |

SMTP 服务器配置在系统设置中统一管理。

### Webhook 渠道配置

| 字段 | 类型 | 说明 |
|------|------|------|
| url | string | Webhook URL |
| method | string | HTTP 方法（默认 POST） |
| headers | object | 自定义请求头 |
| template | string | 请求体模板，支持变量替换 |

#### Webhook 模板变量

| 变量 | 说明 |
|------|------|
| `{rule_name}` | 告警规则名称 |
| `{metric}` | 监控指标 |
| `{value}` | 当前值 |
| `{threshold}` | 阈值 |
| `{level}` | 告警级别 |
| `{time}` | 触发时间 |
| `{message}` | 告警消息 |

### 创建渠道

`POST /api/v1/alert/channels`

### 获取渠道详情

`GET /api/v1/alert/channels/:id`

### 更新渠道

`PUT /api/v1/alert/channels/:id`

### 删除渠道

`DELETE /api/v1/alert/channels/:id`

### 测试渠道

`POST /api/v1/alert/channels/:id/test`

发送一条测试告警消息到该渠道。

### 请求体（可选）

| 字段 | 类型 | 说明 |
|------|------|------|
| message | string | 自定义测试消息内容 |

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

## 告警历史

### 历史列表

`GET /api/v1/alert/history`

获取告警触发历史记录（分页）。

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| pageSize | int | 20 | 每页条数 |
| rule_id | string | - | 按规则 ID 筛选 |
| status | string | - | 按状态筛选：triggered / recovered / notified / failed |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "rule_id": 1,
        "rule_name": "CPU 使用率过高",
        "level": "warning",
        "metric_value": 85.5,
        "threshold": 80,
        "status": "recovered",
        "triggered_at": "2026-09-09T10:00:00Z",
        "recovered_at": "2026-09-09T10:15:00Z",
        "message": "CPU 使用率 85.5% 超过阈值 80%"
      }
    ],
    "total": 50,
    "page": 1,
    "pageSize": 20
  }
}
```

### 历史对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 记录 ID |
| rule_id | uint | 规则 ID |
| rule_name | string | 规则名称 |
| level | string | 告警级别 |
| metric_value | float | 触发时的指标值 |
| threshold | float | 规则阈值 |
| status | string | 状态：triggered / recovered / notified / failed |
| triggered_at | string | 触发时间 |
| recovered_at | string | 恢复时间（如已恢复） |
| message | string | 告警消息 |

## 告警引擎工作流程

1. 监控采集器按间隔采集系统指标
2. 告警引擎对每条启用的规则进行评估
3. 检查当前指标值是否满足触发条件
4. 异常状态持续超过设定时间后触发告警
5. 通过关联的通知渠道发送告警消息
6. 记录告警历史
7. 指标恢复正常后记录恢复时间

## 注意事项

1. **引擎启动**：告警引擎在服务启动时自动初始化并运行
2. **评估频率**：告警规则评估频率与监控采集间隔一致（默认 60 秒）
3. **通知失败**：通知发送失败会记录到告警历史，状态为 failed
4. **规则测试**：测试规则会立即发送一条测试告警，用于验证渠道配置
5. **渠道关联**：删除渠道前需确保没有规则关联该渠道，或先解除关联
