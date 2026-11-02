# 旧值 → Token 迁移映射表

- 依据：docs/design-audit-2026-10-21.md §5.1 Token 体系提案
- 执行日期：2026-10-21（阶段一 / 阶段二随批次执行）
- 验收口径：grep 门（旧值残留 = 0，豁免表除外）

## 迁移映射

| 旧值 | 原次数 | 涉及文件 | 新 token | 阶段 | 状态 |
|---|---|---|---|---|---|
| `color: var(--faint)`（正文级 caption） | ~100 处 / 39 文件 | 全部 views + PageCard/StatCard 等 | `--text-quaternary` | 二 | ✅ 已迁移，残留 0 |
| 十色板 hex（`#1BA784` 等 10 色） | 10 | plugins/index.vue | `--chart-palette-1..10` | 二 | ✅ 已迁移（JS 侧改用 `var(--chart-palette-N)`） |
| `color: #fff` | 2 | ResetPassword.vue / ChangePassword.vue | `--on-color` | 二 | ✅ 已迁移 |
| `rgba(255,255,255,0.04)`（Menu hover） | 1 | App.vue | `--surface2`（经 useThemeTokens） | 二 | ✅ 已迁移 |
| `z-index: 999 / 1000` | 2 | DefaultLayout.vue | `--z-drawer` / `--z-modal` | 一 | ✅ 已迁移 |
| 浅色 `--btn-bg: var(--acc)` | 1 | tokens.css | `#147760`（白字对比 5.47:1）+ `--btn-bg-hover: #105c4a` | 一 | ✅ 已迁移（经 App.vue Button overrides 接线 Naive UI） |
| `width: min(Xpx, 92vw)` 弹窗硬编码 | 21 | 18 个 view 文件 | `--modal-width-sm/md/lg/xl` | 三 | ✅ 已迁移（ContentModal size 档） |
| `target="_blank"` 无 rel | 1 | InstallDialog.vue | 补 `rel="noopener noreferrer"` | 二 | ✅ 已迁移 |
| 12px/13px 图表字号散点 | ~15 | monitor/Dashboard | `--fs-caption`（`fsPx()` 换算） | 四 | ✅ 已迁移，残留 0 |
| 硬编码动效时长 | 3 | Login / ChangePassword / nodes | `--t-fast` / `--t-normal` | 五 | ✅ 已迁移（2 处循环动画豁免注释） |
| 硬编码 px 间距（margin/padding/gap） | ~55 | 19 文件（ChangePassword/InstanceListTab/ResetPassword/Dashboard/container/sessions/SecurityExtras/main.css 等） | `--space-0..8` | 五 | ✅ 已迁移，残留 0（20px→space-5 档前就近、14px→space-4、28/40px→space-6） |
| `'#fallback'` 模式（`tk.xxx || '#…'`） | ~60 | monitor/terminal/Dashboard/App 等 | 维持 | — | 合规降级写法，无需迁移 |
| 字阶 `--fs-hero` / `--fs-stat` | 3 使用点 | StatCard / Login | `--fs-h1`（页面主标题/大数值，1.75rem）；Login 统计数字 → `--fs-title` | 剩余批 | ✅ 已合并，8 档 → 7 档（D03）；残留 0 |

## 对比度修正记录（WCAG 2.x 实算）

| 组合 | 修正前 | 修正后 | 结果 |
|---|---|---|---|
| 浅色 caption：`--text-quaternary` on `--paper` #fff | 3.02:1（旧 --faint #8b95a8） | **5.22:1**（#626d80） | ✅ AA 正文达标 |
| 浅色 caption：`--text-quaternary` on `--surface2` #f0f3f8 | ~2.7:1 | **4.70:1** | ✅ AA 正文达标 |
| 深色 caption：`--text-quaternary` on `--bg` #0e1117 | 3.35:1（旧 --faint #5c6678） | **6.18:1**（#8a94a6） | ✅ AA 正文达标 |
| 深色 caption：`--text-quaternary` on `--surface` #1a2030 | — | **4.85:1** | ✅ AA 正文达标 |
| 浅色主按钮：#fff on `--btn-bg` | 3.04:1（#1ba784） | **5.47:1**（#147760） | ✅ AA 正文达标 |
| 深色主按钮：`--btn-fg` #0e1117 on `--btn-bg` #1ba784 | 6.27:1 | 6.27:1（维持） | ✅ 达标 |

## 职责边界（拆分说明）

- `--faint`：**装饰性**弱信息专用（LED offline 底色、滚动条 hover、终端 bright-black），不承载正文，不受 WCAG 文本对比度约束。
- `--text-quaternary`：**可读 caption** 专用（PageCard/StatCard 副标题、hint、说明文字、面包屑分隔符），浅/深双主题均 ≥4.5:1。
- 新增文本弱色时必须二选一：装饰用 `--faint`，正文用 `--text-quaternary`；不得再造第三档。

## 豁免表

| 位置 | 豁免内容 | 理由 |
|---|---|---|
| main.ts BLOCKED 安全页 | 内联 hex `#0e1117` / `#8b8f98` | 在 token/main.css 加载前执行（挂载前拦截），无法引用 CSS 变量；隔离页不属主题体系管辖 |
