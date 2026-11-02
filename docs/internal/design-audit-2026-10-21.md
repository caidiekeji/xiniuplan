# 前端设计全面审查报告 — 云枢面板 CloudPivot（26 维度全量）

## 审查元数据

- 审查日期：2026-10-21 · 口径版本：measurement_standard v4
- 对象：frontend/src（51 个 Vue 文件 / 9 个共享组件 / 33 个 locale 模块 × 3 语言 / tokens.css 317 行）
- commit 基线：工作区当前状态（2026-10-20 i18n 全量整改后）
- 静态扫描范围：`frontend/src`，排除 node_modules / dist（measurement_standard v4 排除路径口径）
- 运行时环境：dev server（vite 5173）可启动；后端（:8888）未启动 → 页面被安全入口探测拦截（main.ts `detectPanelBase` 返回 BLOCKED 渲染 404 页，属设计行为）
- 前次审计：docs/design-audit.md（2026-09-13，6 维度）；i18n 整改记录：overview.md（2026-10-20）

## 〇、材料完备度声明

**已审材料**：全量样式代码（tokens.css / main.css / 51 个 SFC）、9 个共享组件逐查、路由与布局骨架、主题系统（useThemeTokens / App.vue themeOverrides）、i18n 体系（check-i18n.py 实跑 2216 key 全绿）、部署实例 192.168.31.154 响应头（curl 实测）、scripts/check-i18n.py 质量门。

**本次启用维度**：D01~D23、D25、D26 共 25 维。裁剪理由：D24（设计还原度）无 UI 稿 / Redline 材料 → 整维【证据不足】，不进评级。

**证据不足项（需可运行全栈环境）**：

| 维度 | 未审项 | 取证动作 |
|---|---|---|
| D12 | 点击反馈延迟 <100ms 实测 | devtools Performance 面板录 3 次点击 |
| D13 | 弹窗焦点陷阱 / 关闭后焦点返回 | 启动 `go run` 后端 + 键盘 Tab 走查 5 个弹窗 |
| D14 | 列表→详情→返回滚动位置恢复 | 同上，抽 ssh/audit/file 3 页 |
| D15 | 骨架 vs 真实内容结构对比（PageCard loading 未接线页） | 双截图对比 |
| D16 | 首包 gzip 体积 / FCP / LCP | `vite build` + Lighthouse 跑分 |
| D20 | 375 / 768 断点走查截图 | 浏览器视口 375/768/1280 三档截图 |
| D21 | 逐页暗色截图对照（无漏改白块） | 部署实例 + 测试账号逐页截图 |
| D22 | 读屏（NVDA）核心流程走查 | 读屏模拟 |
| D26 | CSP `unsafe-eval` 必要性 | 构建产物 grep `new Function` |

**关键事实**：2026-10-20 已有 9 页浏览器实证记录（overview.md，中/英/日三语巡检），本报告 D15/D23 部分项引用该记录作为二手证据，已标注。

---

## 一、问题全景统计表

按主维度计数（measurement_standard v4 评级；跨维度引用不重复计分）。「遗留基线」= 项目已上线运营，阈值双列呈现，基线只影响排期不改定级。

| 维度 | P0 | P1 | P2 | P3 | 主问题摘要 |
|---|---|---|---|---|---|
| D01 Token | 0 | 0 | 2 | 2 | 插件十色板硬编码；菜单 hover rgba 不适配浅色 |
| D02 留白 | 0 | 0 | 1 | 0 | 55 处 px 间距散布 19 文件 |
| D03 排版 | 0 | 0 | 1 | 0 | 字阶 8 语义档 > 6 级上限 |
| D04 色彩 | 0 | 1 | 0 | 0 | 对比度 3 组不达标（合并 1 项计） |
| D05 图标 | 0 | 0 | 1 | 1 | 双图标体系并存；尺寸档未收敛 |
| D06 组件 | 1 | 1 | 2 | 1 | 弹窗三实现并存；tiny 触控热区 |
| D07 复用结构 | 0 | 0 | 1 | 0 | NDataTable 直用 3 处绕过封装 |
| D08 API | 0 | 0 | 1 | 0 | base 层组件 0/9 有文档 |
| D09 动效 | 0 | 0 | 0 | 1 | 3 处硬编码时长 |
| D10 图表 | 0 | 0 | 0 | 2 | 无文字摘要；0 值边界待验证 |
| D11 噪音 | 0 | 0 | 1 | 0 | 插件页装饰彩板非语义色 |
| D12 交互 | 0 | 0 | 0 | 2 | 无 Undo；Toast 无合并策略 |
| D13 键盘 | 0 | 0 | 0 | 1 | 无命令面板（B 端建议项） |
| D15 状态 | 0 | 0 | 2 | 1 | 区域级 error 态无统一模式；骨架未推广 |
| D16 加载 | 0 | 0 | 0 | 1 | 性能预算未定义 |
| D17 IA | 0 | 1 | 0 | 0 | 24 项菜单平铺无分组 |
| D19 URL | 0 | 1 | 0 | 0 | 筛选/分页/Tab 全量不 URL 化 |
| D20 响应式 | 0 | 0 | 1 | 0 | safe-area 未处理 |
| D21 暗色 | 0 | 0 | 0 | 1 | 未接 prefers-color-scheme |
| D22 无障碍 | 0 | 0 | 0 | 1 | 页面缺 h1 / landmark |
| D25 可维护性 | 0 | 0 | 0 | 1 | 移动端表格样式重复定义 |
| D26 安全 | 0 | 1 | 1 | 1 | 外链缺 rel；token 明文存储未文档化 |
| **合计** | **1** | **5** | **14** | **18** | 38 项 |

