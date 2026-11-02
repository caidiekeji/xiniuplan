# 设计闭环审查报告 — 云枢面板（镜像拉取新链路专项 + 全局复查）

> 审查日期：2026-10-26　审查人：Ray（只读审查，未修改任何代码）
> 范围：全仓前后端闭环复查；重点细审最近改动的镜像拉取链路（未提交变更：`pkg/docker/registry.go / pull.go`、`api/image_pull.go`、`container.ts pullImage`、`container/index.vue pullModal`）
> 方法：逐链路静态追读（前端入口 → API → service → Docker/DB → 回显）+ 历史审计修复项抽查。
> **实测说明**：线上面板 `http://192.168.31.154:8888` 从本机不可达（TCP connect 超时，疑网络隔离/防火墙），本轮所有结论基于静态代码证据，未做线上接口实证（详见第五节）。

---

## 一、总评

**底层实现质量明显优于上一轮审计时的状态，历史修复项抽查全部成立；但镜像拉取这条新链路存在两个"必死"级缺陷，核心功能实际不可用。** `service/container/service.go:300` 的注释明确写着「不设整体超时：大镜像拉取可能超过 30 分钟」，然而 `pkg/docker/pull.go:46` 的 `withTimeout(ctx)` 会对无 deadline 的 `context.Background()` 再加一层 **30 秒** 默认超时——两层代码互相矛盾，以 30 秒为准，意味着任何下载超过 30 秒的镜像拉取都会以 `context deadline exceeded` 失败。即便修掉这一处，`http.Server.WriteTimeout=60s`（`config.yaml` + `main.go:296`，全程无 `SetWriteDeadline` 覆盖）也会把超过 60 秒的 SSE 拉取流、以及**日志跟随长连接**（上一轮审计修复项 #19）在 60 秒时掐断——后者前端对流结束（`done`）不报错、不重连，UI 会永远停留在「跟随中」的假状态。此外 `GET /system/settings` 明文回传 SMTP/私有仓库密码，属于新的敏感信息泄漏面。其余模块（容器启停/删除/日志非跟随/镜像列表删除/设置读写回显）契约两侧对齐，闭环健康。

闭环健康度估算：容器管理 ~85%（拉取链路 0 分）、系统设置 ~80%（密码泄漏扣分）、其余模块与上轮审计结论一致（已修复项未见回归）。

---

## 二、发现清单

### P0（功能性坏死 / 安全）

**P0-1　镜像拉取 30 秒必死：service 层"不设超时"的承诺被 pull.go 的 withTimeout 覆盖**
- 位置：`backend/pkg/docker/pull.go:46`（`ctx, cancel := withTimeout(ctx)`）＋ `backend/pkg/docker/client.go:80-88`（`withTimeout` 对无 deadline 的 ctx 加 `DefaultTimeout = 30s`）＋ `backend/internal/service/container/service.go:302-310`（用 `context.Background()` 调用，并注释「不设整体超时」）
- 问题：`PullImageWithProgress` 收到的是 `context.Background()`（无 deadline），进入 `pull.go` 后 `withTimeout` 立即挂上 30 秒 deadline。Docker `ImagePull` 的 ctx 到期即中断整个拉取。**任何下载耗时 >30s 的镜像 100% 拉取失败**，前端收到 `{"error":"读取拉取输出失败: context deadline exceeded"}`。
- 证据链：`api/image_pull.go:168` → `svc.PullImageWithProgress`（`service.go:309` 传 `context.Background()`）→ `pull.go:46` `withTimeout(ctx)` → `client.go:85` `context.WithTimeout(ctx, 30*time.Second)` → `pull.go:49` `d.cli.ImagePull(ctx, ...)`。service.go:300 的注释与实际行为直接矛盾，属典型"注释说了没做"。
- 建议修法：`pull.go` 内改为 `if _, ok := ctx.Deadline(); !ok { ctx = context.Background() }`（拉取语义上不该有默认超时，由调用方/HTTP 生命周期控制），或给 `DockerClient` 增加 `PullTimeout` 可配置项（建议 ≥30 分钟）；同时修正 service.go 注释或让注释与实现一致。

