# 认证 API

认证模块提供用户登录、登出、令牌刷新和用户信息获取功能。基于 JWT RS256 非对称加密，支持 TOTP 二次验证。

## 登录

`POST /api/v1/auth/login`

使用用户名和密码登录，返回访问令牌和刷新令牌。

### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

### 请求示例

```json
{
  "username": "admin",
  "password": "admin123"
}
```

### 响应示例（普通登录成功）

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJSUzI1NiIs...",
    "expires_in": 7200,
    "token_type": "Bearer"
  }
}
```

### 响应示例（需要 TOTP 二次验证）

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "require_totp": true,
    "temp_token": "eyJhbGciOiJSUzI1NiIs..."
  }
}
```

### 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| access_token | string | JWT 访问令牌，2 小时有效 |
| refresh_token | string | JWT 刷新令牌，7 天有效 |
| expires_in | int | 访问令牌有效期（秒） |
| token_type | string | 令牌类型，固定为 "Bearer" |
| require_totp | bool | 是否需要 TOTP 二次验证 |
| temp_token | string | TOTP 验证用临时令牌 |

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 参数错误（用户名或密码缺失） |
| 40101 | 用户名或密码错误、账号已禁用 |

## 登出

`POST /api/v1/auth/logout`

登出当前用户，将 access_token 加入黑名单（2 小时内不可使用）。

### 请求头

```
Authorization: Bearer <access_token>
```

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

## 刷新令牌

`POST /api/v1/auth/refresh`

使用 refresh_token 获取新的 access_token 和 refresh_token。

### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| refresh_token | string | 是 | 刷新令牌 |

### 请求示例

```json
{
  "refresh_token": "eyJhbGciOiJSUzI1NiIs..."
}
```

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJSUzI1NiIs...",
    "expires_in": 7200,
    "token_type": "Bearer"
  }
}
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 参数错误（refresh_token 缺失） |
| 40101 | 刷新令牌无效或已过期 |

## 获取当前用户信息

`GET /api/v1/auth/info`

获取当前登录用户的详细信息。

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
    "id": 1,
    "username": "admin",
    "role": "super_admin",
    "status": "active",
    "totp_enabled": true,
    "last_login_at": "2026-09-09T10:30:00Z",
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

### 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 用户 ID |
| username | string | 用户名 |
| role | string | 角色 code |
| status | string | 账号状态（active / disabled） |
| totp_enabled | bool | 是否已启用 TOTP 二次验证 |
| last_login_at | string | 最后登录时间（RFC3339 格式） |
| created_at | string | 账号创建时间（RFC3339 格式） |

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40101 | 未认证或令牌无效 |
| 40401 | 用户不存在 |

## TOTP 相关接口

TOTP 接口由系统设置路由组注册，详细信息请参考 [系统设置 API](./system.md) 中的 TOTP 部分。

主要 TOTP 接口包括：

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/totp/setup` | 初始化 TOTP，返回二维码和密钥 |
| POST | `/api/v1/totp/verify` | 验证 TOTP 验证码并启用 |
| POST | `/api/v1/totp/disable` | 禁用 TOTP |
| POST | `/api/v1/totp/login` | 登录时的 TOTP 二次验证 |
| GET | `/api/v1/totp/recovery-codes` | 获取恢复码 |