---

## 二、P0 / P1 问题详情（逐条）

### P0-01 弹窗语义三种实现并存（D06）

- 【现象】同一「弹窗」语义存在 3 种实现：ModalForm（表单弹窗，9 使用点）、ConfirmDialog（确认弹窗）、裸 NModal 直用 21 处 / 18 文件（backup×3、DockerResources×2、sessions×2、TOTPTab×2、website、InstalledTab、daemon、InstallDialog、deploy、ssl、probe、Compose、Login、container、InstanceListTab、RedisPanel、ExecTerminal）。
- 【证据】grep `<NModal` 40 处（含 3 个组件内部自用，views 直用 21 处）；grep `<n-modal` 0 处。measurement_standard v4 §3.1：同语义组件实现 >1 且总使用点 >3 → P0。
- 【甄别】非复制改造——21 处直用多为「内容展示弹窗」（日志查看 / 终端 / 向导），ModalForm 不覆盖该场景；根因是 base 层缺第三个原子组件。
- 【影响面】弹窗宽度、关闭行为、ESC/遮罩、移动端钳制、焦点管理在各页自行拼装，行为漂移风险随页面数线性增长（如 ExecTerminal 用 `min(900px, 94vw)`、RedisPanel 用 `min(420px, 92vw)`，均绕过 `--modal-width-*` token）。
- 【修复方向】新增 `ContentModal`（内容弹窗 base 组件，宽度接 `--modal-width-md/lg`，统一关闭行为与焦点管理），21 处直用点按 preset 归口：dialog 简单确认 → ConfirmDialog；表单 → ModalForm；内容 → ContentModal。
- 【验收标准】views 层 `<NModal` 直用 = 0；弹窗宽度全部走 `--modal-width-*` token；ESC / 遮罩关闭行为走查一致。
- 【阶段】三。

### P1-01 对比度 3 组不达标（D04）

- 【现象与证据】（WCAG 2.x 相对亮度公式实算，token hex 为输入）
  1. 浅色 `--faint` #8b95a8 on `--paper` #ffffff = **3.02:1**；on `--surface2` #f0f3f8 ≈ **2.7:1**。`--faint` 用于 PageCard 副标题、StatCard 副标题等 12px（fs-caption）正文级文本（PageCard.vue:72、StatCard.vue:128）。AA 要求正文 ≥4.5:1，差 1.5。
  2. 深色 `--faint` #5c6678 on `--bg` #0e1117 = **3.35:1**（tokens.css:237）。差 1.15。
  3. 浅色主按钮：`--btn-fg` #ffffff on `--btn-bg` #1ba784 = **3.04:1**（tokens.css:167-168）。按钮文本 14px 非大字（large text 阈值 18.66px bold），差 1.46。深色同组合 = 6.27:1 达标 → 缺陷仅在浅色侧。
- 【复现】取 tokens.css:34/237/167-168 色值代入 `L = 0.2126R+0.7152G+0.0722B`（线性化后），ratio = (L1+0.05)/(L2+0.05)。
- 【违反标准】WCAG 2.1 AA 1.4.3；measurement_standard v4 §3.2「正文对比度 <4.5:1」。
- 【修复方向】浅色 `--faint` 加深至 #6b7688 量级（≥4.5:1）；深色 `--faint` 提亮至 #7a8598 量级；浅色按钮改用 `--acc-hover` #147760 作按钮底（白字对比 5.5:1）或加深 `--acc` 的按钮专用档。注意 `--faint` 同时承担「装饰性弱信息」与「正文级 caption」两种职责，建议拆分 `--text-quaternary`（可读 caption）与 `--faint`（装饰）。
- 【验收标准】重算全部文本/背景组合 ≥4.5:1（正文）/ ≥3:1（≥18.66px bold），输出对比度对照表进 CI。
- 【阶段】二。

### P1-02 全部列表页筛选 / 分页 / Tab 状态不 URL 化（D19）

