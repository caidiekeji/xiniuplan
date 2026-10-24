# base 组件 API 手册

面板基础组件层唯一文档（审计 D08）。新增/修改组件 props 时**必须**同步本文件。
所有颜色/间距/字号一律走 `var(--*)` token，禁止在组件内硬编码视觉值。

统一约定：

- 事件命名 `update:xxx`（v-model 兼容）+ 语义事件（`search`/`retry`/`select`）
- 文案走 vue-i18n，按钮默认值引用 `common.*`，可被 props 覆盖
- 弹窗三原子归口：**ModalForm**（表单）/ **ConfirmDialog**（确认）/ **ContentModal**（内容展示），views 层禁止直用 NModal

---

## PageCard

页面/区块的卡片容器，标题 + 副标题 + 右上操作位 + 骨架屏。

| Prop | 类型 | 默认值 | 说明 |
|---|---|---|---|
| title | `string` | — | 卡片标题（h2） |
| subtitle | `string` | — | 副标题（fs-caption + text-quaternary） |
| padding | `string` | — | body 区内边距覆盖 |
| loading | `boolean` | `false` | body 渲染骨架屏 |

插槽：`extra`（头部右侧操作区）、default（body）

```vue
<PageCard :title="t('backup.title')" :loading="loading">
  <template #extra><NButton size="small">{{ t('common.create') }}</NButton></template>
  <DataTable :columns="columns" :data="list" />
</PageCard>
```

## DataTable

全站唯一表格出口，封装 NDataTable + 空态/加载态/错误态 + 移动端压缩（768/480 两档，**样式只在此定义**，勿在 main.css 重复）。

| Prop | 类型 | 默认值 | 说明 |
|---|---|---|---|
| columns | `DataTableColumns<any>` | 必填 | 列定义 |
| data | `any[]` | 必填 | 数据（null 兜底为空数组） |
| loading | `boolean` | `false` | NSpin 加载遮罩 |
| error | `string \| null` | `null` | 非空渲染错误态（文案 + 重试按钮），优先于空态 |
| pagination | `DataTableProps['pagination'] \| false` | `{pageSize:20, pageSizes:[10,20,50,100]}` | `false` 关闭 |
| bordered | `boolean` | `false` | |
| striped | `boolean` | `true` | |
| size | `'small' \| 'medium' \| 'large'` | `'medium'` | |
| rowKey | `(row) => string \| number` | `row.id \|\| row.key` | |
| checkedRowKeys | `(string \| number)[]` | — | 多选（v-model） |
| virtualScroll | `boolean` | `false` | |
| maxHeight | `number \| string` | — | |
| scrollX | `number \| string` | `700` | 横向滚动宽度 |

Emits：`update:checkedRowKeys`、`rowDblclick`、`rowContextmenu`、`retry`（错误态点重试）

插槽：`empty`（空/错误态内容，默认 NEmpty）、`retry`（重试按钮自定义）

```vue
<DataTable :columns="cols" :data="rows" :loading="loading" :error="errorMsg" @retry="load" />
```

## SearchForm

列表页筛选条容器，收集表单项后统一抛出。

| Prop | 类型 | 默认值 |
|---|---|---|
| loading | `boolean` | `false`（搜索按钮 loading） |

Emits：`search(values: Record<string, any>)`（FormData 收集）、`reset`

```vue
<SearchForm @search="onSearch" @reset="onReset">
  <NInput name="kw" :placeholder="t('common.search')" />
</SearchForm>
```

## ModalForm

弹窗三原子之一——**表单**场景（创建/编辑）。footer 固定 取消+确认，确认带 loading。

| Prop | 类型 | 默认值 | 说明 |
|---|---|---|---|
| show | `boolean` | 必填 | v-model:show |
| title | `string` | `''` | |
| loading | `boolean` | `false` | 确认按钮 loading |
| width | `number` | `520` | px |
| confirmText | `string` | `common.confirm` | 新建资源传「创建」（词表统一词） |
| cancelText | `string` | `common.cancel` | |

Emits：`update:show`、`submit`（确认；内部自动带 loading 门控）

```vue
<ModalForm v-model:show="show" :title="t('common.create')" :loading="saving" @submit="save">
  <NForm>…</NForm>
</ModalForm>
```

## ContentModal

弹窗三原子之一——**内容展示**场景（日志/终端/详情/导入等），无内置按钮。

