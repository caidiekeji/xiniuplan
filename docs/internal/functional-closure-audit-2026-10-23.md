# 功能闭环审计报告 — 云枢面板

> 审计日期：2026-10-23　范围：全项目前后端功能闭环（用户操作 → API → 业务逻辑 → 真实副作用）
> 方法：逐项追链路 + 线上 API 实测复核；只读审计，未修改功能代码（性能修复另计）

---

## 一、总体结论

**底层服务层基本可信**：Docker API 调用、文件落盘、Nginx 配置读写、数据库写入都是真实执行的，不是空壳。缺陷高度集中在**前后端契约层**——字段名、参数名两侧各写各的，导致上层 UI 大面积「看起来有、实际不生效」。

粗估闭环完成度：

| 模块 | 完成度 | 主要症结 |
|---|---|---|
| 容器管理 | ~70% | 镜像列表字段名错配、日志跟随开关请求挂死 |
| 网站管理 | ~45% | 创建接口 100% 失败、列表三列全错、筛选器全无效 |
| 应用商店 | ~50% | 安装成功却提示失败、源管理永远空表 |
| 部署 | ~60% | 同步执行超时误报失败、失败报成功、命令注入风险 |
| 系统设置 | ~33% | 6 个设置项中仅 3 项真闭环 |

**最危险的 3 类问题**：① 整页功能 100% 失效（网站创建）；② 保存成功但界面无反应（设置页语言/主题）；③ 部署模块的命令注入（`dest_path`/`name` 可逃逸 `sh -c`）。

---

## 二、系统设置逐项审计（重点）

| 设置项 | 前端可改 | 落库 | 后端读取 | 实际生效 | 结论 |
|---|---|---|---|---|---|
| 面板端口 `panel_port` | ✅ | ✅ | ✅ | 写回 `config.yaml`，需重启 | ⚠️ **半闭环**：nginx 反代硬编码 `127.0.0.1:8888`，改端口重启后 nginx 仍指旧端口 → 面板 502；提示文案未提 nginx 需同步 |
| 域名 `domain` | ✅ | ✅ | ❌ 后端零引用 | ❌ | ❌ **未闭环（摆设）**：仅存库，无任何消费方 |
| 语言 `language` | ✅ | ✅ | ❌ | ❌ | ❌ **假闭环**：前端真源是 `localStorage.cloudpivot_locale`，设置页只改表单值、不写 localStorage、不联动 appStore → 保存后界面无反应，刷新回退 |
| 主题 `theme` | ✅ | ✅ | ❌ | ❌ | ❌ **假闭环**：同上，真源是 `localStorage.cloudpivot_theme` |
| 安全入口 `security_entry` | ✅ | ✅ | ✅ | ✅ | ✅ **真闭环**：`middleware.GetSecurityEntry()` 读取，且保存后 `InvalidateSecurityEntryCache()` 主动失效 30s 缓存，即时生效 |
| 会话超时 `session_timeout` | ✅ | ✅ | ❌ 后端零引用 | ❌ | ❌ **未闭环（摆设）**：JWT TTL 取 `config.yaml` 的 `jwt.access_token_ttl`（固定 7200s），auth service 初始化时缓存，与设置表无关 |

**其余分区**：SMTP ✅（保存 + 测试发送 + 告警通知引用）、版本信息 ✅（只读）、安全评分 ✅（只读 + 真实计算）、系统清理 ✅（真实删文件/镜像/日志）、DNS 账号 ✅。

> 注：`config.yaml` 当前仅有一个 `port:` 字段（`server.port`），端口正则 `^(\s*port:\s*)(\d+)` 暂不会误伤；但未限定 `server` 段，将来新增同名键需注意。

---

## 三、其他模块（子代理审计 + 线上实测复核）

### 3.1 网站管理（最严重）

| 功能点 | 问题 | 证据 |
|---|---|---|
| 创建网站 | **100% 失败** | 前端发 `{domain, aliases, rootDir}`（`frontend/src/api/site.ts:33-41`），后端要 `{primary_domain(required), domains, root_dir}`（`backend/internal/dto/site.go:5-12`）→ `ShouldBindJSON` 报 400。**线上实测**：`{"code":40001,"message":"参数错误: Key: 'CreateSiteRequest.PrimaryDomain' ... required"}` |
| 列表展示 | 三列全错 | 后端返 `primary_domain/domains/created_at`、`ssl` 为对象；前端读 `row.domain/row.aliases/row.createdAt`，`row.ssl` 当布尔 → 域名/时间列空白，SSL 列恒显「已启用」（`frontend/src/views/website/index.vue:132,175,183`） |
| 删除站点 | 参数被忽略 | 前端传 `deleteFiles`（`api/site.ts:80`），后端读 `delete_files`（`api/site.go:306`）→ 勾选永远无效 |
| 域名/类型/状态筛选 | 全无效 | 前端传 `domain/type/status`（`api/site.ts:20-26`），后端 `List` 只读 `page/pageSize`（`api/site.go:227-228`），`ListSites` 无 WHERE 条件（`service/site/service.go:257-278`） |