**P0-2　`http.Server.WriteTimeout=60s` 掐断一切 SSE 长连接：镜像拉取 >60s 断流、日志跟随 60s 静默假死**
- 位置：`backend/cmd/agent/main.go:292-297`（`WriteTimeout: 60s`）＋ `backend/config.yaml`（`write_timeout: 60`）；全仓无 `SetWriteDeadline` / `http.NewResponseController` 调用（已 grep 确认）
- 问题 1（拉取）：Go 的 WriteTimeout 从请求头读完起算，覆盖整个响应生命周期。即使 P0-1 修复，拉取流也会在 60 秒被服务端强行关闭。`image_pull.go:151-166` 的 10 秒心跳只能防中间代理空闲超时，防不了自己 `http.Server` 的硬 deadline。
- 问题 2（日志跟随，回归性质）：上一轮审计修复项 #19（日志跟随改 SSE 长连接）在 60 秒后被服务端关闭。而前端 `frontend/src/utils/sse.ts:56-70` 读到 `done` 时只 `break`，**不调用 onError**；`container/index.vue:436-440` 只有 onError 才会复位 `logFollow` —— 结果是 60 秒后连接已死，开关仍显示「跟随中」，再无新日志进来，用户无任何感知。这与 #19 要修的「界面显示跟随中但实际没在跟」是同一症状，只是触发原因从请求堆积变成了服务端断流。
- 建议修法：① 服务端对 SSE 路由（`/images/pull`、`/containers/:id/logs?follow=true`）在写响应头前用 `http.NewResponseController(c.Writer).SetWriteDeadline(time.Time{})` 解除写超时（或 WriteTimeout 设 0、靠 ctx/心跳管理）；② 前端 `subscribeSSE` 在 `done` 分支也回调 onError（或加 onEnd），并考虑自动重连。

### P1（体验断裂 / 中等安全）

**P1-1　`GET /system/settings` 明文回传 SMTP 密码与私有仓库密码**
- 位置：`backend/internal/api/system.go:63-72`（GetSettings 原样返回）＋ `backend/internal/service/system/service.go:61-78`（全 map 无脱敏）＋ `frontend/src/api/system.ts`（新增 `docker_registry_pass` 直接进出）
- 问题：任何持有 `PermSettingView` 的用户（非仅 admin 角色，RBAC 下可授）都能读到 `smtp_pass`、`docker_registry_pass` 明文。新增的 registry 表单虽有 `type="password"`，但只是输入态遮罩，回显时真值已在响应体里。新功能把泄漏面从「管理员自己看自己」扩大到「所有可看设置的账号」。
- 建议修法：GET 时对 `smtp_pass` / `docker_registry_pass` 返回空串或 `"******"` 哨兵；PUT 时收到哨兵值则不更新该键（保留原密码）。`GET /system/smtp` 若有同样问题一并处理。

**P1-2　取消拉取只断前端，后端 Docker 拉取继续狂奔**
- 位置：`backend/internal/api/image_pull.go:168`（`svc.PullImageWithProgress` 不传请求 ctx）＋ `container/service.go:309`（`context.Background()`）＋ `pull.go` 的 `writeEvent`/回调忽略写错误
- 问题：用户关闭弹窗（`index.vue:368-371 closePull → abort()`）或网络断开后，服务端 write 失败被静默吞掉，`ImagePull` 在 `context.Background()` 下继续拉完整个镜像——大镜像会持续占用带宽/磁盘几十分钟，且无任何接口能停掉它。
- 建议修法：把 `c.Request.Context()` 一路传入 `PullImageWithProgress`；客户端断开 → ctx 取消 → Docker 拉取终止。注意与 P0-1 的修法协同（用请求 ctx 后 `withTimeout` 的 30s 逻辑必须先去掉，否则 60s WriteTimeout 会反过来把拉取限制在 60s——正好两处一起修：请求 ctx + 解除该连接写超时）。

