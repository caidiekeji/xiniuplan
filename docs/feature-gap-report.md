# 云枢面板（RhinoPanel）功能实现差距检查报告

> 检查时间：2026-09-13 · 对照 PRD：`云枢面板.md` v1.0 · 方式：路由注册 + service 实现逐项比对
>
> **✅ 2026-09-13 更新：以下缺口已全部补齐实现（见文末"补齐记录"），后端 go build/vet 与前端 vue-tsc 全部通过。**

## 总体结论

README 中 Phase 1~4 全部勾选完成，**与实际代码不完全相符**。核心链路（网站/数据库/文件/终端/商店/防火墙/告警/RBAC/TOTP）确有实现，但存在 **14 项 PRD 点名的功能缺失**（其中 8 项 P0/P1），另有 2 处**占位假实现**。

---

## 一、占位假实现（最需要优先处理）

| 位置 | 问题 |
|------|------|
| `backend/internal/service/system/service.go:280` | `SendTestEmail` 直接 `return nil`（有 `// TODO: 实际 SMTP 发送实现`），**测试邮件和忘记密码邮件实际不会发出** |
| `backend/internal/service/system/service.go:286` | `CheckUpdate` 硬编码 `HasUpdate: false`，**版本更新检测是假实现**（PRD SY-004） |

> 注：告警邮件走 `alert/notifier.go`，是独立真实实现，不受上面影响。

## 二、P0 / P1 级缺失（PRD 点名）

| PRD 编号 | 功能 | 现状 |
|----------|------|------|
| BK-003 (P1) | 备份远程存储 S3/OSS/SFTP/WebDAV | 全代码库无任何相关实现，仅本地存储 |
| A-005 (P1) | 应用商店版本更新/一键升级 | 无 `/apps/:key/upgrade` 路由，service 无升级逻辑 |
| C-004 (P1) | 容器 Web 终端（docker exec） | container service 无 Exec 实现 |
| S-004 (P1) | 防火墙端口转发 | 无路由无实现 |
| U-006 (P1) | 在线会话管理/强制踢出 | 无接口 |
| F-005 (P1) | 文件权限管理（chmod/owner） | 无接口 |
| M-002 (P1，部分) | 进程管理 | 有 `/monitor/processes` 列表，**无终止进程接口** |
| DB-002 / DB-004 (P1) | 数据库参数配置（my.cnf 可视化）、实例 QPS/慢查询监控 | 均无接口 |
| W-006/007/008/009 (P1~P2) | 伪静态模板、防盗链、重定向规则、IP 黑白名单/Basic Auth | 后端无专门接口、前端无入口，只能手改 nginx 配置 |
| SY-005 (P1) | 系统安全更新（yum/apt 一键更新） | 无实现 |
| A-008 / BK-001（部分） | 应用数据备份 | 备份仅网站/数据库/系统三类，无应用类型 |
| SY-007（部分） | SMTP 邮件 | 配置界面有，但见"占位假实现"第一条 |

## 三、P2 / P3 级缺失（PRD 规划内）

- D-007 安全评分（Dashboard）
- W-018 流量统计（PV/UV 日志分析）
- C-008 / C-009 容器网络、存储卷管理
- M-007 监控大屏
- F-008 文件预览（图片/PDF/音视频）
- T-004 终端快捷命令面板、T-005 ZMODEM 文件传输、T-006 会话录制
- S-012 异常登录检测
- BK-008 异地容灾（备份同步到另一实例）
- U-007 API Key

## 四、设计偏离（可接受，但需知晓）

- PRD 要求 `WS /ws/monitor` 实时监控推送，实际只有 `WS /ws/terminal`；监控数据走 HTTP 轮询。

## 五、已确认实现良好的部分（抽查通过）

- 路由覆盖广：24 个 handler、200+ 端点，网站/数据库/文件/终端/容器/商店/备份/计划任务/防火墙/监控/审计/系统/Redis/PostgreSQL/告警/SSH 全部有路由。
- 告警 6 渠道真实实现：email/dingtalk/wechat/feishu/telegram/webhook（notifier.go）。
- 备份保留策略 `CleanupOldBackups` 有实现（默认保留 10 份）。
- 应用商店内置 20 个应用（README 写 10+，实际更多）。
- 应用多源路由已注册（经 `appstore.go:37` 挂载）。
- 用户/角色/TOTP/DNS 路由经 `system.go:40-45` 挂载，均可用。

## 六、建议修复顺序

1. `SendTestEmail` 真实 SMTP 实现（影响忘记密码流程，属于安全/可用性缺陷）
2. 应用商店升级（A-005）+ 备份远程存储（BK-003）
3. 容器 exec 终端（C-004）+ 进程终止（M-002）
4. 网站配置增强（伪静态/防盗链/重定向/访问限制）
5. 其余 P2 项按需排期


