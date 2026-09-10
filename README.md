# 犀牛面板（RhinoPanel）

[![CI](https://github.com/USERNAME/REPO/actions/workflows/ci.yml/badge.svg)](https://github.com/USERNAME/REPO/actions/workflows/ci.yml)
[![Docker Build](https://github.com/USERNAME/REPO/actions/workflows/docker.yml/badge.svg)](https://github.com/USERNAME/REPO/actions/workflows/docker.yml)
[![CodeQL](https://github.com/USERNAME/REPO/actions/workflows/codeql.yml/badge.svg)](https://github.com/USERNAME/REPO/actions/workflows/codeql.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> 徽章链接中的 `USERNAME/REPO` 请替换为实际的 GitHub 仓库所有者和仓库名。

轻量级 Linux 服务器可视化管理平台，基于 Go + Vue3 + Docker 构建，类似 1Panel / 宝塔面板。

## 技术栈

- **后端**: Go 1.25+ / Gin / GORM / SQLite / gorilla/websocket / Docker SDK / gopsutil
- **前端**: Vue 3 + TypeScript / Naive UI / Vite / Pinia / ECharts / xterm.js / Monaco Editor
- **部署**: systemd + 安装脚本（目标 Linux）

## 项目结构

```
cloudpivot/
├── backend/                    # Go 后端
│   ├── cmd/agent/main.go       # 入口
│   ├── internal/
│   │   ├── api/                # HTTP Handler
│   │   ├── middleware/         # 中间件（CORS/JWT/限流/审计）
│   │   ├── model/              # GORM 数据模型
│   │   ├── service/            # 业务逻辑（auth/monitor/site/database/file/terminal）
│   │   ├── dto/                # 请求/响应结构体
│   │   ├── config/             # 配置加载
│   │   └── common/             # 公共工具（响应/安全命令执行）
│   ├── pkg/
│   │   ├── nginx/              # Nginx 配置生成引擎
│   │   ├── docker/             # Docker SDK 封装
│   │   ├── certbot/            # ACME 证书
│   │   └── cron/               # 定时任务
│   ├── config.yaml             # 配置文件
│   └── go.mod
├── frontend/                   # Vue3 前端
│   ├── src/
│   │   ├── api/                # API 请求封装
│   │   ├── components/         # 通用组件（PageCard/DataTable/StatusTag/CodeEditor...）
│   │   ├── layouts/            # 布局
│   │   ├── views/              # 页面
│   │   ├── stores/             # Pinia 状态
│   │   ├── router/             # 路由
│   │   └── styles/             # 设计 token + 全局样式
│   └── package.json
├── store/                      # 应用商店定义
├── scripts/                    # 构建与部署脚本
├── docker/                     # 开发环境 Docker
└── docs/                       # 文档
```

## 快速开始

### 环境要求

- Go 1.25+
- Node.js 18+ / npm 9+
- Docker（运行时，开发环境可选）

### 后端

```bash
cd backend
go mod tidy
go run cmd/agent/main.go
```

- 默认监听端口: **8888**
- 数据库: SQLite（自动创建到 `./data/cloudpivot.db`）
- JWT: RS256，首次启动自动生成 RSA 密钥对

### 前端

```bash
cd frontend
npm install
npm run dev
```

- 开发服务器: http://localhost:5173
- API 代理: `/api` → `http://localhost:8888`

### 默认账号

| 项目 | 值 |
|------|-----|
| 用户名 | `admin` |
| 密码 | `admin123` |

> 首次登录后请立即修改密码。

## API 概览

基础路径: `/api/v1`，统一响应体 `{code, message, data}`，认证方式 `Authorization: Bearer <token>`

| 模块 | 端点 |
|------|------|
| 认证 | `POST /auth/login`, `POST /auth/logout`, `POST /auth/refresh`, `GET /auth/info` |
| Dashboard | `GET /dashboard/overview`, `GET /dashboard/metrics`, `GET /dashboard/audit-logs` |
| 网站 | `GET/POST /sites`, `GET/PUT/DELETE /sites/:id`, `POST /sites/:id/{start,stop,restart}`, `GET/PUT /sites/:id/config`, `POST /sites/:id/ssl/apply` |
| 数据库 | `GET/POST /databases`, `POST /databases/:id/{start,stop,restart}`, `GET/POST /databases/:id/dbs`, `POST /databases/:id/sql`, `POST /databases/:id/import` |
| 文件 | `GET /files`, `POST /files/{mkdir,create,rename,copy,move,compress,decompress,upload}`, `PUT /files/content`, `GET /files/{content,download,search}`, `DELETE /files` |
| 终端 | `WS /ws/terminal` |
| 容器 | `GET /containers`, `POST /containers/:id/{start,stop,restart}`, `DELETE /containers/:id`, `GET /containers/:id/logs`, `GET /images` |
| 应用商店 | `GET /apps`, `GET /apps/:key`, `GET /apps/installed`, `POST /apps/:key/{install,uninstall}`, `GET /apps/:key/status` |
| 备份 | `GET/POST /backups`, `GET/DELETE /backups/:id`, `POST /backups/:id/restore`, `GET /backups/:id/download` |
| 计划任务 | `GET/POST /cron/tasks`, `PUT/DELETE /cron/tasks/:id`, `POST /cron/tasks/:id/{enable,disable,trigger}`, `GET /cron/tasks/:id/logs` |
| 防火墙 | `GET/POST /firewall/rules`, `DELETE /firewall/rules/:id`, `POST /firewall/rules/:id/toggle`, `GET /firewall/suggestions` |
| 监控 | `GET /monitor/{processes,disk,network,history}` |
| 审计 | `GET /audit/logs`, `GET /audit/login-logs`, `GET /audit/export` |
| 系统 | `GET/PUT /system/settings`, `GET /system/version`, `GET /system/info`, `POST /system/clean` |

## Phase 1 MVP 模块

- [x] 项目骨架（monorepo 目录结构）
- [x] 后端基础框架（配置/SQLite/GORM/统一响应/中间件/安全命令执行）
- [x] 认证模块（JWT RS256/bcrypt/初始管理员/登录限流10次/分）
- [x] Dashboard（系统信息/CPU内存磁盘网络概览/15秒实时采集/24h趋势/操作日志）
- [x] 网站管理（静态/运行环境/反向代理 CRUD/Nginx配置引擎/SSL证书/启停控制）
- [x] 数据库管理（MySQL Docker部署/库表CRUD/用户权限/SQL执行器/导入导出）
- [x] 文件管理器（目录浏览/文件操作/Monaco在线编辑/上传下载/zip/tar.gz压缩解压）
- [x] Web 终端（xterm.js + WebSocket SSH/多标签页/PTY/断线重连）
- [x] 前端框架（登录页/主布局/路由/Pinia/亮暗主题/API封装/9个通用组件）
- [x] 前端业务页面（Dashboard/网站管理/SSL证书/数据库管理/文件管理器/Web终端）

## Phase 2 核心增强模块

- [x] 容器管理（Docker SDK封装/容器CRUD/实时日志/镜像管理/Compose支持）
- [x] 应用商店（10个内置应用定义/分类浏览搜索/一键安装引导/已安装管理/环境参数动态表单）
- [x] 备份恢复（网站/数据库/系统三类备份/异步执行/恢复前快照/本地存储/下载）
- [x] 计划任务（robfig/cron引擎/可视化cron配置/4种任务类型/执行日志/手动触发）
- [x] 防火墙管理（端口/IP规则/iptables+firewalld白名单/高危端口安全建议）
- [x] 监控增强（Top N进程管理/磁盘分析/网络连接统计/ECharts历史趋势）
- [x] 审计日志（操作日志多条件筛选/登录日志/CSV导出）
- [x] 系统设置（面板配置/SMTP邮件/版本信息/清理工具/亮暗主题）
- [x] 前端全部 14 个页面功能完整（含计划任务新页面）

## Phase 3 企业特性

- [x] Redis 管理（Key浏览器五类型/在线CLI/内存分析/大Key扫描）
- [x] PostgreSQL 支持（实例部署/库表管理/用户权限/SQL执行器/通用数据库引擎接口）
- [x] 多用户 RBAC 权限（3内置角色+自定义角色/45个权限点/中间件鉴权/JWT权限注入）
- [x] TOTP 二次验证（Google Authenticator兼容/二维码绑定/8个恢复码/登录流程扩展）
- [x] SSH 安全加固（配置可视化/authorized_keys管理/登录日志/Fail2ban集成）
- [x] 告警通知（规则引擎/6种通知渠道/后台60s轮询/触发恢复检测/防抖）
- [x] 安装脚本（install.sh一键安装/uninstall.sh卸载/systemd部署/CentOS/Ubuntu/Debian）

## Phase 4 生态与打磨

- [x] 多语言支持（vue-i18n/中/英/日三语/核心页面国际化/顶栏切换/localStorage持久化）
- [x] 应用商店多源（官方/社区/私有三类型/源CRUD/同步/应用来源标记）
- [x] 云厂商 DNS API（阿里云/腾讯云DNSPod/Cloudflare/抽象接口/SSL DNS-01验证）
- [x] 插件系统架构（Plugin接口/扩展点机制/init自动注册/2个示例插件/前端插件管理页）
- [x] 项目文档站（37篇Markdown/快速开始/用户手册15篇/API文档16篇/开发文档4篇）

## 设计特色

- **原创设计语言**: "深夜数据中心" — 深墨蓝灰基底 + 琥珀撞色 + 青绿状态色
- **签名元素**: 状态指示灯（Status LED）— 机柜指示灯灵感，正常青绿脉冲/警告琥珀/异常红闪烁
- **亮/暗双主题**: CSS 变量 + Naive UI darkTheme 同步切换
- **组件复用**: 9 个通用组件（PageCard/DataTable/SearchForm/ModalForm/StatusTag/StatCard/ConfirmDialog/CodeEditor），页面零重复实现
- **多语言**: 中/英/日三语切换，核心页面国际化，语言偏好持久化
- **插件化**: 标准 Plugin 接口 + 扩展点机制，支持运行时注册菜单/API/Dashboard/通知

## 安全设计

- JWT RS256 非对称签名，RSA 密钥对自动生成
- bcrypt 密码哈希（cost=10）
- 登录限流 10次/分钟（基于 IP 令牌桶）
- 系统命令白名单执行，禁止 shell 元字符注入
- SQL 参数化查询，危险操作默认拦截
- 文件路径穿越防护（safePath 校验）
- 操作审计日志全量记录
- TOTP 二次验证（可选），恢复码一次性使用
- RBAC 细粒度权限控制，超级管理员/管理员/普通用户三级内置角色
- DNS API 凭证 base64 编码存储，API 响应永不返回
- SSH 配置修改前 sshd -t 语法校验，密钥格式校验

## Docker 安装

支持通过 Docker 镜像快速部署 CloudPivot，详见 [Docker 安装指南](docs/docker-install.md)。

## 文档

完整文档位于 [docs/](docs/) 目录，包含：

- [快速开始](docs/quickstart.md)
- [用户手册](docs/user-guide/README.md) — 15 篇功能使用指南
- [API 文档](docs/api/README.md) — 16 篇接口参考
- [开发文档](docs/dev/README.md) — 架构、设计、插件开发、贡献指南

## 许可证

本项目基于 [MIT License](LICENSE) 开源，版权所有 © 2026 CloudPivot。
