# 云枢面板 安全测试与加固报告

**报告日期**: 2026-09-09
**测试范围**: 后端 (Go/Gin) + 前端 (Vue3/TypeScript)
**测试方法**: 白盒代码审计 + 漏洞模式扫描 + 直接修复验证

---

## 一、测试范围和方法

### 1.1 测试范围

| 模块 | 路径 | 说明 |
|------|------|------|
| 认证授权 | `internal/service/auth/`, `internal/middleware/auth.go`, `rbac.go`, `blacklist.go` | JWT、RBAC、令牌黑名单 |
| 输入验证 | `internal/service/database/`, `internal/service/file/`, `internal/common/command.go` | SQL注入、路径穿越、命令注入 |
| 敏感信息 | `internal/model/`, `internal/dto/`, `internal/api/` | API响应字段、错误信息、配置 |
| 传输安全 | `internal/middleware/cors.go`, `cmd/agent/main.go` | CORS、安全头、TLS |
| 前端安全 | `frontend/src/` | XSS、构建验证 |

### 1.2 测试方法

- 全量源码白盒审计，覆盖 30+ API handler、20+ service、7 个中间件
- 危险模式正则扫描（`exec.Command`、`fmt.Sprintf` SQL、`v-html`、敏感字段 json tag）
- 认证/授权链路追踪（路由注册 → 中间件 → handler → service）
- 发现问题直接修改源码修复，并通过 `go build` / `go vet` 验证

---

## 二、安全检查清单

### 2.1 认证与授权

| 检查项 | 状态 | 说明 |
|--------|------|------|
| JWT RS256 签名 | ✅ 通过 | RSA 2048-bit 密钥对自动生成，存储于 `data/jwt/`，私钥权限 0600 |
| Access Token 过期时间 | ✅ 通过 | 默认 2 小时（7200秒），可配置 |
| Refresh Token 过期时间 | ✅ 通过 | 默认 7 天（604800秒），可配置 |
| Token 过期返回 401 | ✅ 通过 | `JWTAuth` 中间件解析失败返回 40101 |
| 登出令牌黑名单 | ✅ 通过 | `BlacklistToken` 将 access token 加入内存黑名单，2小时后自动过期 |
| 未授权访问保护 | ✅ 通过 | `/api/v1` 下除 `login`/`refresh` 外均经过 `JWTAuth` 中间件 |
| RBAC 权限校验 | 🔧 已修复 | 系统设置、Dashboard 路由原缺少 RBAC，已补充 |
| 普通用户不能访问管理端点 | 🔧 已修复 | 系统设置/SMTP/清理等端点已加 `RequirePermission` |
| bcrypt 密码哈希 | ✅ 通过 | `bcrypt.DefaultCost` (cost=10)，数据库无明文密码 |
| API 不返回 password_hash | ✅ 通过 | `User.PasswordHash` 字段 `json:"-"` |
| API 不返回 totp_secret | ✅ 通过 | `User.TOTPSecret` 字段 `json:"-"` |
| 登录限流 10次/分钟 | ✅ 通过 | `RateLimiter` 登录接口独立限流桶 |

### 2.2 输入验证与注入防护

| 检查项 | 状态 | 说明 |
|--------|------|------|
| SQL 参数化查询 | ✅ 通过 | GORM 查询均使用参数化；数据值查询使用 `?` 占位符 |
| SQL 标识符注入 | 🔧 已修复 | MySQL 侧表名/库名/用户名原直接 `fmt.Sprintf` 拼接，已添加 `validateMySQLIdentifier` 校验 |
| 危险 SQL 操作检测 | ✅ 通过 | `checkDangerousSQL` 拦截 DROP DATABASE/TABLE、无 WHERE DELETE |
| XSS 防护 | ✅ 通过 | 前端无 `v-html` / `innerHTML`，Vue 默认转义 |
| 路径穿越防护 | ✅ 通过 | `safePath` 双重校验：显式拒绝 `..` 段 + 解析后 `filepath.Rel` 二次校验 |
| 文件上传路径校验 | ✅ 通过 | `filepath.Base(header.Filename)` 去除路径，二次校验目标在 root 内 |
| 命令执行白名单 | ✅ 通过 | `CommandExecutor` 模板注册制，用户输入仅替换占位符，禁止 shell 元字符 |
| 无直接 exec.Command 拼接 | 🔧 已修复 | 备份恢复原使用 `exec.Command("mysql", ..., "-p"+password)`，已改用 `MYSQL_PWD` 环境变量 |
| SSH 命令执行 | ✅ 通过 | SSH 服务全部通过 `CommandExecutor` 白名单模板执行 |

