# UI 文案词表

> 来源：design-audit-2026-10-21.md §5.3。**新文案先查此表**；表内禁用变体不得出现在界面/代码注释以外的用户可见文本中。
> grep 门（终验脚本）：`grep -rn "确定\|空空如也" frontend/src --include="*.vue" --include="*.ts"` 应为 0 命中（i18n 词条本身除外）。

## 词表

| 类别 | 统一词 | 禁用变体 | 说明 |
|---|---|---|---|
| 主操作 | 保存（配置类）/ 创建（新建资源）/ 安装（应用）/ 确认（通用） | 确定、OK、提交 | ModalForm confirmText 按场景传词表词 |
| 危险操作 | 删除 / 停止 / 卸载 + 确认弹窗必须含后果句 | 移除、清空（除日志场景） | 日志场景允许「清空」 |
| 错误文案 | 「{对象}{动作}失败：{原因}」+ 可重试时附「请重试」 | 裸错误码、Error 500 | 对应 DataTable error 态 + 重试按钮 |
| 空态 | 「暂无{对象}」+ 首次使用附引导动作链接 | 空空如也、无数据 | DataTable 内置「暂无数据」兜底 |
| 称呼 | 零人称（现状：界面几乎不用人称代词，维持） | 你/您混用 | 不新增「您」开头文案 |

## 英文/日文对照（i18n 落地参考）

| 中文统一词 | en-US | ja-JP |
|---|---|---|
| 保存 | Save | 保存 |
| 创建 | Create | 作成 |
| 安装 | Install | インストール |
| 确认 | Confirm | 確認 |
| 删除 | Delete | 削除 |
| 停止 | Stop | 停止 |
| 卸载 | Uninstall | アンインストール |
| 暂无{对象} | No {object} | {object}がありません |
| 失败 | Failed | 失敗 |
| 请重试 | Please retry | 再試行してください |

## 用法示例

```ts
// ✅ 正确：新建备份弹窗主按钮用「创建」
<ModalForm :confirm-text="t('common.create')" …>

// ❌ 错误：主按钮写死「确定」
<ModalForm confirm-text="确定" …>  // 禁用变体 + 绕过 i18n

// ✅ 正确：删除确认含后果句
<ConfirmDialog
  :title="t('common.confirmDelete')"
  :content="t('backup.deleteWarning')"  // 「删除后备份文件不可恢复」
  type="error"
/>
```