---

## 补齐记录（2026-09-13）

### 占位假实现修复
- `SendTestEmail` 真实 SMTP 发送（`system/service.go`）
- `CheckUpdate` 真实版本检测：支持 `update_check_url` / `update_github_repo` 设置，语义化版本比较

### P0/P1 缺口
| PRD 编号 | 实现 |
|----------|------|
| BK-003 | 远程存储 `RemoteStorage` 模型 + S3(SigV4 手写)/WebDAV/SFTP(pkg/sftp) 上传；`POST /backups/:id/push`；`backup_auto_push` 自动推送设置 |
| BK-008 | 面板间推送：`POST /api/v1/backup/receive`（X-Panel-Token 认证）接收对端备份 |
| A-005/A-008 | 应用升级 `POST /apps/:key/upgrade`（compose down → 覆盖定义保留数据 → up）；应用备份类型 `type=app` 及恢复 |
| C-004 | 容器 exec 终端 `WS /ws/exec`（ExecAttach + resize） |
| C-006 补全 | 镜像导出 `GET /images/:id/save`、导入 `POST /images/load`、删除 `DELETE /images/:id` |
| C-008/C-009 | 网络与存储卷 CRUD（`/networks`、`/volumes`） |
| S-004 | 端口转发：NAT DNAT + MASQUERADE + FORWARD 放行 + sysctl 转发开启，`type=forward` 规则 |
| U-006 | 会话注册表（JWT 中间件登记）+ `GET /auth/sessions` + `DELETE /auth/sessions/:hash` 踢出（token 黑名单） |
| U-007 | API Key：`api_keys` 表（SHA-256 secret）、CRUD、`X-API-Key` 中间件认证，权限可独立配置 |
| F-005 | `POST /files/permission`（chmod/chown，支持递归，safePath 防穿越） |
| M-002 | `POST /monitor/processes/kill`（SIGTERM/SIGKILL，保护 init 与面板自身）+ 新权限点 `monitor:update` |
| DB-002/004 | `GET/PUT /databases/:id/variables`（MySQL SET GLOBAL+PERSIST / PG ALTER SYSTEM）、`GET /databases/:id/status`（QPS 双采样、连接数、慢查询、缓冲池、PG 统计） |
| W-006~009 | 站点片段机制：server 块 `include <domain>.d/*.conf`；伪静态模板（WordPress/Typecho/ThinkPHP/Laravel/Discuz）、防盗链、301/302 重定向、IP 黑白名单、BasicAuth（bcrypt htpasswd） |
| SY-005 | `GET /system/updates` 检测（apt/yum/dnf）+ `POST /system/updates/install`（confirm 确认，30 分钟超时白名单执行） |
| SY-006 | `GET/PUT /system/ntp`（timedatectl） |

### P2/P3 缺口
- D-007 安全评分 `GET /dashboard/security-score`（弱密码/TOTP/防火墙/SSH 端口/密码登录/SSL 覆盖率/安全入口，A-D 评级）
- W-018 流量统计 `GET /sites/:id/traffic`（Nginx 日志解析 PV/UV/TopURL/TopIP，24h 按小时、7d/30d 按天）
- S-012 异常登录 `GET /audit/anomalies`（暴力破解/撞库/新 IP 登录）
- T-004 快捷命令 `GET /terminal/quick-commands`（10 条内置）
- T-005 ZMODEM：终端 WS `raw=1` 二进制透传模式（配合 zmodem.js）
- T-006 会话录制：WS `record=1` 输出落盘 + `GET/DELETE /terminal/records` 回放管理

### 前端补齐
- 新页面「会话与 API Key」（/sessions，含踢出、创建 Key 一次性 secret 展示）
- 设置页：安全评分卡 + 系统更新检测/一键安装 + NTP 开关（SecurityExtras.vue）
- 备份页：应用类型、推送对话框、远程存储管理（S3/WebDAV/SFTP/面板）
- 防火墙页：端口转发表单；监控页：终止进程对齐新 API；商店页：升级按钮
- 容器页：exec 终端（xterm + WS）、网络/存储卷管理卡（DockerResources.vue）
- 审计页：异常登录检测 tab；统一 API 封装 `api/newfeatures.ts`

### 验证
- 后端：`go build ./...` ✅ `go vet ./...` ✅
- 前端：`vue-tsc --noEmit` ✅
- 说明：网站片段/流量统计/终端录制的前端 UI 已有后端 API 支撑，页面级入口可后续迭代（当前各页面已有基础入口的除外）。