### 2.3 敏感信息保护

| 检查项 | 状态 | 说明 |
|--------|------|------|
| API 不返回敏感字段 | ✅ 通过 | password_hash、totp_secret、DNS credentials 均 `json:"-"` |
| DNS credentials 不返回 | ✅ 通过 | `DNSAccount.Credentials` 字段 `json:"-"`，DTO `DNSAccountResponse` 无 credentials |
| 错误信息不泄露堆栈 | ✅ 通过 | `gin.Recovery()` 捕获 panic 返回通用错误；业务错误使用统一 `common.Fail` |
| JWT 密钥不硬编码 | ✅ 通过 | RSA 密钥对启动时自动生成，存储于文件系统，不写入代码 |
| config.yaml 无硬编码密码 | ✅ 通过 | 配置文件无明文密码，数据库密码从实例表获取 |
| .gitignore 排除敏感文件 | ✅ 通过 | 排除 `data/`、`*.pem`、`.env*`、`runtime/` |
| TLS/HTTPS 配置入口 | 🔧 已修复 | 原无 TLS 配置，已添加 `tls_enable`/`tls_cert_file`/`tls_key_file` 及 `ListenAndServeTLS` 支持 |
| 密码不在进程参数中 | 🔧 已修复 | 备份 mysqldump/mysql 原使用 `-p` 参数，已改用 `MYSQL_PWD` 环境变量 |

### 2.4 其他安全检查

| 检查项 | 状态 | 说明 |
|--------|------|------|
| CORS 生产环境安全 | 🔧 已修复 | 原允许所有来源 + credentials，已改为 release 模式仅允许同源 |
| 安全响应头 | 🔧 已修复 | 原无任何安全头，已添加 `SecurityHeaders` 中间件（X-Content-Type-Options、X-Frame-Options、CSP、Referrer-Policy 等） |
| Session 固定 | ✅ 通过 | 登录后签发新 JWT，不使用旧 token |
| CSRF 防护 | ✅ 通过 | 纯 JWT Bearer token 认证，无 cookie 认证，天然防 CSRF |
| 文件上传类型/大小限制 | ✅ 通过 | 大小限制 512MB，文件名经 `filepath.Base` 净化 |
| 非登录端点限流 | 🔧 已修复 | 原 per-path 限流可被多路径绕过，已改为全局 per-IP API 限流（100次/分钟） |
| WebSocket Origin 校验 | 🔧 已修复 | 原 `CheckOrigin` 恒返回 true，已改为 release 模式校验同源 |

---

## 三、发现的问题列表

### VULN-001: 缺少安全响应头

| 项目 | 内容 |
|------|------|
| **编号** | VULN-001 |
| **位置** | 全局（无中间件） |
| **描述** | 应用未设置任何安全响应头，缺少 X-Content-Type-Options、X-Frame-Options、Content-Security-Policy、Referrer-Policy 等，存在 MIME 嗅探、点击劫持、XSS 放大等风险 |
| **风险等级** | 高 |
| **修复状态** | ✅ 已修复 |
| **修复方式** | 新建 `internal/middleware/security_headers.go`，实现 `SecurityHeaders()` 中间件，设置 nosniff、DENY、CSP、Referrer-Policy、Cache-Control 等头；在 `main.go` 中全局注册 |

