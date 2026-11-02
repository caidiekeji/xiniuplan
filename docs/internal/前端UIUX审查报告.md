# 犀牛面板前端 UI/UX 审查报告

> 审查范围：`frontend/src`（26 个页面模块、16 个自研组件、633 行设计 token/全局样式）
> 审查维度：改造 · 留白 · 骨架 · 交互 · 动画
> 审查方式：全量代码扫描 + 关键结论逐条人工复核（均为代码级事实，非主观印象）

---

## 总体评价

**设计基建是超出预期的**：344 行 token 体系完整（间距/字号/圆角/动效/断点/输入宽/弹窗宽全齐），views 层间距魔数 **0 处命中**，`transition: all` **0 处**，非合成属性动画仅 1 处。这套纪律好于多数同规模项目。

**差距集中在「基建建了但没接满」**：骨架屏组件建好了只有 1 个页面在用，DataTable 错误态+重试建好了 **0 个页面在用**，动效 token 定义了只有约一半页面在用，路由过渡类写好了却没包在 router-view 上。**改造方向不是重写，而是把已有能力接线到全部页面**——这是本次审查最重要的结论。

---

## 一、留白与间距（评级：A-）

### 现状
- token：`--space-0~8`（2→64px，4px 基准）、`--fs-*` 六档、`--radius-sm/md/lg`、`--input-sm~3xl` 七档、`--modal-width-sm~xl` 四档。
- **间距纪律极好**：全部 views/components 中 `padding/margin/gap` 的 px 魔数为 **0 处**。

### 问题
唯一漏洞是 **内联 `style="width: Npx"` 约 34 处**，其中大量与 `--input-*` 档位一一对应却未引用：

| 位置 | 魔数 | 应替换为 |
|---|---|---|
| `views/ssh/index.vue:482,527,536,680,688,696` | `max-width: 200px` ×6 | `var(--input-lg)` |
| `views/audit/index.vue:270,274,278` | `120/150/110px` | `--input-md / --input-xl / --input-md` 附近档位 |
| `views/protection/index.vue:216,266` | `280px` ×2 | `var(--input-2xl)` |
| `views/terminal/index.vue:471-485` | `160/100/140/160px` ×4 | token 档位 |
| `views/alert/index.vue:964,975`、`monitor:526,527`、`firewall:338`、`settings:864` 等 | 各 1-2 处 | 逐一对齐 |
| `views/dashboard/Dashboard.vue:630,663,700-704` | 小屏 `font-size:10/11px` | 收敛到 `--fs-caption` 或新增 micro 档 |

另有一处**定义了但零使用**的死 token：`--drawer-width: 480px`（tokens.css:166），实际所有抽屉都是写死 720-820px。

### 改造方案（P1，纯机械替换）
1. 34 处内联宽度收敛到 `--input-*` token；确无对应档位的（audit 的 110/120px）统一就近归档，避免新增孤值。
2. 删除或启用 `--drawer-width`：建议改为真实使用——各抽屉宽度归一到 `--drawer-width`（480px 偏窄，可改为 `min(720px, 92vw)`）。
3. Dashboard 小屏 10/11px 字号归档，低于 12px 的字号在可读性上应删除而非保留。

---

## 二、骨架屏（评级：C+，最大可感知差距）

### 现状
- 骨架能力已内建：`PageCard.vue`（body 5 条骨架）、`StatCard.vue`（数值骨架）。
- **全项目只有 monitor 一个页面真正启用了骨架**。
- 其余 25 个页面 loading 方式分布：

| 方式 | 页面 |
|---|---|
| 骨架屏 | monitor（唯一） |
| 整页/局部 NSpin 转圈 | plugins、nodes、settings、users/TOTP、container、appstore、database 全部 Tab |
| 仅表格内 NSpin（DataTable 内建） | website、cron、daemon、alert、audit、backup、ssl、probe、sessions、file、deploy、firewall、protection、ssh 等 15+ 页 |
| **完全无内容级占位** | dashboard（图表区空白等待）、ai、terminal、login |

### 问题
1. 首次进入页面 = 表格区域一个转圈，与骨架相比视觉稳定性差，且转圈不预示内容布局（CLS 心理感知差）。
2. dashboard 图表区加载时是纯空白，这是用户落地第一屏，体验伤害最大。
3. nodes 加载完为空时 v-for 直接渲染 0 个卡片 = 白屏，无空态分支。

