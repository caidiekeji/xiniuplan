# 架构设计

本文档详细描述云枢面板的系统架构，包括后端分层架构、前端架构、数据库模型、安全设计和部署架构。

## 整体架构

云枢面板采用前后端分离架构，后端为 Go 单体应用，前端为 Vue3 SPA。

```
┌─────────────────────────────────────────────────────┐
│                    客户端（浏览器）                      │
│  ┌───────────────────────────────────────────────┐  │
│  │           Vue3 SPA (Naive UI)                 │  │
│  │  Views / Components / Stores / Router / API   │  │
│  └───────────────────┬───────────────────────────┘  │
└──────────────────────┼───────────────────────────────┘
                       │ HTTP / WebSocket
                       ▼
┌─────────────────────────────────────────────────────┐
│              Go 后端 (Gin + GORM)                     │
│  ┌───────────────────────────────────────────────┐  │
│  │  Middleware (CORS/JWT/限流/审计/Recovery)     │  │
│  ├───────────────────────────────────────────────┤  │
│  │  API Handler (api/)                            │  │
│  │  认证/网站/数据库/文件/终端/容器/应用/备份...  │  │
│  ├───────────────────────────────────────────────┤  │
│  │  Service (service/) 业务逻辑层                 │  │
│  │  auth/site/database/file/container/...        │  │
│  ├───────────────────────────────────────────────┤  │
│  │  Model (model/) 数据模型层                      │  │
│  │  GORM Models + SQLite                          │  │
│  ├───────────────────────────────────────────────┤  │
│  │  DTO (dto/) 数据传输对象                       │  │
│  │  Request/Response 结构体                       │  │
│  ├───────────────────────────────────────────────┤  │
│  │  Common (common/) 公共组件                     │  │
│  │  统一响应/错误码/命令白名单                     │  │
│  ├───────────────────────────────────────────────┤  │
│  │  Config (config/) 配置加载                     │  │
│  └───────────────────────────────────────────────┘  │
└──────────┬──────────┬──────────┬────────────────────┘
           │          │          │
     ┌─────▼──┐ ┌────▼────┐ ┌──▼─────────┐
     │ SQLite │ │  Docker  │ │  系统命令   │
     │  数据库 │ │  API     │ │ (白名单)    │
     └────────┘ └─────────┘ └────────────┘
```

## 后端分层架构

后端采用经典的四层架构：API → Service → Model → DTO，配合公共能力 pkg/。

### 目录结构