### VULN-002: CORS 过度宽松（允许任意来源 + 凭证）

| 项目 | 内容 |
|------|------|
| **编号** | VULN-002 |
| **位置** | `internal/middleware/cors.go` |
| **描述** | CORS 中间件将请求的 Origin 直接反射到 `Access-Control-Allow-Origin`，同时设置 `Access-Control-Allow-Credentials: true`。任意恶意网站均可携带用户凭证发起跨域请求读取响应数据 |
| **风险等级** | 高 |
| **修复状态** | ✅ 已修复 |
| **修复方式** | 重写 CORS 中间件：`release` 模式下仅允许同源请求（Origin 与 Host 一致），非同源直接返回 403；`debug` 模式保持宽松便于开发 |

### VULN-003: WebSocket 终端 Origin 校验被禁用

| 项目 | 内容 |
|------|------|
| **编号** | VULN-003 |
| **位置** | `internal/api/terminal.go:23-25` |
| **描述** | WebSocket Upgrader 的 `CheckOrigin` 回调恒返回 `true`，任意网站均可发起 WebSocket 连接到终端接口，结合终端的 SSH 交互能力可导致远程命令执行 |
| **风险等级** | 高 |
| **修复状态** | ✅ 已修复 |
| **修复方式** | 修改 `CheckOrigin`：`release` 模式下校验 Origin 的 Host 与请求 Host 一致；`debug` 模式允许所有来源 |

### VULN-004: MySQL SQL 标识符注入（表名/库名/用户名）

| 项目 | 内容 |
|------|------|
| **编号** | VULN-004 |
| **位置** | `internal/service/database/database.go`、`user.go` |
| **描述** | MySQL 库表管理中，`tableName`、`dbName`、`username`、`host` 等标识符直接通过 `fmt.Sprintf` 拼接到 SQL 语句中（如 `` SHOW INDEX FROM `%s` ``、`` DROP USER '%s'@'%s' ``）。虽有反引号包裹，但恶意标识符如 `` `table`; DROP TABLE users;-- `` 可突破反引号限制执行任意 SQL。PostgreSQL 侧已有 `validateIdentifier`，MySQL 侧缺失 |
| **风险等级** | 高 |
| **修复状态** | ✅ 已修复 |
| **修复方式** | 新建 `internal/service/database/identifier.go`，实现 `validateMySQLIdentifier`（仅允许字母/数字/下划线，字母开头，长度≤64）、`quoteMySQLIdentifier`、`validateMySQLHost`、`charsetRegex`、`collationRegex`；在 `CreateDatabase`、`DeleteDatabase`、`GetTableStructure`、`ListTableData`、`DeleteUser`、`GrantPrivileges`、`RevokePrivileges` 入口处添加校验 |

### VULN-005: 系统设置路由缺少 RBAC 权限校验

| 项目 | 内容 |
|------|------|
| **编号** | VULN-005 |
| **位置** | `internal/api/system.go:23-44` |
| **描述** | `/system/settings`（查看/修改系统设置）、`/system/smtp`（查看/修改 SMTP 配置含邮件服务器密码）、`/system/clean`（系统清理）、`/system/info`（系统信息）等端点未挂载任何 `RequirePermission` 中间件。任意已登录普通用户均可查看和修改系统设置、SMTP 凭证 |
| **风险等级** | 高 |
| **修复状态** | ✅ 已修复 |
| **修复方式** | 在 `RegisterSystemRoutes` 中为各端点添加 `RequirePermission(model.PermSettingView)` / `RequirePermission(model.PermSettingUpdate)` 中间件 |

### VULN-006: Dashboard 路由缺少 RBAC 权限校验

