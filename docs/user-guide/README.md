# 用户手册

本手册详细介绍云枢面板各功能模块的使用方法，帮助您快速上手并充分利用面板的各项能力。

## 模块索引

| 模块 | 文档 | 核心功能 |
|------|------|----------|
| 仪表盘 | [dashboard.md](./dashboard.md) | 系统概览、资源监控图表、审计日志 |
| 网站管理 | [website.md](./website.md) | 站点创建、Nginx 配置、SSL 证书、访问日志 |
| 数据库管理 | [database.md](./database.md) | MySQL/PostgreSQL/Redis 实例、库表、用户权限、SQL 执行 |
| 文件管理器 | [file.md](./file.md) | 文件浏览、上传下载、编辑、压缩解压、搜索 |
| Web 终端 | [terminal.md](./terminal.md) | 浏览器内 SSH 终端、会话管理 |
| 容器管理 | [container.md](./container.md) | Docker 容器列表、启停、日志、统计、镜像 |
| 应用商店 | [appstore.md](./appstore.md) | 一键安装常用应用、应用源管理 |
| 备份恢复 | [backup.md](./backup.md) | 网站/数据库/系统备份、恢复、下载 |
| 计划任务 | [cron.md](./cron.md) | 定时任务创建、执行日志、手动触发 |
| 防火墙 | [firewall.md](./firewall.md) | 端口规则管理、安全建议 |
| SSH 安全 | [ssh.md](./ssh.md) | SSH 配置、密钥管理、登录日志、Fail2ban |
| 告警通知 | [alert.md](./alert.md) | 告警规则、通知渠道、告警历史 |
| 用户与权限 | [users.md](./users.md) | 用户管理、角色权限、RBAC、TOTP |
| 系统设置 | [settings.md](./settings.md) | 系统配置、SMTP 邮件、版本更新、系统清理 |

## 快速上手流程

1. **登录面板** — 使用默认账号 `admin/admin123` 登录，首次登录后修改密码
2. **查看仪表盘** — 了解服务器 CPU、内存、磁盘、网络等资源状况
3. **创建网站** — 在网站管理中添加站点，配置域名和 Nginx
4. **安装数据库** — 通过数据库管理创建 MySQL/PostgreSQL/Redis 实例
5. **配置安全** — 设置防火墙规则、SSH 加固、启用告警通知
6. **定期备份** — 配置自动备份计划，确保数据安全

## 通用操作说明

### 分页列表

大部分列表页面支持分页，参数说明：

| 参数 | 说明 | 默认值 |
|------|------|--------|
| page | 页码，从 1 开始 | 1 |
| pageSize | 每页条数 | 10 |

### 操作确认

删除、停止等危险操作会弹出确认对话框，请仔细核对后再执行。

### 状态说明

各模块通用状态标识：

- `running` / `active` / `enabled` — 运行中/已启用
- `stopped` / `disabled` — 已停止/已禁用
- `error` / `failed` — 异常/失败
- `pending` — 等待中