- 【现象】37 个列表页的分页、筛选、排序、Tab 全部存于组件内存。全项目 `route.query` 仅 3 处且均为登录 redirect / 重置 token（Login.vue:108/141、ResetPassword.vue:13）。
- 【证据】grep `route.query|router.replace` views 目录 → 3 处；grep `useRouteQuery` → 0 处。
- 【复现】审计页筛选后按 F5 → 筛选丢失、回到第 1 页；SSH 页切 Tab 后刷新 → 回默认 Tab；复制 URL 给他人 → 对方看到默认视图。
- 【违反标准】dimensions.md D19 检查项「筛选状态 URL 化 → 进 query 参数」「Tab 状态保留」「深链可达」全项违反。
- 【影响面】37 页全量；运维高频场景（审计检索、日志定位）刷新即丢现场。
- 【修复方向】封装 `useQueryState(key, default)` composable（读写 route.query + router.replace），优先接入高频 6 页（audit / ssh / file / container / backup / website），其余按迭代铺开。
- 【验收标准】抽样 6 页：筛选 + 分页 + Tab 后刷新 URL 状态完整还原；复制链接他人打开看到相同视图。
- 【阶段】四。

### P1-03 侧边栏 24 项菜单平铺无分组（D17）

- 【现象】menuOptions 24 项顺序平铺（DefaultLayout.vue:58-83），无分组标题、无分隔。
- 【证据】DefaultLayout.vue:58-83 数组 24 个条目；导航层级深度 1（合规），但单列 24 项超出扫描辨识上限（行业基准 7±2 分组 × 5~7 项）。
- 【影响面】新用户定位成本高；高频页（dashboard）与低频页（audit/sessions）权重无差异；侧栏滚动后当前项可视性差。
- 【修复方向】按运维动线分 5 组：概览（dashboard/monitor/probe）、应用（website/deploy/database/container/appstore/plugins）、运维（file/terminal/cron/daemon/backup）、安全（firewall/protection/ssh/alert/ssl）、系统（users/audit/sessions/nodes/ai/settings）。NMenu 原生支持 group label，无需自研。
- 【验收标准】菜单分组 ≤7 组，组内 ≤6 项；三语言 menu.ts 补 group label key。
- 【阶段】四。

### P1-04 外链缺 rel="noopener noreferrer"（D26）

- 【现象】唯一一处 `target="_blank"` 无 rel。
- 【证据】InstallDialog.vue:261：`<a :href="installResult.access_url" target="_blank" class="mono">`。
- 【违反标准】dimensions.md D26「noopener noreferrer」；Tabnabbing 风险面（access_url 来自应用目录数据，非完全可信输入）。
- 【修复方向】补 `rel="noopener noreferrer"`。
- 【验收标准】grep `target="_blank"` 全部携带 rel，纳入 CI grep 门。
- 【阶段】二（随 P0 批次顺手修，5 分钟工作量）。

### P1-05 移动端触控热区 47 处 tiny 按钮 < 44×44px（D06）

- 【现象】`size="tiny"`（Naive 高度 24px）47 处 / 16 文件（RedisPanel 16、file 6、DatabaseManagerTab 3、nodes 3…）。
- 【证据】grep `size="tiny"` 47 处；tokens.css `--size-sm: 32px`；main.css:196-199 仅对 `.n-data-table .n-button` 强制 min-height 32px，表格外 tiny 按钮无补偿。
- 【对照阈值】measurement_standard v4 §3.2「触控热区 <44×44px」→ P1（移动端可点元素口径）。遗留基线说明：桌面鼠标场景 24px tiny 是面板类惯例（1Panel 同款），双列呈现——目标标准 44px（移动端）、遗留基线 24px（桌面端维持，移动端补齐）。
- 【修复方向】移动端断点下 `size="tiny"` 按钮 min-height 提至 32px（对齐既有表格按钮决策）；≥44px 仅强制操作列等高频点击位。
- 【验收标准】375 视口走查：高频操作按钮热区 ≥32px（遗留基线口径）/ 44px（目标口径双列记录）。
- 【阶段】四。

---

## 三、分维度审查详情

### A 组 · 设计体系基础

**D01 Token 统一性 — P2×2 / P3×2**

- token 体系 317 行三层（基础 60-30-10 / 语义 / 组件级），双主题 + ≤768/≤480 双断点重写，命名语义化（--err 非 --red-500）。62 处 SFC hex 中约 60 处为 `tk.xxx || '#fallback'` 模式（token 读取失败兜底，monitor 18、terminal 11、Dashboard 8、App 8、ExecTerminal 3、TOTPTab 2），属合规的降级写法。
- P2①：plugins/index.vue:33-34 十色板 10 个 hex 完全脱离 token（`'#1BA784', '#2DD4A8'…`），且与 token 值形似而大小写不一。建议抽 `--chart-palette-1..10` 或直接引用 token 变量数组。
- P2②：App.vue:45 `itemColorHover: 'rgba(255,255,255,0.04)'` 硬编码——该值只在深色下可见（浅色 bg2 #eceff4 上叠加 4% 白 ≈ 无感），浅色主题菜单 hover 反馈近似丢失。改用 `tk.surface2 || fallback`。
- P3①：login 两处 `color: #fff`（ResetPassword.vue:180、ChangePassword.vue:245）应为 `var(--on-color)`。
- P3②：main.ts:44-47 BLOCKED 安全页内联 2 hex——隔离页可豁免，建议注释标注豁免理由。
- 间距 55 处 px / 19 文件（详见 D02）；字号 21 处 / 10 文件（集中在 12/13px 图表与终端场景）；圆角 8 处 / 阴影 7 处 / 动效时长 3 处——收敛度均优于阈值。z-index 4 文件 5 处，999/1000 与 --z-modal 自洽。

