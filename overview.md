# 前端设计体系整改报告（2026-10-20，全量完成）

整改范围：`frontend/src` 全部 view 组件 + 共享组件 + tokens.css + 三语言 locale。**全部改动已部署至 192.168.31.154 并浏览器中/英/日三语实证。**

## 一、i18n 全量整改（本轮主体）

### 结果
- **33 个 locale 模块 × 3 语言**全部填充完毕（原 29 个 + 新注册 file/cron/backup/container 4 个）。
- `scripts/check-i18n.py`：引用 key **2216 个，全部在 zh-CN 存在**，退出码 0（脚本 import 正则原本漏掉 `./views/` 路径，已修复——此前 1275 个 key 实际未被核对）。
- `vue-tsc --noEmit` 全绿 —— MessageSchema=typeof zhCN 保证三语言 key 集合零差异。
- 用户可见中文残留终扫：**仅剩 2 处约定保留项**（settings 品牌名"阿里云 DNS/腾讯云 DNSPod/简体中文"9 行、DefaultLayout '中文' fallback 1 行）。注释/console/技术标识按契约不抽。

### 执行方式
- 主控预生成骨架 + 注册 index（避免并发冲突）→ 17 组并行 agent 分批抽取 → 三层验证（check-i18n / vue-tsc / 残留终扫）。
- 补齐了首轮遗漏：4 个空骨架模块（ai/compose/dockerResources/nodes）、database 5 个 Tab（RedisPanel 等）、appstore 3 个子组件、file/cron/backup/container 大页、全局残留（main.ts/request.ts/router/ErrorBoundary/Login 忘记密码）。
- 关键机制：模块级含 t() 的数组/映射包 `computed(() => [...])`；局部 `t` 变量机械改名 `tk`/`tp`；动态拼接改 `{name}`/`{n}` 插值；shared.ts 纯模块用「传入 t 参数」方案（不 import 组合式 API）。

### 顺带发现并修复的真 bug
- **NConfigProvider 硬编码 `zhCN`**：naive-ui 内置文案（分页"条/页"、日期选择器）切英文后仍中文 → App.vue 改 computed 按 locale 切 enUS/jaJP + dateLocale。

## 二、此前审查修复项（已在前轮上线）

1. **`--warn` 语义缺陷**（青绿→琥珀）、**7 个幽灵 token** 全部改项目 token。
2. **monitor/dashboard 图表**：CPU=acc 青绿、内存=info 蓝双色区分；time 真实时间轴 + 断档插 null 断线。
3. **monitor 假数据兜底删除**（编造的 CPU 23% → 真实空态）。
4. **nodes 硬编码阈值色** → var(--err)/var(--warn)/var(--ok)。
5. **formatTime/formatBytes 重复 8 处** → utils/format.ts 单一出口。

## 三、组件与工程化

- **PageCard 新增 `loading` prop**：为 true 时 body 渲染骨架条（NSkeleton ×5），纯增量不破坏现有 29 个使用方；默认 false 走原 slot。
- **SearchForm 决策**：保留封装（已接 i18n、有移动端适配），定位为新页面搜索栏默认选择，不强制推广（避免无谓重构引入回归）。
- **deploy.py 日志落盘加固**：log()/run() 双写 `scripts/deploy.log` —— 已知"stdout 重定向偶发静默 exit 1"问题发生时可追溯现场（本轮已复现一次，重跑即过，日志文件确认落盘生效）。

## 四、验证记录

- `python scripts/check-i18n.py` → 2216 key 全绿。
- `npx vue-tsc --noEmit` → 0 错误（多轮）。
- `vite build` + `go build`（30MB 交叉编译）+ `scripts/deploy.py` → 健康检查 ok、1990 HTTP 200。
- **浏览器实证**（agent-browser）：中文态 dashboard 图表 canvas 正常；切 English 后 9 页巡检（file/cron/backup/container/database/appstore/ai/nodes/monitor/settings）UI 文案 cn≈0，分页后缀 "20 / 页"→英文；切日本語 后菜单/页面全假名汉字正常；切回中文回归正常。
- 英文态残留仅剩**后端数据**（非前端文案，属后端 i18n 范畴，未在本轮范围）：cron 种子任务名/描述（`backend/internal/service/cron/tasks_extra.go`）、安全评分项名称（`system/security_score.go`）、应用目录描述（appstore 目录数据）。

## 五、遗留（可选后续）

1. 后端数据 i18n（种子任务/安全评分/应用目录描述）——需要后端返回 key + 前端映射或后端多语言，单独立项。
2. 骨架屏：PageCard loading 已提供基建，各页面按需接线（目前多数页仍 NSpin 转圈）。
3. 部署静默 exit 1 根因未查明（已加日志落盘，下次复现可定位）。