```
backend/
├── cmd/
│   └── agent/
│       └── main.go              # 应用入口，初始化各组件和路由
├── internal/
│   ├── api/                     # API Handler 层
│   │   ├── auth.go              # 认证 API
│   │   ├── dashboard.go         # 仪表盘 API
│   │   ├── site.go              # 网站管理 API
│   │   ├── database.go          # 数据库管理 API
│   │   ├── file.go              # 文件管理 API
│   │   ├── terminal.go          # Web 终端 API (WebSocket)
│   │   ├── container.go         # 容器管理 API
│   │   ├── appstore.go          # 应用商店 API
│   │   ├── backup.go            # 备份恢复 API
│   │   ├── cron.go              # 计划任务 API
│   │   ├── firewall.go          # 防火墙 API
│   │   ├── ssh.go               # SSH 安全 API
│   │   ├── alert.go             # 告警通知 API
│   │   ├── user.go              # 用户管理 API
│   │   ├── role.go              # 角色权限 API
│   │   ├── system.go            # 系统设置 API
│   │   ├── postgres.go          # PostgreSQL 专用 API
│   │   ├── redis.go             # Redis 专用 API
│   │   ├── dns.go               # DNS 管理 API
│   │   ├── totp.go              # TOTP 二次验证 API
│   │   ├── monitor.go           # 监控增强 API
│   │   └── audit.go             # 审计日志 API
│   ├── service/                 # 业务逻辑层
│   │   ├── auth/                # 认证服务（JWT/登录/TOTP）
│   │   ├── site/                # 网站服务（Nginx/SSL）
│   │   ├── database/            # 数据库服务（实例/库表/用户/SQL）
│   │   ├── file/                # 文件服务（沙箱文件管理）
│   │   ├── terminal/            # 终端服务（SSH/PTY/WebSocket）
│   │   ├── container/           # 容器服务（Docker API 封装）
│   │   ├── appstore/            # 应用商店服务
│   │   ├── backup/              # 备份服务
│   │   ├── cron/                # 计划任务服务（Cron 引擎）
│   │   ├── firewall/            # 防火墙服务
│   │   ├── ssh/                 # SSH 安全服务
│   │   ├── alert/               # 告警服务（规则引擎/通知）
│   │   ├── user/                # 用户服务
│   │   ├── monitor/             # 监控服务（采集器/指标查询）
│   │   └── system/              # 系统服务（设置/SMTP/清理）
│   ├── model/                   # 数据模型层
│   │   ├── db.go                # 数据库初始化（GORM + SQLite）
│   │   ├── user.go              # 用户模型
│   │   ├── rbac.go              # 角色/权限模型
│   │   ├── site.go              # 网站模型
│   │   ├── database.go          # 数据库实例模型
│   │   ├── backup.go            # 备份模型
│   │   ├── cron.go              # 计划任务模型
│   │   ├── firewall.go          # 防火墙规则模型
│   │   ├── alert.go             # 告警规则/渠道/历史模型
│   │   ├── app.go               # 应用模型
│   │   ├── setting.go           # 设置模型
│   │   ├── audit.go             # 审计日志模型
│   │   ├── monitor.go           # 监控数据模型
│   │   ├── container.go         # 容器模型
│   │   ├── dns.go               # DNS 模型
│   │   ├── app_source.go        # 应用源模型
│   │   └── recovery_code.go     # TOTP 恢复码模型
│   ├── dto/                     # 数据传输对象
│   │   ├── auth.go              # 认证相关 DTO
│   │   ├── site.go              # 网站相关 DTO
│   │   ├── database.go          # 数据库相关 DTO
│   │   └── ...                  # 其他模块 DTO
│   ├── middleware/              # 中间件
│   │   ├── jwt.go               # JWT 认证中间件
│   │   ├── cors.go              # CORS 跨域中间件
│   │   ├── ratelimit.go         # 限流中间件
│   │   ├── audit.go             # 审计日志中间件
│   │   └── permission.go        # RBAC 权限中间件
│   ├── common/                  # 公共组件
│   │   ├── response.go          # 统一响应/错误码/分页
│   │   └── command.go           # 命令白名单执行器
│   └── config/                  # 配置
│       └── config.go            # 配置加载（config.yaml）
├── config.yaml                  # 配置文件
├── go.mod                       # Go 模块定义
└── go.sum                       # 依赖校验
```

### 各层职责

| 层级 | 目录 | 职责 | 依赖方向 |
|------|------|------|----------|
| API 层 | api/ | HTTP 请求处理、参数校验、调用 Service、返回响应 | → Service |
| Service 层 | service/ | 业务逻辑、事务管理、外部系统交互（Docker/SSH/系统命令） | → Model |
| Model 层 | model/ | 数据模型定义、数据库操作（GORM）、数据库初始化 | → SQLite |
| DTO 层 | dto/ | 请求/响应数据结构定义，与 API 层配合 | 无外部依赖 |
| Middleware | middleware/ | 横切关注点：认证、授权、限流、审计、CORS | → Model/Service |
| Common | common/ | 公共工具：统一响应、错误码、命令执行 | 无外部依赖 |
| Config | config/ | 配置文件加载和解析 | 无外部依赖 |

### 数据流

典型请求处理流程：