**D02 留白 — P2×1**

- 内容区 `--space-5`(24px)、PageCard 内 `--space-5`、max-width 1400px、移动降级链完整（DefaultLayout.vue:444-451 + media queries）。
- P2：55 处硬编码 px 间距散布 19 文件（ChangePassword 6、InstanceListTab 5、ResetPassword 4、Dashboard 4、container 4、sessions 4、SecurityExtras 4…），其中含与 token 等值（12/16/24）的冗余写法。收敛至 token 后离散值可控制在 8 档内。

**D03 排版与骨架 — P2×1**

- Type Scale 8 个语义档（display/hero/stat/headline/title/body/caption/mono）+ 行高配对 + 字重 3 档。P2：档数 8 > 6 级上限——但每档有明确语义命名且 stat/mono 为面板特色档，建议合并 hero/stat（差值仅 0.5rem）或豁免记录。行长：表单/表格场景不适用行长规则；`.mono` + tabular-nums 仪表盘刻度感完整（main.css:46-51）；数字格式统一出口 utils/format.ts（8 处重复已收敛）✅。

**D04 色彩搭配 — P1×1**（见 P1-01）

- 结构：60-30-10 完整，主色占比克制，功能色 ok/warn/err/info 语义无错位（上轮 warn 青绿→琥珀缺陷已修，tokens.css:39 注释明确）。色相数 ≤3（青绿/蓝/琥珀，不含功能色）✅。

**D05 图标与资源 — P2×1 / P3×1**

- 专属图标集 PanelIcons.ts 已落地（上轮唯一结构性缺口已修复）：10 个器物图标，stroke 1.7 统一、圆头、currentColor、aria-hidden（PanelIcons.ts:13-21）✅。
- P2：专属集（10 个）与 @vicons/ionicons5（34 文件引用）双体系并存——侧栏核心入口用专属、页内操作用通用。现状可用，但需一页决策表明确「哪里必须专属」。P3：图标尺寸档 16/20/34 未收敛至 2~3 档声明。

### B 组 · 组件体系

**D06 全组件盘点 — P0×1 / P1×1 / P2×2 / P3×1**

- 共享组件 242 使用点 / 37 文件（ssh 21、monitor 14、settings 15、file 14…），零裸写表格/表单/空态。组件状态完备性抽查：DataTable（loading NSpin / empty NEmpty / Go nil slice 兜底 DataTable.vue:42-44）、ModalForm（loading / 防重复提交）、ConfirmDialog（LED 语义灯 + loading）、StatCard（loading 骨架 / hover lift / 键盘态联动 main.css:120-125）✅。
- P0：弹窗三实现（见 P0-01）。P1：tiny 触控热区（见 P1-05）。
- P2①：NDataTable 直用 3 处绕过封装（protection/index.vue:291、probe/index.vue:389、Compose.vue:269）——丧失统一空态/表头样式。P2②：NCard 直用 ~12 处，其中 AppStoreTab:181 / plugins:156 / nodes:187 用 NCard 做卡片网格而非 PageCard 语义，卡片底色/圆角来源分裂。
- P3：NSwitch 14 处 / 11 文件的危险开关二次确认覆盖未逐点验证（待验证清单 #4）。

**D07 组件复用化结构 — P2×1**

- 三层雏形存在：components/（base 9 个）+ views 内业务子组件（database 5 Tab、users 3 Tab、appstore 3 子组件）✅；复制型使用 0（formatTime/formatBytes 8 处重复已收敛为 utils/format.ts 单一出口）✅；样式定制全走 token / :deep 规范 ✅。
- P2：3 处 NDataTable 直用（同 D06，主维度计 D06，此处引用不重复计分）+ SearchForm 仅 3 页使用（overview.md 已声明「保留封装、不强制推广」的决策记录 ✅——按规则算有理由豁免）。

**D08 组件 API 与文档 — P2×1**

- props 设计：变体全 props 化（StatCard `accent` 6 枚举、StatusTag `status` 5 枚举、DataTable 12 props 含 virtualScroll/scrollX）✅；默认值合理（DataTable pageSize 20、striped true）✅；事件命名 `update:xxx` + `handleXxx` 全站统一 ✅。
- P2：base 层 9 组件文档 0/9——无 props 表 / 使用示例 / 变更记录。建议 README.md 一页表（props/emits/默认值/示例）即可达标，无需 storybook。

