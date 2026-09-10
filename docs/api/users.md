# 用户/角色 API

用户与权限模块提供基于 RBAC 的用户管理、角色管理和权限查询功能。

## 用户管理

### 用户列表

`GET /api/v1/users`

获取用户列表（分页）。需要 `user:view` 权限。

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| pageSize | int | 20 | 每页条数 |
| keyword | string | - | 按用户名搜索 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "username": "admin",
        "role": "super_admin",
        "role_id": 1,
        "role_name": "超级管理员",
        "status": "active",
        "totp_enabled": true,
        "last_login_at": "2026-09-09T10:30:00Z",
        "created_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 20
  }
}
```

### 用户对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 用户 ID |
| username | string | 用户名 |
| role | string | 角色 code |
| role_id | uint | 角色 ID |
| role_name | string | 角色显示名称 |
| status | string | 状态：active / disabled |
| totp_enabled | bool | 是否启用 TOTP |
| last_login_at | string | 最后登录时间 |
| created_at | string | 创建时间 |

### 创建用户

`POST /api/v1/users`

需要 `user:create` 权限。

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名（3-32 字符，唯一） |
| password | string | 是 | 密码（至少 8 位） |
| role_id | uint | 是 | 角色 ID |
| status | string | 否 | 状态（默认 active） |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 2,
    "username": "ops"
  }
}
```

### 获取用户详情

`GET /api/v1/users/:id`

需要 `user:view` 权限。

### 更新用户

`PUT /api/v1/users/:id`

需要 `user:update` 权限。

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| role_id | uint | 否 | 角色 ID |
| status | string | 否 | 状态 |

### 删除用户

`DELETE /api/v1/users/:id`

需要 `user:delete` 权限。不能删除当前登录用户自己。

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

### 修改密码

`PUT /api/v1/users/:id/password`

需要 `user:update` 权限。

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| old_password | string | 否 | 旧密码（修改自己的密码时必填） |
| new_password | string | 是 | 新密码（至少 8 位） |

> 修改自己的密码需要验证旧密码；管理员修改他人密码无需旧密码。

### 禁用用户

`POST /api/v1/users/:id/disable`

需要 `user:update` 权限。禁用后用户无法登录，已登录会话失效。

### 启用用户

`POST /api/v1/users/:id/enable`

需要 `user:update` 权限。

## 角色管理

### 角色列表

`GET /api/v1/roles`

需要 `user:view` 权限。

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "超级管理员",
      "code": "super_admin",
      "description": "拥有全部权限",
      "permissions": ["*"],
      "is_system": true,
      "created_at": "2026-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "name": "管理员",
      "code": "admin",
      "description": "除用户删除和系统设置更新外的全部权限",
      "permissions": ["site:view", "site:create", "..."],
      "is_system": true,
      "created_at": "2026-01-01T00:00:00Z"
    },
    {
      "id": 3,
      "name": "普通用户",
      "code": "user",
      "description": "仅拥有查看类权限",
      "permissions": ["site:view", "database:view", "..."],
      "is_system": true,
      "created_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

### 角色对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 角色 ID |
| name | string | 角色名称 |
| code | string | 角色唯一标识码 |
| description | string | 描述 |
| permissions | array | 权限码列表（`*` 表示全部权限） |
| is_system | bool | 是否为内置角色 |
| created_at | string | 创建时间 |

### 创建角色

`POST /api/v1/roles`

需要 `user:create` 权限。

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 角色名称 |
| code | string | 是 | 角色 code（唯一，仅字母数字下划线） |
| description | string | 否 | 描述 |
| permissions | array | 否 | 权限码列表 |

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 角色 code 已存在 |

### 更新角色

`PUT /api/v1/roles/:id`

需要 `user:update` 权限。内置角色（is_system=true）不可修改。

### 请求体

所有字段可选，只传需要修改的字段。

### 删除角色

`DELETE /api/v1/roles/:id`

需要 `user:delete` 权限。

### 删除限制

- 内置角色不可删除
- 角色下仍有用户时不可删除（需先将用户移到其他角色）

## 权限树

`GET /api/v1/permissions`

获取所有权限码的树状结构，用于角色编辑时的权限选择。需要 `user:view` 权限。

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "module": "site",
      "label": "网站管理",
      "items": [
        { "code": "site:view", "label": "查看网站" },
        { "code": "site:create", "label": "创建网站" },
        { "code": "site:update", "label": "编辑网站" },
        { "code": "site:delete", "label": "删除网站" }
      ]
    },
    {
      "module": "database",
      "label": "数据库管理",
      "items": [
        { "code": "database:view", "label": "查看数据库" },
        { "code": "database:create", "label": "创建数据库" },
        { "code": "database:update", "label": "编辑数据库" },
        { "code": "database:delete", "label": "删除数据库" }
      ]
    }
  ]
}
```

### 权限树结构

| 字段 | 类型 | 说明 |
|------|------|------|
| module | string | 模块标识 |
| label | string | 模块中文名称 |
| items | array | 权限项列表 |
| items[].code | string | 权限码 |
| items[].label | string | 权限中文名称 |

### 权限模块列表

系统包含 15 个权限模块：

| 模块 | 权限码前缀 | 权限数 |
|------|-----------|--------|
| 网站管理 | site: | 4 |
| 数据库管理 | database: | 4 |
| 文件管理 | file: | 4 |
| Web 终端 | terminal: | 2 |
| 容器管理 | container: | 4 |
| 应用商店 | appstore: | 2 |
| 防火墙 | firewall: | 2 |
| 备份恢复 | backup: | 3 |
| 监控 | monitor: | 1 |
| 用户管理 | user: | 4 |
| 审计日志 | audit: | 1 |
| 系统设置 | setting: | 2 |
| 告警管理 | alert: | 4 |
| 计划任务 | cron: | 4 |
| **合计** | | **45** |

## TOTP 接口

TOTP 二次验证接口由系统设置路由组注册，主要接口：

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/totp/setup` | 初始化 TOTP，返回二维码和密钥 |
| POST | `/api/v1/totp/verify` | 验证 TOTP 验证码并启用 |
| POST | `/api/v1/totp/disable` | 禁用 TOTP（需密码验证） |
| POST | `/api/v1/totp/login` | 登录时的 TOTP 二次验证 |
| GET | `/api/v1/totp/recovery-codes` | 获取恢复码 |

## 注意事项

1. **权限缓存**：用户权限在登录时加载到 JWT Token 中，修改角色权限后用户需重新登录生效
2. **内置角色**：三个内置角色（超级管理员、管理员、普通用户）不可修改和删除
3. **超级管理员**：超级管理员拥有 `*` 权限，不受权限中间件限制
4. **密码安全**：密码使用 bcrypt 哈希存储，无法查看明文
5. **用户删除**：删除用户不会删除其操作审计日志