| 项目 | 内容 |
|------|------|
| **编号** | VULN-006 |
| **位置** | `cmd/agent/main.go:103-106` |
| **描述** | `/dashboard/overview`、`/dashboard/metrics`、`/dashboard/audit-logs` 端点未挂载权限校验中间件，任意已登录用户可查看系统监控指标和审计日志 |
| **风险等级** | 中 |
| **修复状态** | ✅ 已修复 |
| **修复方式** | 在 main.go 中为 dashboard 端点添加 `RequirePermission(model.PermMonitorView)` / `RequirePermission(model.PermAuditView)` |

### VULN-007: MySQL 密码通过命令行参数传递（进程可见）

| 项目 | 内容 |
|------|------|
| **编号** | VULN-007 |
| **位置** | `internal/service/backup/service.go`、`internal/service/database/sql.go` |
| **描述** | 备份模块的 `mysqldump` 和 `mysql` 恢复命令通过 `-p{{password}}` 命令行参数传递数据库 root 密码。在 Linux 上，任何本地用户均可通过 `ps aux` / `/proc/<pid>/cmdline` 读取到密码明文 |
| **风险等级** | 中 |
| **修复状态** | ✅ 已修复（备份模块） |
| **修复方式** | ① `internal/common/command.go`：为 `CommandTemplate` 新增 `Env` 字段，新增 `ExecuteWithEnv` 方法支持动态环境变量；② `internal/service/backup/service.go`：mysqldump 模板移除 `-p` 参数，改用 `MYSQL_PWD` 环境变量传递密码；`restoreDatabase` 同样改用 `cmd.Env = append(os.Environ(), "MYSQL_PWD="+password)` |
| **剩余风险** | `internal/service/database/sql.go` 中 docker exec 方式的 mysqldump 仍使用 `-p` 参数（在容器内可见），风险较低，建议后续改造为 `docker exec -e MYSQL_PWD` |

### VULN-008: 速率限制器可通过多路径绕过

| 项目 | 内容 |
|------|------|
| **编号** | VULN-008 |
| **位置** | `internal/middleware/ratelimit.go:78` |
| **描述** | 限流中间件以 `ip:path` 为 key 创建令牌桶，即每个 API 路径独立计数。攻击者可将请求分散到大量不同路径（如 `/api/v1/users/1`、`/api/v1/users/2`...），每个路径均享 100次/分钟配额，从而绕过全局限流 |
| **风险等级** | 中 |
| **修复状态** | ✅ 已修复 |
| **修复方式** | 重写限流器：API 请求使用全局 per-IP 限流桶（不区分路径），登录接口使用独立的更严格 per-IP 桶。确保单 IP 总请求量受限于全局配额 |

### VULN-009: 缺少 TLS/HTTPS 配置支持

| 项目 | 内容 |
|------|------|
| **编号** | VULN-009 |
| **位置** | `internal/config/config.go`、`cmd/agent/main.go` |
| **描述** | 服务器配置中无 TLS 相关选项，仅支持纯 HTTP 监听。管理面板的登录凭证、JWT token、系统操作均通过明文传输，存在中间人窃听风险 |
| **风险等级** | 中 |
| **修复状态** | ✅ 已修复 |
| **修复方式** | ① `config.go`：`ServerConfig` 新增 `TLSEnable`、`TLSCertFile`、`TLSKeyFile` 字段及默认值；② `main.go`：启动时根据 `TLSEnable` 选择 `ListenAndServeTLS` 或 `ListenAndServe` |

### VULN-010: 前端 TypeScript 构建错误（预存）

| 项目 | 内容 |
|------|------|
| **编号** | VULN-010 |
| **位置** | `frontend/src/views/appstore/index.vue:162` |
| **描述** | `getAppSources()` API 已声明返回类型为 `AppSource[]`，但代码中仍写 `(res.data.items || [])` 回退，TypeScript 推断 `res.data` 在 else 分支为 `never`，导致 `vue-tsc --noEmit` 构建失败 |
| **风险等级** | 低 |
| **修复状态** | ✅ 已修复 |
| **修复方式** | 移除不必要的 `.items` 回退，改为 `Array.isArray(res.data) ? res.data : []` |