### C 组 · 动效与图形

**D09 动效与动画 — P3×1**

- motion token 分档（--t-fast/normal/slow + --ease/--ease-out）✅；`transition: all` 0 处（上轮 2 处已修）✅；布局动画 0 ✅；prefers-reduced-motion 全局通配降级（main.css:128-136）✅；动效功能性（LED 状态脉冲、路由 fade、按钮 lift）均服务反馈，无纯炫技 ✅。
- P3：3 处硬编码时长（Login.vue / ChangePassword.vue / nodes/index.vue 各 1）接入 token 即可。

**D10 SVG 与图表 — P3×2**

- ECharts 全部经 useThemeTokens 读 token（monitor/Dashboard），上轮缺陷（双色区分、真实时间轴、断档 null、假数据删除）全部确认已修 ✅。
- P3①：图表无文字摘要 / aria-label（数据无障碍通道缺失）。P3②：连接分布饼图 OTHER=0 时是否崩坏未验证（待验证清单 #5）。

**D11 视觉噪音 — P2×1**

- 单元素视觉手法 ≤2 种 ✅；LED 动效密度：同屏同时运动 ≤1 类（仪表 LED 恒动属状态语义，豁免）✅；阴影与层级一一对应（sm/md/lg 三级）✅。
- P2：plugins 十色板（D01 主维度计分，此处引用）属于「彩色不承载语义」的装饰噪音，插件字母头像场景建议收敛为灰阶+单强调色或语义色板。

### D 组 · 交互体验

**D12 交互细节 — P3×2**

- 破坏性保护：NPopconfirm / ConfirmDialog / dialog.warning 69 处 / 29 文件 ✅；忙态保护：ModalForm / SearchForm / 提交按钮 loading 覆盖 ✅；边界兜底：DataTable Go nil slice、request.ts 拦截器、ErrorBoundary 全局+app.config.errorHandler 双层 ✅。
- P3①：无 Undo 机制（删除靠确认弹窗，无撤销）——面板类可接受，列为打磨。P3②：Toast 无同类合并策略（Naive 默认堆叠），批量操作连发时会堆叠轰炸。

**D13 键盘与焦点 — P3×1 + 运行时项证据不足**

- 全局 `:focus-visible`（main.css:113-117）✅、可点击卡片键盘态（main.css:120-125）✅、Naive UI 弹窗焦点陷阱内建 ✅、tabindex 误用 0 处 ✅。
- P3：无命令面板（Cmd+K）——24 功能入口的 B 端工具建议具备（配合 P1-03 分组一起做收益最大）。焦点陷阱实测 / 快捷键冲突检测：证据不足（见声明）。

**D14 滚动行为 — 证据不足为主**

- 布局滚动架构正确：侧栏独立、内容区 `calc(100vh - header)` 自滚（DefaultLayout.vue:444-451 注释明确设计意图）✅；分页制无无限滚动 ✅。滚动位置恢复 / 弹窗滚动穿透：证据不足（取证动作见声明）。

### E 组 · 状态与容错

**D15 状态完备性 — P2×2 / P3×1**

- 四态：loading（NSpin 39 处 + PageCard 骨架基建）/ empty（DataTable 内置 NEmpty）/ success ✅；ErrorBoundary 模块级隔离 ✅。
- P2①：error 态无统一区域级模式——请求失败全部降级为 message.error 裸 toast（368 处），页面区域无「失败 + 重试」态，列表加载失败后页面停留空白表格。建议 DataTable 增加 `error` prop + 重试插槽，或封装 useAsyncPage。P2②：PageCard loading 骨架已建（PageCard.vue:25-28）但多数页仍 NSpin 转圈（overview.md 遗留 #2 自认），骨架结构还原度因未接线无法对比。P3：NEmpty 有说明无行动入口（空态引导下一步动作缺）。

**D16 加载策略 — P3×1**

- 路由级懒加载 26 路由全量 ✅；monaco 按需（editor.api 内核 + 异步组件 CodeEditor，components/index.ts:12-14）✅；PWA + workbox 发版缓存（dist/sw.js 实存）✅；长任务（应用安装/部署）有日志流反馈 ✅。
- P3：性能预算未定义（首 JS <300KB gzip 等无声明）；FCP/LCP 未实测（证据不足）。

### F 组 · 信息架构与文案

**D17 信息架构 — P1×1**（见 P1-03）

- 其余项全过：核心功能 1 层直达 ✅；面包屑（DefaultLayout.vue:129-135，含长路径截断）✅；每页独立 title 同步（router/index.ts:213-216）✅；菜单名 = 页面内容（i18n 全量核对）✅。

**D18 UI 文案 — P3×1 + 待验证**