```
HTTP Request
    │
    ▼
Gin Router (路由匹配)
    │
    ▼
Middleware 链:
  1. CORS()          — 跨域处理
  2. RateLimit()     — 限流（登录10/min，其他100/min）
  3. Audit()         — 审计日志记录
  4. JWTAuth()       — JWT 令牌验证（公开路由跳过）
  5. RequirePermission() — RBAC 权限校验（部分路由）
    │
    ▼
API Handler (参数绑定与校验)
    │
    ▼
Service (业务逻辑处理)
    │  ├── 数据库操作 → Model → GORM → SQLite
    │  ├── Docker 操作 → Docker API
    │  ├── 系统命令 → Common.Command (白名单)
    │  └── SSH 操作 → golang.org/x/crypto/ssh
    │
    ▼
DTO Response (组装响应数据)
    │
    ▼
Common.Success / Common.Fail (统一响应格式)
    │
    ▼
HTTP Response (JSON)
```

## 前端架构

前端采用 Vue 3 + TypeScript + Naive UI 技术栈，使用 Vite 构建，Pinia 管理状态，Vue Router 管理路由。

### 目录结构

```
frontend/
├── src/
│   ├── api/                     # API 请求封装
│   │   ├── auth.ts              # 认证 API
│   │   ├── dashboard.ts         # 仪表盘 API
│   │   ├── website.ts           # 网站 API
│   │   ├── database.ts          # 数据库 API
│   │   ├── file.ts              # 文件 API
│   │   ├── terminal.ts          # 终端 API
│   │   ├── container.ts         # 容器 API
│   │   ├── appstore.ts          # 应用商店 API
│   │   ├── backup.ts            # 备份 API
│   │   ├── cron.ts              # 计划任务 API
│   │   ├── firewall.ts          # 防火墙 API
│   │   ├── ssh.ts               # SSH API
│   │   ├── alert.ts             # 告警 API
│   │   ├── users.ts             # 用户/角色 API
│   │   └── system.ts            # 系统设置 API
│   ├── components/              # 通用组件
│   │   ├── layout/              # 布局组件
│   │   ├── common/              # 通用 UI 组件
│   │   └── business/            # 业务组件
│   ├── views/                   # 页面视图
│   │   ├── login/               # 登录页
│   │   ├── dashboard/           # 仪表盘
│   │   ├── website/             # 网站管理
│   │   ├── database/            # 数据库管理
│   │   ├── file/                # 文件管理器
│   │   ├── terminal/            # Web 终端
│   │   ├── container/           # 容器管理
│   │   ├── appstore/            # 应用商店
│   │   ├── backup/              # 备份恢复
│   │   ├── cron/                # 计划任务
│   │   ├── firewall/            # 防火墙
│   │   ├── ssh/                 # SSH 安全
│   │   ├── alert/               # 告警通知
│   │   ├── users/               # 用户与权限
│   │   ├── settings/            # 系统设置
│   │   ├── monitor/             # 监控
│   │   ├── audit/               # 审计日志
│   │   └── ssl/                 # SSL 管理
│   ├── stores/                  # Pinia 状态管理
│   │   ├── auth.ts              # 认证状态（令牌/用户信息）
│   │   ├── app.ts               # 应用全局状态
│   │   └── ...                  # 其他模块状态
│   ├── router/                  # 路由配置
│   │   └── index.ts             # 路由定义与导航守卫
│   ├── layouts/                 # 布局组件
│   │   ├── DefaultLayout.vue    # 默认布局（侧边栏+顶栏+内容区）
│   │   └── BlankLayout.vue      # 空白布局（登录页等）
│   ├── composables/             # 组合式函数
│   │   ├── useTable.ts          # 表格数据管理
│   │   ├── usePagination.ts     # 分页逻辑
│   │   └── ...                  # 其他组合式函数
│   ├── locales/                 # 国际化
│   │   ├── zh-CN.ts             # 中文
│   │   └── en-US.ts             # 英文
│   ├── styles/                  # 全局样式
│   │   └── index.css            # 样式入口
│   ├── assets/                  # 静态资源
│   ├── App.vue                  # 根组件
│   ├── main.ts                  # 应用入口
│   └── env.d.ts                 # 类型声明
├── index.html                   # HTML 入口
├── vite.config.ts               # Vite 配置（含代理）
├── package.json                 # 依赖配置
└── tsconfig.json                # TypeScript 配置
```

