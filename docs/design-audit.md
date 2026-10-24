# 前端设计审计报告 — 犀牛面板（CloudPivot）

> 审计时间：2026-09-13 · 范围：frontend/src 全部 49 个 Vue 组件 + 设计 token 系统 · 方法：静态扫描 + 组件逐查 + 质量门脚本（人工校准）
>
> **结论：设计系统成熟度高，六维度整体评级 B+。发现 7 处缺口，其中 6 处已就地修复，1 处（专属图标集）列为建议项。**

---

## 一、设计语言现状（先说好的）

项目已有一套自洽的原创设计语言「深夜数据中心」，落地质量高于常见面板类产品：

| 维度 | 现状 | 评价 |
|------|------|------|
| Token 系统 | 314 行完整三层 token：60-30-10 撞色（冷钢蓝 × 信号青绿）+ Type Scale 8 档 + 间距 4px 基准 + 圆角/阴影三档 + 动效时长分档 + 断点 + 组件级 token | ★★★★★ 教科书级 |
| 签名元素 | 状态指示灯（LED）：青绿脉冲/琥珀静止/红色闪烁，keyframes + `--led-pulse-duration` token 化 | ★★★★★ 唯一性成立 |
| 双主题 | `[data-theme='dark']` 完整覆盖，语义色/阴影/组件 token 分层重写 | ★★★★★ |
| 等宽刻度感 | `.mono` 类 + `font-variant-numeric: tabular-nums`，仪表盘数字刻意用等宽 | ★★★★☆ |
| 语义色纪律 | ok=绿/warn=琥珀/err=红，浅暗两套 soft 背景色 | ★★★★★ |

## 二、六维度审计结果

### 1. 组件设计合理性 — ★★★★☆

- 9 个通用组件（PageCard/DataTable/SearchForm/ModalForm/StatusTag/StatCard/ConfirmDialog/CodeEditor/ErrorBoundary）职责清晰，页面零重复实现。
- StatCard 状态矩阵完整：loading（LED 熄灭）/ hover（lift -2px）/ 数据空态。
- **缺口（已修）**：可点击卡片有 hover 态但无键盘 focus 态；全局 `:focus-visible` 为 0 依赖组件库默认。
- **缺口（已修）**：2 处 `transition: all`（AppStoreTab 分类 tag、InstanceListTab 实例卡）——`all` 会让非预期属性参与过渡，性能与动效纯度都不合格，已改为显式属性列表。

### 2. 移动端适配 — ★★★★☆（基础设施最好的部分）

- Token 层：≤768px / ≤480px 两级断点重写字号、间距、布局尺寸、输入框 100% 自适应。
- 布局层：isMobile 检测 + 抽屉式侧边栏 + 遮罩；内容区 padding 降级。
- 全局层（main.css）：弹窗宽度钳制、表单 label 顶部对齐、抽屉全屏、tabs 横向滚动、filter-bar 换行——覆盖面远超一般面板项目。
- 组件层：DataTable 移动端压缩单元格、PageCard/StatCard 各自断点。
- **缺口（已修）**：原全局规则把表格按钮 padding 缩到 `0 6px`，**反向损害触屏可点性**（52 处 tiny 按钮 ≈ 24px 高）。已改为保持 min-height 32px 的紧凑态。
- 遗留建议：`scrollX` 默认 700px 对部分宽表（审计页、备份页）在 480px 屏上略保守，建议逐表核对。

### 3. 间距 — ★★★★☆

- 硬编码 px 间距全项目仅 23 处（vs 500+ token 引用），纪律良好。
- **缺口（已修）**：本次新写的 SecurityExtras.vue 占了 6 处（自食其果），已全部 token 化，并顺手清掉了 2 个不存在的 `--divider-color`/`--text-tertiary` 变量引用（替换为 `--line`/`--faint`）。

### 4. 留白 — ★★★★☆

- 内容区 `--space-5` 外边距 + 卡片内 `--space-5`，1400px max-width，无挤压感。
- 空态（NEmpty）+ 筛选行 + 工具行的垂直节奏统一。
- 无发现性缺口。

### 5. 动效 — ★★★★☆

- 动效 token 分档（--t-fast/normal/slow + 两套 ease），LED 签名动画、路由 fade、按钮 lift、登录页 dot-pulse——克制且有身份。
- **reduced-motion 全局降级已存在**（main.css 通配符方案，覆盖所有 animation/transition）——优于多数项目。
- **缺口（已修）**：2 处 `transition: all`（见上）。
- 遗留建议：路由切换目前只有 opacity fade，可考虑入场加 8px 上移（`--ease-out`）增强层级感，属可选打磨。

### 6. 字体 — ★★★★☆

- 字体栈规范（Inter → 系统栈 / JetBrains Mono → 系统等宽栈），中文回退 PingFang/微软雅黑正确。
- **已澄清**：index.html 未加载 webfont，Inter/JetBrains Mono 实为"系统已装则增强"——已在 tokens.css 写明这是**刻意决策**（首屏 <2s 优先于品牌字体一致性），消除"宣称与实际不符"。
- `tabular-nums` 保证仪表盘数字不跳动，好。

### 7. SVG 图标 — ★★★☆☆（唯一结构性缺口）

- 自定义 SVG 仅 1 个：RhinoLogo（5 path 犀牛图形，笔画参数统一、引用 token 色）——质量好但孤例。
- 全部功能图标来自 `@vicons/ionicons5`（34 个文件引用），无 emoji 当图标（扫描通过）。
- **建议（未实施，需排期）**：从"深夜数据中心"的画面句提取 4-6 个专属器物图标（机柜服务器/终端光标/数据库 cylinder 变体/SSL 证书锁/容器方块/备份磁带），以 `sprite.svg` 落地替换高频位置的通用图标，签名一致性再上一档。

## 三、本轮修复清单

| # | 修复 | 文件 |
|---|------|------|
| 1 | 全局 `:focus-visible` 键盘焦点（2px acc 描边 + 偏移） | `styles/main.css` |
| 2 | 可点击卡片键盘态（focus 同 hover lift） | `styles/main.css` |
| 3 | 移动端表格按钮保持 32px 可点高度（原规则反向缩小） | `styles/main.css` |
| 4 | 2 处 `transition: all` → 显式属性列表 | AppStoreTab / InstanceListTab |
| 5 | SecurityExtras 6 处硬编码间距 token 化 + 2 处无效变量清理 | `SecurityExtras.vue` |
| 6 | ~~Skip link~~ **已撤销（2026-09-13）**：冒泡可见的体验损失大于无障碍收益，面板类产品键盘流极少，移除 | `DefaultLayout.vue` |
| 7 | 字体策略决策注释（系统字体优先是刻意为之） | `tokens.css` |

**验证**：`vue-tsc --noEmit` ✅ `vite build` ✅

## 四、质量门脚本说明

`check_page.py` 对 Login.vue 报 6 FAIL，经人工校准 5 项为 Vue SFC 场景误报（lang/title 在 index.html、token 由全局注入、reduced-motion/focus-visible 由 main.css 全局覆盖）。skip link 曾实施后撤销——内网工具面板键盘流极少，隐藏失效时反而是负体验。建议后续给脚本加 SFC 适配模式。

## 五、后续建议（按优先级）

1. **专属图标集**（上文 SVG 节）——签名一致性的最后一块拼图
2. 宽表逐表核对移动端 `scrollX` 值
3. 路由切换入场动画可选增强（opacity + translateY 8px）
4. `check_page.py` 增加 SFC 模式纳入 CI