- i18n 2216 key × 3 语言全绿（check-i18n.py 实跑退出码 0）✅；naive-ui 内置文案随语言切换（App.vue:17-23，上轮 zhCN 硬编码 bug 已修）✅。
- P3：无统一 UI 词表（保存/确认/取消/删除按钮用语靠 common.* 约定，无文档固化）。称呼「您/你」一致性、时间显示规则（formatTime 24h 内相对？）未逐点验证（待验证清单 #6）。

**D19 URL 状态 — P1×1**（见 P1-02）

### G 组 · 环境适应

**D20 响应式与兼容 — P2×1 + 证据不足**

- 三层适配体系（token 层断点重写 / 布局层抽屉 / 全局组件钳制）覆盖面优于同类面板项目；弹窗钳制 `calc(100vw - 24px)`、表格按钮触屏补偿、tabs 横滚均已处理 ✅。
- P2：`env(safe-area-inset-*)` 0 处——移动浏览器全面屏底部横条可能遮挡页面底部操作按钮（移动端使用为次要场景，优先级随移动端定位调整）。375/768 断点实测：证据不足。

**D21 暗黑模式 — P3×1 + 证据不足**

- token 层完整双主题（语义色 / 阴影 / 组件 token 成对重写）+ Naive darkTheme + ECharts/xterm/monaco 全部接线 ✅；深色阴影策略（微光替代）✅；默认 dark（stores/app.ts:8）。
- P3：未接 prefers-color-scheme 跟随系统（仅手动切换）。逐页截图对照：证据不足。

**D22 无障碍 — P3×1 + 证据不足**

- 焦点可见 ✅；图标 aria-hidden ✅；img alt 1/1 ✅；纯色彩双通道：StatusTag = LED 灯 + 文字、趋势 = ↑↓ 箭头 + 色、ConfirmDialog = LED + 文案 ✅。
- P3：PageCard 标题 h2 起步，页面缺 h1 / landmark（header/nav/main 裸 div）。读屏走查：证据不足。

**D23 国际化就绪 — 达标（引用二手证据）**

- 3 语言 × 2216 key 全绿、MessageSchema 类型约束三语言 key 集合零差异、弹性布局容纳 1.5 倍膨胀（英文态 9 页巡检通过，overview.md 2026-10-20）✅；残留仅后端数据（范围外，已立项记录）。日期格式化是否走 Intl：待验证清单 #6。

### H 组 · 还原与维护

**D24 设计还原度 —【证据不足】**：无 UI 稿 / Redline，整维不评级。

**D25 可维护性 — P3×1**

- 死组件 0（9 组件全有使用点）✅；依赖健康：echarts/monaco/xterm/pinia 各一无重复 ✅；注释代码块 0（抽样）✅。
- P3：移动端表格压缩样式在 main.css:247-251 与 DataTable.vue:125-131 重复定义（padding/font-size 双份 !important），单点收敛到 DataTable。

**D26 前端安全 — P1×1 / P2×1 / P3×1**

- v-html / innerHTML / dangerouslySetInnerHTML：**0 处** ✅（XSS 注入面干净）；响应头实测（curl 192.168.31.154）：CSP / X-Frame-Options: DENY / nosniff / Referrer-Policy / Permissions-Policy 全配置 ✅。
- P1：外链缺 rel（见 P1-04）。P2：token 明文存 localStorage（auth.ts:22）——面板类产品常见决策但**风险决策未文档化**，建议在 docs/dev/design.md 记录「为何不用 HttpOnly cookie」（安全入口 + 内网场景权衡）。P3：CSP 含 `unsafe-eval`，Vite 构建产物为 runtime-only 不需要 eval，可尝试移除（取证动作见声明）。

---

## 四、竞品调研与方向选项

**品类确认**：服务器运维面板 = tool_class × b_admin 混合，主要载体桌面 Web，移动端为辅助查看。

**参考系事实**（该竞品确有此特性）：

| 产品 | 事实 |
|---|---|
| 1Panel（同类开源标杆） | 左侧分组导航、高密度表格、工具类绿色主色、无命令面板 |
| Cockpit（Red Hat） | 极简灰阶、卡片化概览、无品牌色面积 |
| Linear（tool_class 代表） | Cmd+K 命令面板、键盘优先、灰阶克制、动效服务空间关系 |

**可选方向**（取舍，最终由用户决策）：