### 前端数据流

```
用户操作
    │
    ▼
View (页面组件)
    │
    ├── 调用 Store Action → Pinia Store (状态管理)
    │                         │
    │                         ▼
    │                    API Module (请求封装)
    │                         │
    │                         ▼
    │                    HTTP Client (axios)
    │                         │
    │                         ▼
    │                    后端 API
    │
    └── 直接调用 API Module (简单场景)
```

### 路由与权限

- 路由定义在 `router/index.ts`，包含路由元信息（标题、图标、权限要求）
- 导航守卫检查登录状态和权限
- 侧边栏菜单根据用户权限动态生成
- 无权限的路由和菜单自动隐藏

## 数据库模型

### 数据库选型

使用 SQLite 作为默认数据库，基于 GORM ORM 操作。SQLite 具有零配置、单文件、易备份的特点，适合轻量级面板场景。

数据库文件默认路径：`./cloudpivot.db`（可在 config.yaml 中配置）。

### 核心模型关系

```
┌──────────┐     ┌──────────┐     ┌──────────────┐
│   User   │────▶│   Role   │────▶│ Permission   │
│ (用户)   │ N:1 │ (角色)   │ 1:N │ (权限码)     │
└──────────┘     └──────────┘     └──────────────┘
     │
     │ 1:N
     ▼
┌──────────┐     ┌──────────┐     ┌──────────────┐
│ AuditLog │     │   Site   │     │ DatabaseInst │
│(审计日志) │     │ (网站)   │     │ (数据库实例)  │
└──────────┘     └──────────┘     └──────────────┘
                       │                  │
                       │ 1:N              │ 1:N
                       ▼                  ▼
                 ┌──────────┐     ┌──────────────┐
                 │  Backup  │     │  CronTask    │
                 │ (备份)   │     │ (计划任务)    │
                 └──────────┘     └──────┬───────┘
                                            │ 1:N
                                            ▼
                                      ┌──────────┐
                                      │ CronLog  │
                                      │(执行日志) │
                                      └──────────┘

┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ AlertRule    │     │ AlertChannel │     │ AlertHistory │
│ (告警规则)    │────▶│ (通知渠道)    │     │ (告警历史)    │
└──────────────┘     └──────────────┘     └──────────────┘

┌──────────────┐     ┌──────────────┐
│ FirewallRule │     │  Setting     │
│ (防火墙规则)  │     │  (设置键值对) │
└──────────────┘     └──────────────┘
```

### 主要模型表

| 表名 | 模型 | 说明 |
|------|------|------|
| users | User | 用户账号 |
| roles | Role | 角色（含权限列表 JSON） |
| sites | Site | 网站 |
| database_instances | DatabaseInstance | 数据库实例 |
| backups | Backup | 备份记录 |
| cron_tasks | CronTask | 计划任务 |
| cron_logs | CronLog | 任务执行日志 |
| firewall_rules | FirewallRule | 防火墙规则 |
| alert_rules | AlertRule | 告警规则 |
| alert_channels | AlertChannel | 通知渠道 |
| alert_histories | AlertHistory | 告警历史 |
| installed_apps | InstalledApp | 已安装应用 |
| app_sources | AppSource | 应用源 |
| settings | Setting | 系统设置（键值对） |
| audit_logs | AuditLog | 审计日志 |
| monitor_data | MonitorData | 监控数据 |
| recovery_codes | RecoveryCode | TOTP 恢复码 |

