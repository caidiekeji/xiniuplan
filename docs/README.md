# 云枢面板（CloudPivot）文档

> 轻量级 Linux 服务器可视化管理平台

云枢面板是一款基于 Go + Vue3 构建的现代化服务器管理面板，提供网站管理、数据库管理、文件管理、Web 终端、容器管理、应用商店、备份恢复、计划任务、防火墙、SSH 安全加固、告警通知等 14+ 功能模块，帮助您轻松管理 Linux 服务器。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.22+ / Gin / GORM / SQLite |
| 前端 | Vue 3 / TypeScript / Naive UI / Pinia |
| 容器 | Docker API |
| 认证 | JWT RS256 / bcrypt / TOTP 二次验证 |
| 部署 | 单二进制 + 前端静态资源 |

## 快速导航

### 🚀 入门

- [快速开始](./quickstart.md) — 环境要求、安装部署、默认账号、常见问题

### 📖 用户手册

- [用户手册索引](./user-guide/README.md)
- [仪表盘使用](./user-guide/dashboard.md) — 系统概览、资源监控、审计日志
- [网站管理](./user-guide/website.md) — 站点创建、Nginx 配置、SSL 证书
- [数据库管理](./user-guide/database.md) — MySQL/PostgreSQL/Redis 实例与库表管理
- [文件管理器](./user-guide/file.md) — 文件浏览、上传下载、压缩解压
- [Web 终端](./user-guide/terminal.md) — 浏览器内 SSH 终端
- [容器管理](./user-guide/container.md) — Docker 容器与镜像管理
- [应用商店](./user-guide/appstore.md) — 一键安装常用应用
- [备份恢复](./user-guide/backup.md) — 数据备份与恢复
- [计划任务](./user-guide/cron.md) — 定时任务管理
- [防火墙](./user-guide/firewall.md) — 端口规则与安全建议
- [SSH 安全](./user-guide/ssh.md) — SSH 配置、密钥管理、Fail2ban
- [告警通知](./user-guide/alert.md) — 告警规则与通知渠道
- [用户与权限](./user-guide/users.md) — 用户管理、角色权限、RBAC
- [系统设置](./user-guide/settings.md) — 系统配置、SMTP、版本更新

### 🔌 API 文档

- [API 概览](./api/README.md) — 认证方式、统一响应、分页、错误码
- [认证 API](./api/auth.md) — 登录、登出、刷新令牌、用户信息
- [Dashboard API](./api/dashboard.md) — 概览、指标、审计日志
- [网站 API](./api/website.md) — 站点 CRUD、启停、配置、SSL、日志
- [数据库 API](./api/database.md) — 实例、库表、用户、SQL 执行
- [文件 API](./api/file.md) — 文件浏览、读写、上传下载、压缩
- [终端 API](./api/terminal.md) — WebSocket 终端协议
- [容器 API](./api/container.md) — 容器列表、启停、日志、统计、镜像
- [应用商店 API](./api/appstore.md) — 应用列表、安装卸载、状态
- [备份 API](./api/backup.md) — 备份 CRUD、恢复、下载
- [计划任务 API](./api/cron.md) — 任务 CRUD、启停、触发、日志
- [防火墙 API](./api/firewall.md) — 规则 CRUD、切换、安全建议
- [SSH API](./api/ssh.md) — SSH 配置、密钥、登录日志、Fail2ban
- [告警 API](./api/alert.md) — 规则、渠道、历史记录
- [用户/角色 API](./api/users.md) — 用户管理、角色权限、权限树
- [系统设置 API](./api/system.md) — 设置、版本、清理、SMTP、系统信息

### 🛠 开发文档

- [架构设计](./dev/architecture.md) — 分层架构、目录结构、数据流、安全设计
- [插件开发指南](./dev/plugin-dev.md) — 插件接口、示例代码、扩展点
- [贡献指南](./dev/contributing.md) — 开发规范、提交流程、代码审查
- [设计规范](./dev/design.md) — 设计系统、Design Token、组件规范

## 默认信息

| 项目 | 值 |
|------|-----|
| 后端端口 | 8888 |
| 前端开发端口 | 5173 |
| 默认账号 | admin |
| 默认密码 | admin123 |
| API 前缀 | `/api/v1` |
| 健康检查 | `GET /health` |

## 项目结构

```
fuwuqi/
├── backend/              # Go 后端
│   ├── cmd/agent/        # 入口 main.go
│   ├── internal/
│   │   ├── api/          # API Handler 层
│   │   ├── service/      # 业务逻辑层
│   │   ├── model/        # 数据模型层
│   │   ├── dto/          # 数据传输对象
│   │   ├── middleware/   # 中间件（JWT/CORS/限流/审计）
│   │   ├── common/       # 公共组件（响应/命令白名单）
│   │   └── config/       # 配置加载
│   └── config.yaml        # 配置文件
├── frontend/             # Vue3 前端
│   └── src/
│       ├── api/          # API 请求封装
│       ├── components/   # 通用组件
│       ├── views/        # 页面视图
│       ├── stores/       # Pinia 状态管理
│       ├── router/       # 路由配置
│       ├── layouts/      # 布局组件
│       └── locales/      # 国际化
├── docs/                 # 项目文档（本目录）
└── 云枢面板.md            # 产品需求文档
```