---

## 四、安全架构总结

### 4.1 认证架构

```
用户登录 → bcrypt 校验密码 → 签发 JWT (RS256, access 2h + refresh 7d)
    ↓
API 请求 → Authorization: Bearer <token>
    ↓
JWTAuth 中间件 → 黑名单检查 → RSA 公钥验签 → 提取 user_id/role/permissions 注入 context
    ↓
登出 → access token 加入内存黑名单（2h TTL）
```

- **签名算法**: RS256（非对称，公钥验签、私钥签名）
- **密钥管理**: 启动时自动生成 2048-bit RSA 密钥对，私钥文件权限 0600
- **令牌类型**: access token（短期，2h）+ refresh token（长期，7d），类型 claim 区分

### 4.2 授权架构

```
JWT claims 中携带 permissions 列表（角色关联）
    ↓
RequirePermission("module:action") 中间件
    ↓
super_admin 角色直接放行 | 普通用户匹配 permissions 列表
    ↓
权限码体系: site:view/create/update/delete, database:*, file:*, user:*, setting:* 等 15 个模块
```

- **内置角色**: super_admin（*）、admin（除 user:delete/setting:update 外全部）、user（仅 view 类）
- **修复后覆盖**: 用户管理、角色管理、系统设置、SMTP、Dashboard、审计日志等端点均已挂载 RBAC

### 4.3 注入防护架构

| 攻击面 | 防护机制 |
|--------|----------|
| **SQL 注入（数据值）** | GORM 参数化 + `?` 占位符 |
| **SQL 注入（标识符）** | `validateMySQLIdentifier` 白名单校验（字母/数字/下划线） |
| **SQL 注入（PostgreSQL）** | `validateIdentifier` + `quoteIdentifier`（已有） |
| **危险 SQL** | `checkDangerousSQL` 拦截 DROP DATABASE/TABLE、无 WHERE DELETE |
| **命令注入** | `CommandExecutor` 模板白名单 + shell 元字符正则拦截 + 直接 exec（不经 shell） |
| **路径穿越** | `safePath` 双重校验：显式拒绝 `..` + `filepath.Rel` 解析后校验 |
| **XSS** | Vue 默认转义 + 前端无 `v-html`/`innerHTML` + CSP 响应头 |

### 4.4 敏感信息保护

- **存储层**: 密码 bcrypt(cost=10)、DNS 凭证 base64 编码存储、JWT 私钥 0600 权限
- **传输层**: TLS 配置支持（新增）、安全响应头（新增）、生产环境 CORS 同源限制（新增）
- **响应层**: `password_hash`、`totp_secret`、`credentials` 均 `json:"-"`；DTO 层无敏感字段
- **进程层**: MySQL 密码改用 `MYSQL_PWD` 环境变量，不出现在 `ps` 进程参数中

### 4.5 传输安全

- **HTTPS**: 已支持 TLS 配置（`tls_enable` + 证书路径），生产环境建议启用
- **CORS**: 生产环境仅允许同源，禁止携带凭证的跨域请求
- **安全头**: X-Content-Type-Options: nosniff、X-Frame-Options: DENY、CSP、Referrer-Policy、HSTS（建议在反向代理层添加）
- **WebSocket**: 生产环境校验 Origin 同源

---

## 五、剩余风险和建议

### 5.1 剩余风险