### 关键模型字段

#### User（用户）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| username | string | 用户名（唯一） |
| password_hash | string | bcrypt 密码哈希（不返回前端） |
| role | string | 角色 code（兼容字段） |
| role_id | uint | 关联 roles.id |
| totp_secret | string | TOTP 密钥（不返回前端） |
| status | string | active / disabled |
| last_login_at | time | 最后登录时间 |
| created_at | time | 创建时间 |

#### Role（角色）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | string | 角色名称 |
| code | string | 角色标识码（唯一） |
| description | string | 描述 |
| permissions | StringSlice | 权限码列表（JSON 存储） |
| is_system | bool | 是否内置角色 |
| created_at | time | 创建时间 |

#### Site（网站）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| primary_domain | string | 主域名 |
| domains | StringSlice | 附加域名列表 |
| type | string | static / runtime / reverse_proxy |
| runtime | string | 运行时类型 |
| root_dir | string | 网站根目录 |
| nginx_conf | string | Nginx 配置路径 |
| ssl_enabled | bool | 是否启用 SSL |
| ssl_provider | string | manual / letsencrypt |
| ssl_cert_path | string | 证书路径 |
| ssl_key_path | string | 私钥路径 |
| ssl_expire_at | time | 证书到期时间 |
| ssl_auto_renew | bool | 自动续期 |
| status | string | running / stopped / error |
| container_id | string | 关联容器 ID |
| remark | string | 备注 |

## 安全设计

### 认证安全

| 机制 | 实现 | 说明 |
|------|------|------|
| JWT 认证 | RS256 非对称加密 | 使用 RSA 密钥对签名，公钥验证，私钥签名 |
| 令牌类型 | Access Token + Refresh Token | Access 2小时，Refresh 7天 |
| 令牌黑名单 | 登出时加入黑名单 | 登出后 Token 2小时内不可使用 |
| 密码存储 | bcrypt 哈希 | cost=10，不可逆加密 |
| TOTP 二次验证 | RFC 6238 标准 | 支持 Google Authenticator 等 |
| TOTP 恢复码 | 一次性恢复码 | 手机丢失时紧急登录 |

### 授权安全（RBAC）

- 基于角色的访问控制（RBAC）
- 15 个模块、45 个权限点
- 权限码格式：`模块:动作`（如 `site:create`）
- 三个内置角色：超级管理员（`*`）、管理员、普通用户
- 支持自定义角色和权限组合
- API 级别权限校验：`middleware.RequirePermission(perm)`
- 前端菜单和按钮级权限控制

### 输入安全

| 机制 | 实现 | 说明 |
|------|------|------|
| SQL 注入防护 | GORM 参数化查询 | 所有数据库操作使用预编译语句 |
| 命令注入防护 | 命令白名单 | common/command.go 限制可执行命令 |
| XSS 防护 | 前端转义 + CSP | Vue 默认转义，配置内容安全策略 |
| 文件上传安全 | 类型校验 + 大小限制 | 限制上传文件类型和大小 |
| 路径遍历防护 | 沙箱目录限制 | 文件操作限制在指定沙箱目录内 |

### 传输安全

- 生产环境建议使用 HTTPS
- WebSocket 终端使用 WSS 加密
- JWT Token 通过 HTTP Header 传输
- 敏感配置（SMTP 密码等）加密存储

### 审计安全

- 所有 API 请求自动记录审计日志
- 审计内容：用户、时间、IP、User-Agent、操作、资源、结果
- 审计日志不可篡改（仅追加）
- 登录失败、权限拒绝等安全事件重点记录
- 仪表盘展示最近审计日志

### 限流防护

| 接口类型 | 限流 | 说明 |
|----------|------|------|
| 登录接口 | 10 次/分钟 | 基于 IP，防止暴力破解 |
| 其他接口 | 100 次/分钟 | 基于 IP，防止滥用 |