**P1-3　拉取弹窗默认选中加速源，拉私有仓库镜像必然报错**
- 位置：`frontend/src/views/container/index.vue:290-292`（`pullForm.mirror = firstMirror ? firstMirror.id : ''`）＋ `backend/pkg/docker/registry.go:142-144`（`ApplyMirror` 拒绝已含 registry 的镜像）
- 问题：只要设置里配了加速源，打开弹窗就默认选中它；用户此时输入 `harbor.example.com/a/b:1` 这类私有仓库镜像点拉取，后端直接报「镜像 … 已指定仓库 …，不使用加速源」，需要用户自己理解并手动切回「直连 Docker Hub」。加速源越常用，踩中概率越高。
- 建议修法：前端在提交前按后端同款规则（第一段含 `.`/`:` 或为 localhost）判断镜像是否自带仓库，是则自动忽略 mirror（或置灰下拉框并提示「已识别为私有仓库镜像，不走加速」）。

### P2（不一致 / 打磨项）

**P2-1　`PullProgressWriter` 是死代码**
- 位置：`backend/pkg/docker/pull.go:116-166`。全仓无引用（已 grep 确认），且它实现的 200ms 节流恰恰是 `image_pull.go` 缺的——见 P2-2。要么接入 API 层，要么删除。

**P2-2　SSE 进度帧未节流，与代码注释相悖**
- 位置：`backend/internal/api/image_pull.go:168-170`（每条 Docker 进度消息直接 `writeEvent`）；对照 `pull.go:118-119` 注释「按 200ms 节流…只推送百分比变化 >= 1 的消息」——实际 API 层没有做任何节流。大镜像每层每秒多条消息，帧量可观。建议复用/搬移 `PullProgressWriter` 的节流逻辑。

**P2-3　`openPull` 预填私有仓库用户名的条件与作用域混乱**
- 位置：`frontend/src/views/container/index.vue:293-295`：仅当「同时存在加速源和私有仓库」才预填 `username`；且该 username 会被无条件随请求发送（`handlePull` → `username: pullForm.username || undefined`），即使选的是普通加速源，后端也会把它编进 RegistryAuth 发给加速源（`image_pull.go:114-120`）。当前因 password 为空、`RegistryAuth` 判空返回 ""（registry.go:183-185）而无实害，但逻辑纯属巧合性安全。建议仅当 `mirror === '@private'` 时才发送凭据。

**P2-4　`ParseImageRef` 对 `@sha256` 摘要镜像 + 加速源的组合会产出损坏引用**
- 位置：`backend/pkg/docker/registry.go:74-78`（摘要形式整体塞进 `Repository`、`Registry` 留空）＋ `ApplyMirror:142-144` 只拦截 `Registry != ""`。若用户输入 `nginx@sha256:...` 并选加速源，会改写成 `mirror/library/nginx@sha256:...`——多数加速源不支持摘要拉取，报错信息还会指向摘要而非改写行为。`ApplyMirror` 应同样拒绝含 `@` 的引用。

**P2-5　`ListRegistries` 每请求 4 次全量读设置表**
- 位置：`backend/internal/api/image_pull.go:43-44,61`（4 次 `settingValue`，每次 `GetSettings()` 全表 `Find`）。拉取弹窗每次打开都触发。建议一次读取复用。同类：`Setting.Value` 定义 `size:1024`（model/setting.go:8），mirrors 为 textarea 多行输入，超长有截断风险（取决于 DB 迁移产物），建议放宽至 2048 或文档约束。

**P2-6　同一镜像可并发重复拉取**
- `image_pull.go` 无按镜像名的互斥/去重，两个用户（或双击）可对同一镜像同时发起 Docker pull。Docker daemon 层会去重层下载，实害有限，记为低优先级。

---

## 三、镜像拉取新链路专项审查结论

链路图：`container/index.vue openPull/handlePull` → `api/container.ts pullImage`（fetch+ReadableStream 解析 SSE，POST 带 Bearer，实现正确）→ `api/image_pull.go PullImage`（SSE、心跳、mutex 写帧、权限点 `PermContainerView/Update` 对齐）→ `container/service.go` → `pkg/docker/pull.go`。

**契约两侧对齐良好**：`/images/registries` 返回的 `registries/has_private/private_user` 与前端 `RegistryListData` 一致；SSE 帧格式（进度帧 `layer_id/status/current/total/percent`、心跳 `{"status":"pulling"}`、错误帧 `{"error":...}`、完成帧）前端 `pullImage` 均正确解析并分发；错误/完成/心跳三态在 `index.vue:322-348` 的处理分支完整，完成后会 `loadImages()` 刷新镜像列表，失败会 `pullLoading=false` 并弹错。镜像名解析（`ParseImageRef`）有 29 个用例的单测覆盖（`registry_test.go`），边界处理（tag/端口区分、摘要不拆分、localhost 判定）正确。