| Prop | 类型 | 默认值 | 说明 |
|---|---|---|---|
| show | `boolean` | 必填 | v-model:show |
| title | `string` | — | |
| size | `'sm' \| 'md' \| 'lg' \| 'xl'` | `'md'` | 映射 `--modal-width-*`（480/560/800/960，均含 92/94vw 移动钳制） |
| maskClosable | `boolean` | `true` | 表单进行中关遮罩误触 |
| escClosable | `boolean` | `true` | |

Emits：`update:show`

插槽：default（内容）、footer（可选，透传 NModal footer）

```vue
<ContentModal v-model:show="show" :title="t('container.execTitle')" size="xl">
  <div ref="termEl" />
</ContentModal>
```

## ConfirmDialog

弹窗三原子之一——**确认**场景（删除/停止等危险操作），content 必须含后果句（词表）。

| Prop | 类型 | 默认值 | 说明 |
|---|---|---|---|
| show | `boolean` | 必填 | v-model:show |
| title | `string` | `''` | |
| content | `string` | — | 后果描述 |
| confirmText | `string` | `''` | 空 = common.confirm |
| cancelText | `string` | `''` | 空 = common.cancel |
| type | `'warning' \| 'error' \| 'info'` | `'warning'` | |
| loading | `boolean` | `false` | |

Emits：`update:show`、`confirm`、`cancel`

## StatusTag

LED 状态语言的状态徽章。status 驱动颜色 + 文案（i18n），text 可覆盖。

| Prop | 类型 | 默认值 | 说明 |
|---|---|---|---|
| status | `'running' \| 'stopped' \| 'error' \| 'warning' \| 'pending'` | 必填 | |
| text | `string` | 按 status 取 i18n | 覆盖显示文案 |

```vue
<StatusTag :status="row.status" />
```

## StatCard

指标卡（概览行）。

| Prop | 类型 | 默认值 | 说明 |
|---|---|---|---|
| title | `string` | 必填 | |
| value | `string \| number` | 必填 | |
| unit | `string` | — | 单位后缀 |
| subtitle | `string` | — | 副标题（text-quaternary） |
| trend | `number` | — | 趋势百分比 |
| trendLabel | `string` | — | 趋势说明 |
| loading | `boolean` | `false` | |
| accent | `'acc' \| 'acc2' \| 'ok' \| 'warn' \| 'err' \| 'info'` | `'acc2'` | 强调色 token |

## CodeEditor

Monaco 封装（monaco chunk 异步加载，见 index.ts）。

| Prop | 类型 | 默认值 |
|---|---|---|
| modelValue | `string` | `''` |
| language | `string` | `'javascript'` |
| theme | `string` | `'vs-dark'` |
| readOnly | `boolean` | `false` |
| height | `string` | `var(--editor-height)` |
| minimap | `boolean` | `true` |

Emits：`update:modelValue`

字号走 `--fs-body`（fsPx 换算）。

## CommandPalette

Cmd/Ctrl+K 命令面板（审计 D13）。菜单项由布局层传入（与侧栏分组同源）。

| Prop | 类型 | 默认值 | 说明 |
|---|---|---|---|
| items | `CommandItem[]` | 必填 | `{ key, label, icon?, group? }` |
| show | `boolean` | 受控可选 | 不传则内部自管（Ctrl+K） |

Emits：`update:show`、`select(key)`（路由 key 或 `action:*` 前缀动作）

键盘：`↑/↓` 移动、`Enter` 执行、`Esc` 关闭。

## ErrorBoundary

渲染错误隔离（ErrorBoundary 包裹路由出口）。捕获子树渲染错误 → 全页错误态 + 重试。
无必填 props；错误详情 console 输出，UI 只展示 message。

## PanelIcons.ts / RhinoLogo.vue

- `PanelIcons`：面板专属 SVG 图标集（Rack/Globe/Db/Cube/Terminal/Gauge/Tape/Firewall/Cert/Cron），inline SVG、currentColor 驱动，禁止 emoji 图标。
- `RhinoLogo`：品牌 logo 组件，`size` prop（默认 34）。

---

## 变更记录

| 日期 | 变更 |
|---|---|
| 2026-10-21 | 设计审计整改：新增 ContentModal / CommandPalette；DataTable 增 `error`/`retry`；ModalForm 增 `confirmText`/`cancelText`；表格移动端样式收口 DataTable |