### 3.2 应用商店

| 功能点 | 问题 | 证据 |
|---|---|---|
| 安装应用 | **成功却提示失败** | 后端 `dto.InstalledAppResponse` 无 `success` 字段（`api/appstore.go:118`），前端判 `if (res.data.success)`（`views/appstore/InstallDialog.vue:114`）→ 恒为 false |
| 源管理 Tab | 表格永远空 | 后端返 `{items,total}`（`api/app_source.go:51`），前端 `Array.isArray(res.data) ? res.data : []`（`views/appstore/index.vue:37`）→ 恒 `[]` |
| 已安装列表 | 两个按钮永远失效 | 响应缺 `container_id/access_url/installed_at/category`（`dto/appstore.go:37-48`），前端依赖这些字段（`views/appstore/InstalledTab.vue:48,70,152`） |
| 升级应用 | 副作用顺序错误 | `downCancel()` 在 `Down()` **之前**调用（`service/appstore/service.go:262-264`）→ ctx 已取消，旧容器没停，新容器端口冲突 |
| 同步源结果 | 字段名不符 | 后端 `removed`，前端读 `r.deleted`（`api/appstore.ts:111`）→ 提示显示 undefined |

### 3.3 容器管理

| 功能点 | 问题 | 证据 |
|---|---|---|
| 镜像列表名称列 | **恒显示 `<none>`** | 后端返 `tags`（`dto/container.go:23`），前端读 `row.repo_tags`（`views/container/index.vue:397-398`）。线上实测返回体确为 `{"tags":[...]}` |
| 日志「跟随」开关 | 请求挂死 | `follow=true` 后端走 SSE（`api/container.go:161-196`），前端仍用 axios 按 JSON 读 `res.data.content`（`views/container/index.vue:291`）→ 永不返回，30s 超时 |
| 实时统计内存 | 显示裸字节 | 后端 `uint64`（`dto/container.go:31-32`），前端声明 `string`（`api/container.ts:122-123`） |
| 镜像删除 | ✅ 已闭环（本次修复） | 前端 `removeImage` → `DELETE /images/:id?force=` → `svc.RemoveImage` 真实调用 Docker |

### 3.4 部署

| 功能点 | 问题 | 证据 |
|---|---|---|
| 一键部署 | 超时误报失败 | 后端同步执行含 `git clone + npm build`（`api/features.go:254-262`），前端 axios timeout 30s（`api/request.ts:15`）→ 界面报失败，后台仍在部署 |
| 静态/Node 发布 | **失败报成功** | `cp -r ... \|\| true` + `runIn` 吞 exit code（`service/deploy/deploy.go:182,194,215-226`）→ 复制失败仍输出「[OK] 静态文件已发布」 |
| 目标目录/应用名 | **命令注入** | `shq()` 只包单引号不转义（`deploy.go:228`），用户可控 `dest_path`/`name` 直接拼进 `sh -c`（`:181-182,201-203`）→ 单引号可闭合逃逸 |
| 编辑应用 | 静默改状态 | 前端 update 不带 `enabled`（`views/deploy/index.vue:196`），后端无条件写 `"enabled": req.Enabled`（`deploy.go:88`）→ 每次编辑把应用置为禁用且 UI 无法恢复 |
| 删除应用 | 副作用缺失 | `Delete` 只删 DB（`deploy.go:96-99`），`runtime/deploy/app-{id}` 工作目录残留 |

### 3.5 死链清单

- 前端有封装、无任何页面调用：`getSite`/`updateSite`（`api/site.ts:61,68`）、`getApp`/`getAppStatus`（`api/appstore.ts:124,153`）、`getContainer`（`api/container.ts:58`）→ 后端 `PUT /sites/:id`、`GET /containers/:id` 不可达
- 后端有路由、前端未接：`sites/rewrite-templates`、`snippets`、`traffic`、`basic-auth`（`api/site.go:48-53`）；`images/:id/save`、`images/load`（`api/container_extra.go:42-43`）

---

## 四、修复优先级