**但链路终点（真正把镜像拉下来）被 P0-1 的 30 秒超时判了死刑**——这是本次审查最重要的结论：该功能在「名字解析、UI、流式协议、凭据」上都做对了，唯独在最核心的执行层因 `withTimeout` 的复用而无法工作。修复 P0-1 + P0-2（两者必须一起修，见 P1-2 修法说明）+ 顺手做 P1-2（取消传播）后，此链路才能达到「完整闭环」。

其他专项结论：`base64Encode` 手写实现与标准库一致（有测试兜底）；`RegistryAuth` 结构体编码顺序稳定；`sanitizeFilePart` 处理了文件名注入；exec WebSocket 的 `CheckOrigin` 放行所有 Origin（`container_extra.go:20-22`），但因 JWT 走 query token 且面板同源部署，风险可控，建议后续收紧。

---

## 四、历史修复项抽查（audit-2026-10-23 编号 1~22）

抽查 6 项，静态复核结果：

| 编号 | 抽查点 | 结论 |
|---|---|---|
| #1/#4 | `frontend/src/api/site.ts` `toSite`/`toCreatePayload` 映射、删除传 `delete_files` | ✅ 成立（site.ts:79-150） |
| #12 | `session_timeout` 被 auth 侧真实读取 | ✅ 成立（`service/auth/totp.go:171` 直读 settings 表 `key=session_timeout`） |
| #16 | 部署命令注入修复 | ✅ 成立：`shq()` 为标准 `'\''` 转义（deploy.go:385-389）；`validateDeployPath` 拒绝元字符与相对路径、Create/Update 双入口校验（deploy.go:107-153） |
| #19 | 日志跟随改 SSE | ⚠️ 前端实现成立，但被 **P0-2（WriteTimeout 60s + done 不回调）** 削弱为 60 秒假死——属于修复不彻底而非回归，需按 P0-2 补齐 |
| #13 | 安全入口默认空串 | ✅ 成立（settings/index.vue 默认 `security_entry: ''` 及注释） |
| #22 | `panel_domain` 双向映射 | ✅ 成立（system.ts toSettings/toSettingsPayload） |

未发现上轮已修项出现回归；上轮「仍未修」清单中的部署模块两项（#16/#17）已确认落地。

---

## 五、实测记录

| 动作 | 结果 |
|---|---|
| `GET http://192.168.31.154:8888/health` | ❌ TCP connect timeout（10s/60s 两档均超时，非 4xx/5xx） |
| `POST /api/v1/auth/login` | ❌ 未达（依赖上一步连通性） |

结论：本机到 192.168.31.154:8888 的网络路径不通（连接超时而非拒绝，疑为跨网段隔离/防火墙）。本轮 P0/P1 结论均有完整静态证据链（文件:行号），置信度高；若队长需要线上实证，建议在可达该网段的机器上执行：① `curl -N -m 120 -X POST /api/v1/images/pull -d '{"image":"alpine"}'` 观察约 30 秒是否出现 `context deadline exceeded`（验证 P0-1）；② `curl -N -m 90 "/api/v1/containers/<id>/logs?follow=true&tail=1"` 观察 60 秒时连接是否被服务端关闭（验证 P0-2）；③ `GET /api/v1/system/settings` 检查响应中 `smtp_pass`/`docker_registry_pass` 是否明文（验证 P1-1）。

---

## 六、亮点（保持）

- `registry_test.go` 覆盖解析/改写/还原/编码 29 个用例，含 round-trip，是全仓测试质量标杆。
- SSE 帧协议设计清晰（心跳/错误/完成三态），前端 fetch 流式解析与 `AbortController` 取消实现规范。
- 容器 service 层 `errDockerUnavailable` 统一降级、日志 follow 的 ctx 生命周期注释带线上复盘日期，可维护性好。
- docker-compose 强制 `ADMIN_PASSWORD` 非空（`:?` 语法），消除了默认弱密码路径。

---

*本报告为只读审查结论，未修改任何代码。问题修法仅为建议，采纳与否由队长决定。*