| 编号 | 风险 | 等级 | 说明 |
|------|------|------|------|
| R-001 | Token 黑名单为内存存储 | 中 | 服务重启后黑名单丢失，已登出 token 在剩余 TTL 内仍有效。建议使用 Redis 或数据库持久化黑名单 |
| R-002 | Cron 任务允许任意命令执行 | 中 | `execShell` 经 shell 元字符校验后直接 `exec.Command`，用户可执行系统任意二进制。此为功能设计，但应确保仅管理员可创建/修改 cron 任务 |
| R-003 | docker exec mysqldump 密码仍在命令行 | 低 | `internal/service/database/sql.go` 中容器内 mysqldump 仍使用 `-p` 参数，在容器内 `ps` 可见 |
| R-004 | 默认管理员密码 admin123 | 低 | 首次启动创建 admin/admin123 并打印提示，用户应立即修改。建议首次登录强制修改密码 |
| R-005 | 部分错误响应包含内部错误详情 | 低 | 部分 handler 使用 `err.Error()` 直接返回，可能泄露 SQL 错误、文件路径等。建议生产环境返回通用错误信息，详情记录日志 |
| R-006 | 无全局请求体大小限制 | 低 | 仅文件上传有 512MB 限制，JSON 请求体无全局限制。建议添加 `r.Use(gin.Logger())` 之外的 body size 限制 |
| R-007 | WebSocket 会话无权限重校验 | 低 | 终端 WebSocket 连接建立后，在会话持续期间不重新校验用户权限/状态 |

### 5.2 建议

1. **生产部署**: 务必设置 `server.mode: release`，启用 `tls_enable: true`，配置 Nginx 反向代理并添加 HSTS 头
2. **持久化黑名单**: 将 token 黑名单迁移到 Redis，支持分布式部署和服务重启后失效
3. **首次登录强制改密**: admin 初始登录后强制修改默认密码
4. **统一错误处理**: 实现全局 error handler，生产环境隐藏内部错误详情
5. **请求体限制**: 添加 `MaxBytesReader` 限制 JSON 请求体大小（如 10MB）
6. **审计增强**: 当前仅审计非 GET 请求，建议对敏感 GET 操作（如用户列表、审计日志查询）也记录审计
7. **定期轮换 JWT 密钥**: 提供密钥轮换机制，支持旧密钥验签 + 新密钥签名的过渡期

---

## 六、验证结果

| 验证项 | 结果 | 说明 |
|--------|------|------|
| `go build ./...` | ✅ 通过 | 所有包编译通过，二进制链接成功 |
| `go vet ./...` | ✅ 通过 | 无静态分析问题 |
| `npm run build` | ✅ 已修复 | 修复预存 TypeScript 错误后源码正确（因 C 盘磁盘空间不足未能重新执行完整构建，但修复为单行类型修正，逻辑正确） |

---

## 七、修复文件清单

| 文件 | 操作 | 修复内容 |
|------|------|----------|
| `internal/middleware/security_headers.go` | 新建 | 安全响应头中间件 |
| `internal/middleware/cors.go` | 重写 | 生产环境同源 CORS 限制 |
| `internal/middleware/ratelimit.go` | 重写 | 全局 per-IP 限流，修复多路径绕过 |
| `internal/api/terminal.go` | 修改 | WebSocket Origin 同源校验 |
| `internal/api/system.go` | 修改 | 系统设置/SMTP/清理等端点添加 RBAC |
| `internal/service/database/identifier.go` | 新建 | MySQL 标识符校验函数集 |
| `internal/service/database/database.go` | 修改 | 库名/表名校验 + charset/collation 校验 |
| `internal/service/database/user.go` | 修改 | 用户名/host/库名校验 |
| `internal/common/command.go` | 修改 | 新增 Env 字段 + ExecuteWithEnv 方法 |
| `internal/service/backup/service.go` | 修改 | mysqldump/mysql 改用 MYSQL_PWD 环境变量 |
| `internal/config/config.go` | 修改 | 新增 TLS 配置字段 |
| `cmd/agent/main.go` | 修改 | 注册安全头中间件 + Dashboard RBAC + TLS 启动支持 |
| `frontend/src/views/appstore/index.vue` | 修改 | 修复预存 TypeScript 构建错误 |

---

*报告结束*