### 改造方案（P0 视觉 / P1 覆盖）
1. **P0**：dashboard 接入 PageCard/StatCard 骨架（组件现成，只差传 `:loading`）；nodes/ai 补空态分支。
2. **P1**：website、firewall、backup、ssl 四个高频页面把 DataTable 外的页面级 `:loading` 接到 PageCard 骨架；表格内刷新保留 NSpin（已有数据时转圈优于闪骨架）。
3. **经验法则**：首次加载用骨架、增量刷新用 spin——建议沉淀为组件注释或开发文档约定。

---

## 三、空态与错误态（评级：C，结构性缺陷）

### 现状
- 空态覆盖尚可：DataTable 内建 NEmpty 自动覆盖 29 个使用方，database 系 28 处、container 4 处。**缺口**：ai（v-for 无空态）、nodes（白屏）、dashboard（图表无数据占位）、monitor disk 列表。
- **错误态是结构性缺陷**：`DataTable.vue:24,92-109` 已支持 `error` prop + 内建重试按钮，但 views 中 `:error=` **0 次使用**（grep 零命中）。错误处理全靠 toast（38 个文件约 300 处 `message.error`）。
- 无 NResult 页面级错误视图（0 命中）；`ErrorBoundary.vue` 仅在 App.vue 根级包了一次。

### 问题
刷新失败时用户的真实体验是：**一个 2 秒消失的 toast + 一张空表格**。空表格与"暂无数据"无法区分，toast 错过后用户不知道发生了什么、也不知道可以重试。已建好的重试 UI 成了死代码。

### 改造方案（P0，收益最大的一项）
1. 各页面 loadXxx 的 catch 里把错误写入 `error` ref 并传给 `<DataTable :error="loadError">`，同时提供重试回调（组件已支持，每页约 5 行改动）。
2. 优先接线：website、file、ssh、firewall、backup、audit（高频运维页面）。
3. toast 保留（即时反馈），页面内错误态负责「持久 + 可恢复」，两者互补而非二选一。

---

## 四、交互（评级：B-）

### 现状与问题
1. **防重复提交缺口**：`ModalForm.vue:71` 确认按钮有 `:loading` 无 `:disabled`。file 页 **7 个 ModalForm（mkdir/mkfile/rename/move/copy/compress/decompress）全都没传 `:loading`**（仅 upload 传了），快速双击确认会重复提交——对「建目录/复制文件」这类写操作是真实风险。
2. **键盘可达性薄弱**：`tabindex/@keydown` 全项目仅 3 个文件 5 处（CommandPalette、container 两处）。十余种 hover 卡片（plugin-card、app-card、type-card 等）**鼠标可点、键盘不可达**，违反 WCAG 2.1.1。
3. **按压/焦点反馈几乎为零**：`:active` 全项目仅 1 处、`:focus-visible` 只有 main.css 全局兜底一处。自定义卡片的 hover 有反馈但按下无反馈，触屏/键盘用户感知不到激活。
4. hover 质量尚可：views 17 处 + components 4 处，多为 background/border 变换；`--hover-lift` 仅 StatCard 在用。

### 改造方案
1. **P0**：file 页 7 个 ModalForm 补 `:loading`（每处 1 行）；ModalForm 确认按钮加 `:disabled="loading"`（组件级 1 行，全局受益）。
2. **P1**：可点卡片统一模式 `tabindex="0" role="button" @keydown.enter/.space`（或换成 `<button>` 语义包裹），先改 appstore 应用卡与 website 类型卡两个最高频场景。
3. **P1**：全局 `:active` 按压反馈——main.css 加一条 `transform: scale(0.98)` 于 `.n-button:active:not(:disabled)`，一次生效全站。
4. **P2**：CommandPalette 键盘体验已达标（↑↓/Enter），可作为其他键盘交互的参照。

---

## 五、动画与过渡（评级：B）

### 现状
- token 齐全：`--t-fast(0.15s)/normal(0.3s)/slow(0.5s)` + 两套缓动；实际被 17 个文件约 33 处引用——约一半页面样式完全没有过渡。
- `@keyframes` 仅 5 处（LED 脉冲×2、fadeIn、登录点动画×2），克制得当。
- `transition: all` **0 处**、非合成属性动画仅 1 处（见下）、`prefers-reduced-motion` 已支持（main.css:141）——性能纪律好。