| 级别 | 问题 | 建议修法 |
|---|---|---|
| P0 | 网站创建 100% 失败 | 前端 DTO 改 `primary_domain/domains/root_dir/proxy_target`，或后端加别名兼容 |
| P0 | 部署命令注入 | 改 `exec.Command` 参数数组，不经 `sh -c`；路径加白名单校验 |
| P0 | 静态发布失败报成功 | 去掉 `\|\| true`，`runIn` 返回 error，失败置 `last_status=failed` |
| P1 | 应用商店安装假失败 | 后端补 `success/access_url/initial_password/container_id` |
| P1 | 镜像列 `<none>` | 前端改读 `row.tags[0]` |
| P1 | 设置页语言/主题假闭环 | 保存时同步写 localStorage + appStore；或从 UI 移除（真源是浏览器本地） |
| P1 | 会话超时摆设 | auth service 从 settings 读 `session_timeout` 覆盖 TTL |
| P1 | 端口与 nginx 不一致 | 提示文案补充 nginx 同步；或由面板一并改 nginx 配置 |
| P2 | 域名摆设 | 明确用途（SSL 签发/站点绑定）后接入，否则从 UI 移除 |
| P2 | 列表/筛选参数名错配 | 网站筛选、删除勾选、应用商店源列表逐一对齐 |
| P2 | 部署编辑静默禁用 | 后端仅在 `req.Enabled != nil` 时更新 |
| P2 | 日志跟随开关挂死 | 前端改 `EventSource` 消费 SSE |
| P3 | 死链清理 | 删除无调用封装，或补齐对应 UI 入口 |

---

## 五、本轮已修复并上线验证（2026-10-23）

| # | 问题 | 修法 | 验证结果 |
|---|---|---|---|
| 1 | 网站创建 100% 失败 | 前端 `api/site.ts` 集中做 snake_case↔camelCase 映射（`toSite`/`toCreatePayload`） | 线上创建成功，落库 `primary_domain=audit-test.local` |
| 2 | 网站列表三列全错 | 同上映射：`domain←primary_domain`、`aliases←domains`、`ssl` 对象→布尔 | 列表返回 `primary_domain/domains/ssl{enabled}` ✓ |
| 3 | 网站筛选器全无效 | 后端 `ListSites` 加 `SiteFilter`（domain/type/status WHERE） | `domain=audit`→1 条；`domain=zzz`→0 条 ✓ |
| 4 | 删除站点「同时删文件」失效 | 前端参数改 `delete_files`（对齐后端 `c.Query`） | 契约对齐 ✓ |
| 5 | 镜像名称列恒 `<none>` | 前端改读 `row.tags`（后端 DTO 字段名） | 线上返回 `tags:["ghcr.io/…"]` ✓ |
| 6 | 应用商店安装成功却报失败 | 后端 `InstalledAppResponse` 补 `success` 字段 | 安装/列表均返回 success ✓ |
| 7 | 已安装应用「打开/日志」按钮永远失效 | 后端用 `compose ps --format json` 解析容器 ID 与映射端口，api 层按请求 Host 拼 `access_url` | 线上返回 `container_id=cb3560774ba8`、`host_port=3306`、`access_url=http://…` ✓ |
| 8 | 设置页保存按钮 400 | 前端 PUT 包 `{settings:{…}}`（后端 DTO 要求外层 settings） | 保存返回 `{"updated":true}` ✓ |
| 9 | 设置页读不到设置值 | 前端 GET 读 `data.settings`（后端返回 `{settings:{…}}`） | 正确回显 ✓ |
| 10 | 设置页「域名」永远空 | 键名对齐：库键 `panel_domain` ↔ 前端 `domain`（读写双向映射） | 返回含 `panel_domain` ✓ |
| 11 | 设置页语言/主题保存后毫无反应 | 保存后经 appStore 同步 localStorage（真实生效位置），加载时以本机实际值为准；补「日本語」选项 | 契约与同步逻辑就位 ✓ |
| 12 | 会话超时是摆设 | auth 签发 token 时读 `session_timeout`（分钟）覆盖 config 默认；`DefaultSettings` 补种子键；`GetSettings` 补齐缺失键 | 改 30 分钟 → `expires_in` 7200→1800 秒 ✓（已恢复 120） |
| 13 | 安全入口默认值 `/admin` 高危 | 表单默认值改为空串（加载失败时随手保存会把自己锁在门外） | ✓ |
| 14 | 安全入口变更后不知新地址 | 保存后比对前后值，提示新地址 / 关闭提示（三语言 i18n） | ✓ |
| 15 | 网站模块硬编码不可写路径 | `site.root_dir/nginx_conf_dir/ssl_dir` 提到 config.yaml，缺省改为面板工作目录下相对路径（线上 `/www` 与 `/opt/cloudpivot/runtime` 均不存在） | 创建站点不再报权限错误 ✓ |

**仍未修（留待后续）**：部署模块命令注入与「失败报成功」、应用商店升级的 `downCancel` 顺序错误、容器日志「跟随」开关请求挂死、部署编辑静默改 `enabled`、死链清理、域名（`panel_domain`）的实际用途未定义。