- **方向 A · 运维效率型（对标 1Panel + 补键盘化）**：保持现有高密度紧凑，落地 P1-03 分组导航 + D13 命令面板 + D19 URL 状态。取舍：改动集中交互层，视觉不动，收益是高频运维效率；牺牲：暂不做视觉差异化的进一步深化。
- **方向 B · 工具流极简型（对标 Linear/Cockpit）**：弱化青绿在非交互位的面积（StatCard 数值色、sysbar），灰阶承担 90% 界面，青绿只出现在可点击/活动态。取舍：专业感与辨识度更向 Linear 靠拢；牺牲：现有「深夜数据中心」撞色签名弱化，与已有品牌语言有割裂成本。
- **方向 C · 签名深化型（维持现有语言，补完签名体系）**：保持 60-30-10 撞色，把 LED 状态语言推广到全部状态位（徽章、Tab 指示、列表行状态），专属图标集扩展到页内高频操作位。取舍：品牌一致性最强、改动最便宜；牺牲：不解决 P1-03/P1-02 这类交互效率问题（需与方向 A 合并执行）。

**推荐**：A + C 组合（交互效率问题用 A 的清单解决，视觉只做 C 的低成本深化），B 不建议——与既有品牌资产冲突。此为推断，附适配条件：团队 1-2 人维护规模、桌面为主场景。

---

## 五、新设计方案

### 5.1 Token 体系提案

现有体系已达标 90%，以下为补全项（非重构）：

| 项 | 现状 | 提案 |
|---|---|---|
| color | 语义完整 | 新增 `--text-quaternary`（可读 caption，≥4.5:1）与装饰用 `--faint` 分离；新增 `--chart-palette-1..10` 收编 plugins 色板；浅色按钮底色档 `--btn-bg` 改 #147760 |
| spacing | 9 档 ✅ | 维持 [2,4,8,12,16,24,32,48,64]，清退 55 处散布 px |
| typography | 8 档 | 合并 hero/stat → 7 档（或豁免记录）；新增 `--fs-h1` 供页面主标题 |
| radius / elevation | 3 档 / 3 级 ✅ | 维持 |
| motion | 3 档 2 曲线 ✅ | 维持，清退 3 处硬编码 |
| z-index | 3 层 ✅ | 新增 `--z-drawer: 900`（现 999 魔法值入 token） |

**旧值→token 迁移映射表**（按出现次数降序）：

| 旧值 | 次数 | 文件清单 | 新 token | 阶段 | 验收 |
|---|---|---|---|---|---|
| 12/16/24px 等值间距 | ~35 | 19 个 view 文件 | --space-2/4/5 | 二 | grep 间距 px = 0（豁免表除外） |
| '#fallback' 模式 | ~60 | monitor/terminal/Dashboard/App 等 | 维持（合规降级） | — | 无需迁移 |
| 十色板 hex | 10 | plugins/index.vue:33-34 | --chart-palette-* | 二 | 该文件 hex = 0 |
| color:#fff | 2 | ResetPassword:180 / ChangePassword:245 | --on-color | 二 | 同上 |
| rgba(255,255,255,0.04) | 1 | App.vue:45 | tk.surface2 | 二 | 浅色菜单 hover 可见 |
| width:'min(…px, 92vw)' | ~21 | 18 个 view 文件 | --modal-width-md/lg | 三 | views 层 NModal = 0 |
| 12px/13px 图表字号 | ~15 | monitor/Dashboard | --fs-caption（图表 ctx 统一读） | 四 | 字号散点 ≤3 档 |
| 硬编码动效时长 | 3 | Login/ChangePassword/nodes | --t-fast/normal | 四 | transition 硬编码 = 0 |

### 5.2 组件复用化方案

**去重矩阵**：

| 同语义实现 | 使用点 | 目标组件 | 动作 |
|---|---|---|---|
| 表单弹窗 = ModalForm | 9 | ModalForm | 维持 |
| 确认弹窗 = ConfirmDialog / NPopconfirm / dialog.warning | 69 | ConfirmDialog（页内常驻）/ dialog.warning（即时） | 维持双轨（现状合理） |
| 内容弹窗 = 裸 NModal | 21 | **新增 ContentModal**（base 层：宽度 token 化、关闭行为、焦点管理、移动端钳制内建） | 全量替换 |
| 表格 = DataTable / 裸 NDataTable | 242+3 | DataTable | 3 处直用替换 |
| 卡片 = PageCard / 裸 NCard 网格 | ~30+12 | PageCard（页级）/ 新增 `CardGrid`+`GridCard`（可选）或 NCard 声明豁免 | 先声明豁免边界 |

**三层结构**（现状微调即可，不推倒）：

- base：PageCard / DataTable / SearchForm / ModalForm / **ContentModal（新增）** / StatusTag / StatCard / ConfirmDialog / CodeEditor / ErrorBoundary —— 零业务逻辑 ✅（现状已满足）
- components：无独立层需求，维持平层 + views 业务子组件（规模不支撑三层目录，避免过度工程）
- business：database 5 Tab / users 3 Tab / appstore 3 子组件 ✅

**API 规则**：变体走 props 枚举（现状 ✅）；新增组件必须附 README props 表（D08 缺口）。

**states_checklist**：ContentModal 验收须含 default / loading（footer 按钮）/ ESC / 遮罩关闭 / 移动端钳制 / 焦点陷阱 / 关闭后焦点返回七项。