### SSH 安全

- 支持修改 SSH 端口、禁用密码登录、禁用 root 登录
- SSH 密钥管理（添加/删除公钥）
- Fail2ban 集成（自动封禁暴力破解 IP）
- SSH 登录日志审计

## 部署架构

### 开发环境

```
开发者机器
├── 后端: go run cmd/agent/main.go (端口 8888)
├── 前端: npm run dev (端口 5173, 代理 /api → 8888)
└── 数据库: SQLite (./cloudpivot.db)
```

### 生产环境（单机部署）

```
用户浏览器
    │
    ▼ HTTPS
┌─────────────────────────────────┐
│         Nginx (反向代理)          │
│  ├── 前端静态资源 (dist/)        │
│  └── /api/* → 后端 :8888        │
└──────────────┬──────────────────┘
               │
    ┌──────────▼──────────┐
    │  云枢面板后端 (8888)  │
    │  systemd 服务管理     │
    │  开机自启             │
    └──────┬───────┬──────┘
           │       │
    ┌──────▼──┐ ┌──▼──────┐
    │ SQLite  │ │ Docker   │
    │ 数据库   │ │ 容器引擎  │
    └─────────┘ └─────────┘
```

### 生产部署组件

| 组件 | 说明 |
|------|------|
| 后端二进制 | `go build` 编译的单二进制文件 |
| 前端静态资源 | `npm run build` 产出的 dist/ 目录 |
| systemd 服务 | `cloudpivot.service`，管理进程生命周期 |
| Nginx | 反向代理 + 静态资源服务 + HTTPS 终端 |
| SQLite | 单文件数据库，随面板一起备份 |
| Docker | 数据库实例和应用商店应用的运行时 |
| 配置文件 | config.yaml，包含端口、数据库路径、JWT 密钥目录等 |

### 端口规划

| 端口 | 服务 | 说明 |
|------|------|------|
| 80 | Nginx HTTP | 网站访问（HTTP，自动跳转 HTTPS） |
| 443 | Nginx HTTPS | 网站访问（HTTPS） |
| 8888 | 云枢面板后端 | 面板 API（生产环境建议不对外暴露，由 Nginx 代理） |
| 5173 | Vite Dev Server | 前端开发服务器（仅开发环境） |
| 22 | SSH | 远程登录（建议修改为非标准端口） |
| 3306 | MySQL | 数据库实例（按需，建议不对外暴露） |
| 5432 | PostgreSQL | 数据库实例（按需） |
| 6379 | Redis | 数据库实例（按需） |

### 数据目录

```
/opt/cloudpivot/              # 安装目录
├── cloudpivot                # 后端二进制
├── config.yaml               # 配置文件
├── cloudpivot.db             # SQLite 数据库
├── dist/                     # 前端静态资源
├── runtime/
│   └── keys/                 # JWT RSA 密钥对
│       ├── private.pem
│       └── public.pem
├── sandbox/                  # 文件管理器沙箱目录
├── backups/                  # 备份文件存储
└── logs/                     # 运行日志
```

### 备份策略

| 数据 | 备份方式 | 频率建议 |
|------|----------|----------|
| SQLite 数据库 | 文件复制 | 每日 |
| 网站文件 | tar 压缩 | 每日/每周 |
| 数据库实例 | mysqldump/pg_dump | 每日 |
| JWT 密钥 | 文件复制 | 一次（变更时） |
| 配置文件 | 文件复制 | 变更时 |

### 高可用考虑

当前版本为单机部署，高可用需考虑：

- 数据库：SQLite 不支持多写入，高可用需迁移到 MySQL/PostgreSQL
- 文件存储：网站文件和备份需共享存储（NFS/对象存储）
- 会话：JWT 无状态，多实例共享 RSA 公钥即可
- Docker：数据库实例和应用需容器编排（K8s/Swarm）