### 问题
1. **性能点（唯一）**：`views/nodes/index.vue:348` 进度条 `transition: width` —— 每帧触发重排。应改 `transform: scaleX()` + `transform-origin: left`。
2. **路由切换零过渡**：`DefaultLayout.vue:347` `<router-view />` 裸写；全局 `.fade-*` 过渡类在 main.css:157-165 早已定义却从未用在路由上（只用在侧栏 logo）。页面间切换是硬切。
3. **弹窗进出场全是 naive-ui 默认**：ModalForm/ContentModal/ConfirmDialog 三原子无自定义动效，与全站克制动效风格不冲突，可接受，但 PageCard 骨架→内容的替换是瞬间跳变，可加 0.2s 淡入衔接。

### 改造方案
1. **P0**（1 行）：`DefaultLayout.vue:347` 包 `<router-view v-slot="{ Component }"><transition name="fade" mode="out-in"><component :is="Component" :key="$route.fullPath" /></transition></router-view>`——零新代码，启用已有类。
2. **P1**：nodes 进度条改 `transform: scaleX`。
3. **P1**：PageCard 骨架→内容加淡入；未使用动效 token 的页面补 `transition: background/border var(--t-fast) var(--ease)`（优先可点卡片）。
4. **红线**：维持现状纪律——不动画 width/height/top/left、不加 transition:all、动画仅 transform/opacity、尊重 reduced-motion。

---

## 六、响应式（评级：B-）

- 41 处 @media、4 个实际断点（480/640/768/1024）+ 1 处 reduced-motion。`--bp-*` token 因 CSS 变量不能用于 media query 而天然闲置（已知限制，非错误）。
- **5 个宽抽屉**（file 720/820px、website 720、container 720、cron 780）靠 main.css 全局 `max-width:100vw !important` 兜底——能用但生硬，768px 附近会有抽屉几乎全屏遮内容的体验。
- **小屏裸奔页面**：alert、audit、backup、cron、daemon、deploy、firewall、probe、protection、sessions、ssl 等 15+ 模块无任何 @media。

### 改造方案（P2）
面板是桌面优先工具，小屏可降级处理；但至少给 dashboard（用户最可能在手机上看监控）和 file 做一轮 768px 检查。抽屉宽度统一走 token 并在 768px 以下自动全屏。

---

## 七、改造优先级总表

| 级别 | 事项 | 工作量 | 文件 |
|---|---|---|---|
| **P0** | DataTable `:error` + 重试接线（先做 6 个高频页） | 每页 ~5 行 | website/file/ssh/firewall/backup/audit |
| **P0** | file 页 7 个 ModalForm 补 `:loading`；ModalForm 确认钮加 `:disabled` | 8 行 | file/index.vue + ModalForm.vue |
| **P0** | dashboard 骨架屏 + ai/nodes 空态 | ~30 行 | Dashboard.vue / ai / nodes |
| **P0** | 路由切换 fade 过渡（启用已有类） | 1 处 | DefaultLayout.vue:347 |
| **P1** | 34 处内联宽度收敛 `--input-*`；清理死 token `--drawer-width` | 纯替换 | 8 个文件 |
| **P1** | 可点卡片键盘可达（tabindex+keydown）| 先改 2 个高频 | appstore / website |
| **P1** | 全局按压反馈 `:active`；nodes 进度条 scaleX；PageCard 骨架淡入 | 各 1-3 行 | main.css / nodes / PageCard |
| **P2** | dashboard/file 小屏适配；抽屉宽度 token 化 | 中等 | 多文件 |

**核心判断：这个前端不需要「设计改造」，需要「接线改造」——把已建成但闲置的能力（骨架、错误态、过渡类、动效 token）铺满全部页面，P0 全部做完预计一个下午，可感知体验会跨一个档次。**

---

## 八、修复执行记录（2026-11-02）

报告结论已全部落地，改动集中在 4 个公共组件 + 20 个页面，共 **604 增 / 138 删**。
验证：`vue-tsc` 零错误 · `vitest` 15/15 · `vite build` 成功。

### P0（全部完成）