### 5.3 UI 文案词表 ★

| 类别 | 统一词 | 禁用变体 |
|---|---|---|
| 主操作 | 保存（配置类）/ 创建（新建资源）/ 安装（应用）/ 确认（通用） | 确定、OK、提交 |
| 危险操作 | 删除 / 停止 / 卸载 + 确认弹窗必须含后果句 | 移除、清空（除日志场景） |
| 错误文案 | 「{对象}{动作}失败：{原因}」+ 可重试时附「请重试」 | 裸错误码、Error 500 |
| 空态 | 「暂无{对象}」+ 首次使用附引导动作链接 | 空空如也、无数据 |
| 称呼 | 统一「您」或零人称（现状待验证后定） | 你/您混用 |

### 5.4 分五阶段迁移路径

| 阶段 | 内容 | 验收标准 | 回滚触发 | 回滚动作 |
|---|---|---|---|---|
| 一 | token 补全（--text-quaternary / --chart-palette / --z-drawer / --btn-bg 浅色档）+ 旧值映射表落 docs | check-i18n 同款 grep 门脚本覆盖上表全部旧值；token 文件 vue-tsc 通过 | 与现有构建冲突 | token 文件仅新增不替换 |
| 二 | P1 批次：对比度 3 组修正 + P1-04 外链 rel + plugins 色板 + #fff×2 + App.vue hover | grep `target="_blank"` 全带 rel；对比度对照表全过；视觉回归截图对比 | 页面视觉回归失败 | 按文件回滚，映射表保留 |
| 三 | ContentModal 新建 + 21 处 NModal 直用替换 + 3 处 NDataTable 收编 | views 层 `<NModal` = 0；弹窗走查七态全过 | 业务功能回归失败 | 旧 NModal 用法保留一个版本周期 |
| 四 | D19 useQueryState 接 6 高频页 + P1-03 菜单分组 + P1-05 移动热区 + D15 DataTable error 态 | 刷新还原走查 6 页全过；菜单 ≤7 组；375 视口热区达标 | 交互回归失败 | useQueryState 按页开关回退 |
| 五 | P3 打磨：命令面板、prefers-color-scheme、空态引导、词表落地、组件 README | 词表 grep 门 + README 9/9 | 用户反馈负面 | 按功能开关回退 |

---

## 六、待验证清单

| # | 疑点 | 所需证据 | 取证动作 | 回填后评级 |
|---|---|---|---|---|
| 1 | 弹窗焦点陷阱 / 关闭后焦点返回是否 Naive 默认可用 | Tab 走查录屏 | 后端启动后 5 弹窗键盘走查 | 若失效升 P1 |
| 2 | 列表滚动位置恢复 | 复现记录 | ssh/audit/file 返回测试 | 若丢失维持 P3 打磨 |
| 3 | 首包体积 / FCP / LCP | Lighthouse 报告 | build + 跑分 | >300KB gzip 升 P2 |
| 4 | 14 处 NSwitch 危险开关二次确认覆盖 | 走查记录 | settings/protection 逐开关 | 缺失升 P1 |
| 5 | 连接分布饼图 0 值边界 | 截图 | 全部连接关闭时打开 monitor | 崩坏升 P2 |
| 6 | 「您/你」称呼一致性；formatTime 是否 Intl | grep + 代码读 | 全量 locale 抽查 | 不一致记 P3 |
| 7 | CSP unsafe-eval 可移除性 | 构建产物 grep | `new Function` 检索 | 可移除记 P3 建议 |
| 8 | 375/768 断点 + 暗色逐页截图 | 截图组 | 部署实例 + 测试账号 | 白块/溢出升 P1 |

## 七、优点（≤20% 篇幅，每条附证据）

1. **Token 体系成熟**：317 行三层 + 双主题 + 双断点成对重写（tokens.css 全文），60 处 JS 侧消费全部经 useThemeTokens 读 token，仅留合规 fallback。
2. **组件复用纪律**：共享组件 242 使用点 / 37 文件、裸写表格 0、`transition: all` 0、v-html 0、复制型工具函数 0（format.ts 单一出口）。
3. **i18n 工程化**：2216 key × 3 语言全绿 + MessageSchema 类型约束 + naive locale 联动（check-i18n.py 实跑 + App.vue:17-23）。
4. **安全基本面**：响应头 CSP/XFO/nosniff 全配置（curl 实测）、XSS 注入面 0、ErrorBoundary + errorHandler 双层兜底（main.ts:69-88）。
5. **性能意识**：26 路由全懒加载、monaco editor.api 精简按需（CodeEditor.vue:2-14）、PWA 发版缓存、reduced-motion 全局降级（main.css:128-136）。
6. **签名语言落地**：LED 状态灯 + 10 个专属器物图标（上轮缺口已修复），命名注释体现设计意图（PanelIcons.ts:1-11）。