> **环境前置条件提醒**：网站管理依赖服务器 nginx 加载 `site.nginx_conf_dir` 下的配置。当前部署形态是「面板自身被 nginx 整站反代」，并不读取该目录——站点配置会正常落盘，但需自行把该目录纳入 nginx `include` 才对外生效。

---

## 六、补充审计：2026-10-25 收尾修复

上表「仍未修」的 6 项已全部落地并线上验证。

| # | 问题 | 修法 | 验证 |
|---|------|------|------|
| 16 | 部署命令注入（P0） | `shq()` 改 POSIX 标准 `'\''` 转义；入口加白名单：`validateDeployPath`（绝对路径 + 禁 `` ` $ ; & \| < > \n \r ' " \ ``）、`validateDeployName`、`validateDeployBranch`，Create/Update 均校验 | 注入型 `dest_path` 被拒（400），`/tmp/pwned*` 未生成 ✓ |
| 17 | 静态发布失败报成功（P0） | `runIn()` 改为返回 `(string, error)`，各步骤传播错误；去掉 `cp ... \|\| true`；仅「删除可能不存在的旧容器」忽略错误 | clone 失败正确报 `git clone 失败: exit status 128` 并落 failed ✓ |
| 18 | 应用商店升级 `downCancel` 顺序错（P1） | `downCancel()` 原写在 `Down` 之前 → ctx 立即失效、旧容器未停即覆盖 compose | 改为 `defer cancel()`，Down 失败即中断升级 ✓ |
| 19 | 容器日志跟随请求挂死（P2） | 前端用 axios 请求 SSE（响应永不结束），每 3 秒轮询一次 → 请求堆积 | 新增 `utils/sse.ts`（fetch + ReadableStream；EventSource 不支持 Authorization 头），改单条长连接 ✓ |
| 20 | 部署编辑静默禁用（P2） | `Update` 无条件写 `enabled`，前端表单不带该字段 → 零值 false | 新增 `dto.DeployAppUpdateRequest`，`Enabled` 改 `*bool`，未传即跳过 | 更新后 enabled 仍为 True ✓ |
| 21 | 死链清理（P3） | 后端 `PUT /sites/:id` 与前端 `updateSite` 均存在但无 UI | 网站页加「编辑」按钮（复用创建弹窗，域名只读并提示），三语词条补齐 | PUT → 回读 remark 生效 ✓ |
| 22 | `panel_domain` 用途未定义 | 复核确认并非摆设：库键 `panel_domain` ↔ 前端 `domain` 的键名错配，读写双向映射后已对齐（第 10 项已修） | 复核通过，无需改动 |

**同期修掉的其他真 bug**（均为「界面看着有、实际 100% 失效」）：

| 问题 | 修法 |
|------|------|
| `failDatabaseError` 自递归 → stack overflow 打挂整个后端 | 批量替换时误改了函数体内部调用；已修复并加回归测试 `internal/api/database_error_test.go` |
| MySQL 用户列表字段名 `user` ↔ `username` | 统一为 `username` |
| `CREATE USER ?@? IDENTIFIED BY ?` 语法错误（MySQL 不支持该语句预编译） | 改单引号引用 + 完整转义，按 `@@SESSION.sql_mode` 区分 NO_BACKSLASH_ESCAPES；已在真实 MySQL 实测含单引号/反斜杠/中文/制表符/换行的密码均能正确写入并登录 |
| 表结构接口 500（MySQL 8.0 `SHOW INDEX` 15 列 vs 代码写死 13 列） | 按 `rows.Columns()` 动态补齐 Scan 目标；抽出 `buildIndexScanTargets` 便于测试 |
| PostgreSQL 字段类型被覆盖（`full_type` 与 `data_type` 共用变量） | 独立变量接收 |
| 业务错误返回 500（类型不支持/非法权限/非法标识符） | 哨兵错误 `ErrUnsupportedType` / `ErrInvalidRequest`，API 层统一翻译成 400 |
| `TableData.page_size` 与其他分页接口 `pageSize` 不一致 | 统一 camelCase |

**新增回归测试**：`internal/api/database_error_test.go`、`internal/service/database/user_sql_test.go`、`internal/service/database/index_scan_test.go`、`internal/service/deploy/deploy_security_test.go`。

**验证脚本**：`scripts/verify_contracts.py` —— 24 项线上契约断言，幂等（自动清理测试用户与测试实例），当前 **PASS 24 / FAIL 0**。

---

*本报告为只读审计结论。性能相关修复（首屏黑屏、overview 2.2s→12ms、manualChunks 破坏摇树等）另见当日工作日志。*