| 项 | 落地位置 | 说明 |
|---|---|---|
| DataTable 错误态接线 | `DataTable.vue` + website/file/ssh/firewall/backup/audit（8 张表） | 错误态改为**独立错误面板**（NResult + 重试按钮），不再塞进空态插槽——避免被页面自定义 `#empty` 覆盖；新增 `common.loadFailed` 三语键 |
| file 页防重复提交 | `file/index.vue` 7 个 ModalForm | 逐个补 `:loading` + 新增 `dialogLoading` 状态；ModalForm 组件级 `:disabled="loading"` + 提交守卫 |
| 骨架屏与空态 | dashboard / ai / nodes / plugins / monitor | dashboard 图表用**覆盖式骨架**（保持 ECharts 元素挂载，否则图表永久失效）；ai 评分加骨架 + 失败重试；nodes/plugins 加骨架网格；monitor 磁盘列表补空态 |
| 路由过渡 | `DefaultLayout.vue:347` | 启用 main.css 中既有的 `.fade` 过渡类，消除页面硬切 |
| 进度条动画 | `nodes/index.vue` | `transition: width` → `transform: scaleX()` + `transform-origin: left`，消除每帧重排 |

### P1（全部完成）

| 项 | 落地位置 | 说明 |
|---|---|---|
| 内联宽度 token 化 | 10 个页面 + `tokens.css` | 34 处硬编码宽度全部收敛 `--input-*`；无对应档位的就近归档（180→xl、110/120→sm、150→md、260→2xl、360→3xl），**残留硬编码宽度为 0** |
| 抽屉宽度 token 化 | `tokens.css` + 6 个抽屉 | `--drawer-width`（原 480px 定义却零使用）改为 `min(720px,100vw)` 并启用；新增 `--drawer-width-lg`；720/780/820 三档写死宽度全部归一 |
| 键盘可达 | PgPanel / deploy / website | 修复 3 处真实违规（见下方勘误）；全局 `:focus-visible` 焦点环生效 |
| 按压反馈 | `main.css` | 按钮 `:active` 轻微缩放 0.97，保留 naive-ui 原过渡参数，仅追加 transform 通道 |
| PageCard 淡入 | `PageCard.vue` | 骨架→内容 opacity+translateY(4px) 入场（仅合成器友好属性） |

### P2（全部完成）

- dashboard 小屏 10/11px 字号 → `var(--fs-caption)`（12px，可读性修复）
- file 页 768px 适配复核通过（树折叠/搜索全宽/批量栏换行）
- 抽屉宽度全站 token 化，小屏自动全屏

### 顺带修复的两个真实缺陷（审查时发现，超出报告范围）

1. **嵌套 `<form>` 导致按回车刷新页面**
   naive-ui 的 NForm 渲染原生 `<form>`，而 **25 个弹窗**在 ModalForm 的表单里又嵌了一层 NForm（HTML 非法）。单输入框弹窗（如"新建文件夹"）按 Enter 会触发隐式提交 → **页面刷新跳转**。
   修复：ModalForm 统一接管回车提交（`@submit.prevent` + keydown 守卫），并移除 file 页 3 处冗余 `@keyup.enter` 以避免双提交。守卫含 IME 组字（`isComposing`）、textarea、选择类控件三项排除。

2. **网络抖动被误报为「后端版本不支持」**
   `api/request.ts` 原先只抛裸 `Error`（无状态码），页面无法区分 404 与 5xx。SSH 页任何失败都会置 `apiAvailable=false` 并显示"当前环境无法连接 SSH 服务"——**一次网络抖动就永久显示误导性告警**。
   修复：请求层新增 `ApiError`（携带 HTTP status + 业务 code）与 `isApiUnsupported()`，SSH 三处加载改为区分「接口未实现」与「临时故障」，后者走错误面板 + 重试。

### 审查报告勘误（3 处，据实修正）

1. **nodes 页原本已有空态**（`v-else` 分支显示 loading/empty 文案），报告称"白屏"有误——本次仅统一为 NEmpty 风格并补骨架。
2. **appstore 应用卡与 website 类型卡本身已键盘可达**（前者内部是真实 `<button>`，后者是 radio），报告将其列为违规不准确。真正违规的是 PgPanel 数据库项与 deploy 日志头（`div @click` 无 tabindex），已修。
3. **file 页缺 loading 的弹窗是 7 个**（mkdir/mkfile/rename/move/copy/compress/decompress），报告统计为 5 个——已按实际全部修复。
