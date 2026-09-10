# API 概览

云枢面板提供完整的 RESTful API，所有接口基于 HTTP 协议，返回 JSON 格式数据。API 前缀为 `/api/v1`，后端默认监听 `8888` 端口。

## 基础信息

| 项目 | 值 |
|------|-----|
| API 前缀 | `/api/v1` |
| 默认端口 | `8888` |
| 数据格式 | JSON (`application/json`) |
| 字符编码 | UTF-8 |
| 健康检查 | `GET /health` |

## 认证方式

### Bearer Token

除登录和刷新令牌接口外，所有 API 都需要在请求头中携带 JWT 访问令牌：

```http
Authorization: Bearer <access_token>
```

### 获取令牌

1. 调用 `POST /api/v1/auth/login` 传入用户名和密码
2. 成功后返回 `access_token` 和 `refresh_token`
3. `access_token` 有效期为 2 小时（可配置）
4. `refresh_token` 有效期为 7 天，用于刷新访问令牌

### 令牌刷新

当 `access_token` 过期后，使用 `refresh_token` 调用 `POST /api/v1/auth/refresh` 获取新的令牌对。

### TOTP 二次验证

如用户启用了 TOTP，登录流程为：

1. 调用登录接口，返回 `require_totp: true` 和 `temp_token`
2. 使用 `temp_token` 调用 TOTP 验证接口
3. 验证通过后返回正式的 `access_token`

## 统一响应格式

所有 API 响应遵循统一格式：

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 业务状态码，200 表示成功 |
| message | string | 响应消息，成功为 "success"，失败为错误描述 |
| data | object/array/null | 响应数据，失败时为 null |

### HTTP 状态码

HTTP 状态码与业务错误码对应：

| 业务码 | HTTP 状态码 | 说明 |
|--------|------------|------|
| 200 | 200 OK | 请求成功 |
| 40001 | 400 Bad Request | 参数错误 |
| 40101 | 401 Unauthorized | 未认证或令牌无效 |
| 40301 | 403 Forbidden | 无权限访问 |
| 40401 | 404 Not Found | 资源不存在 |
| 50001 | 500 Internal Server Error | 服务器内部错误 |

## 错误码表

| 错误码 | 标识 | 说明 |
|--------|------|------|
| 200 | `CodeSuccess` | 请求成功 |
| 40001 | `CodeParamError` | 请求参数错误，包含参数缺失、格式错误、校验失败等 |
| 40101 | `CodeUnauthorized` | 未认证，令牌缺失、无效或已过期 |
| 40301 | `CodeForbidden` | 已认证但无权限执行该操作 |
| 40401 | `CodeNotFound` | 请求的资源不存在 |
| 50001 | `CodeServerError` | 服务器内部错误，如数据库操作失败、外部服务异常等 |

### 错误响应示例

```json
{
  "code": 40001,
  "message": "参数错误: Key: 'LoginRequest.Username' Error:Field validation for 'Username' failed on the 'required' tag",
  "data": null
}
```

## 分页格式

列表类接口使用统一的分页响应格式，`data` 字段结构为：

```json
{
  "items": [],
  "total": 100,
  "page": 1,
  "pageSize": 10
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| items | array | 当前页数据列表 |
| total | int64 | 总记录数 |
| page | int | 当前页码（从 1 开始） |
| pageSize | int | 每页条数 |

### 分页请求参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码，从 1 开始 |
| pageSize | int | 10 | 每页条数 |

> 注意：部分接口使用 `page_size`（下划线）作为参数名，具体以各接口文档为准。

## 限流策略

API 接口实施限流保护：

| 接口类型 | 限流 | 说明 |
|----------|------|------|
| 登录接口 | 10 次/分钟 | 基于 IP 限流，防止暴力破解 |
| 其他接口 | 100 次/分钟 | 基于 IP 限流，防止滥用 |

超过限流将返回 429 Too Many Requests。

## 安全相关

- 所有写操作（POST/PUT/DELETE）需要 CSRF 防护（通过 JWT 实现）
- 敏感操作（删除用户、修改系统设置）需要特定权限
- 所有 API 请求会被审计日志记录（用户、时间、IP、操作、结果）
- 登出后 access_token 会被加入黑名单，2 小时内不可使用

## API 模块索引

| 模块 | 文档 | 路由前缀 |
|------|------|----------|
| 认证 | [auth.md](./auth.md) | `/auth` |
| Dashboard | [dashboard.md](./dashboard.md) | `/dashboard` |
| 网站 | [website.md](./website.md) | `/sites` |
| 数据库 | [database.md](./database.md) | `/databases` |
| 文件 | [file.md](./file.md) | `/files` |
| 终端 | [terminal.md](./terminal.md) | `/ws/terminal` |
| 容器 | [container.md](./container.md) | `/containers`, `/images` |
| 应用商店 | [appstore.md](./appstore.md) | `/apps` |
| 备份 | [backup.md](./backup.md) | `/backups` |
| 计划任务 | [cron.md](./cron.md) | `/cron/tasks` |
| 防火墙 | [firewall.md](./firewall.md) | `/firewall` |
| SSH | [ssh.md](./ssh.md) | `/ssh` |
| 告警 | [alert.md](./alert.md) | `/alert` |
| 用户/角色 | [users.md](./users.md) | `/users`, `/roles`, `/permissions` |
| 系统设置 | [system.md](./system.md) | `/system` |

## cURL 示例

### 登录获取令牌

```bash
curl -X POST http://localhost:8888/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 携带令牌请求

```bash
curl -X GET http://localhost:8888/api/v1/dashboard/overview \
  -H "Authorization: Bearer <access_token>"
```

### 分页请求

```bash
curl -X GET "http://localhost:8888/api/v1/sites?page=1&pageSize=20" \
  -H "Authorization: Bearer <access_token>"
```
