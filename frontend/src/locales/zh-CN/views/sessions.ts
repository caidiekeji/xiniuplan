// sessions 页面文案
export default {
  apiKeyTab: 'API Key',
  title: '在线会话',
  refresh: '刷新',
  createKey: '创建 API Key',
  cancel: '取消',
  create: '创建',
  columns: {
    user: '用户',
    type: '类型',
    loginAt: '登录时间',
    lastActive: '最近活跃',
    actions: '操作',
    name: '名称',
    status: '状态',
    lastUsed: '最近使用',
    expiresAt: '过期时间'
  },
  actions: {
    revoke: '踢出',
    delete: '删除'
  },
  status: {
    active: '启用',
    inactive: '禁用',
    neverUsed: '从未',
    neverExpires: '永不'
  },
  form: {
    name: '名称',
    namePlaceholder: '如：CI 部署脚本',
    expiresInDays: '有效期(天)',
    neverExpiresHint: '0 = 永不过期'
  },
  secret: {
    title: '请立即保存 API Key',
    warning: 'Secret 仅显示一次，关闭后无法再次查看！调用方式：请求头',
    saved: '我已保存'
  },
  deleteDialog: {
    title: '删除 API Key',
    content: '确定删除「{name}」？使用该 Key 的程序将立即失效。'
  },
  messages: {
    loadSessionsFailed: '加载会话失败',
    revoked: '会话已踢出',
    revokeFailed: '踢出失败',
    loadKeysFailed: '加载 API Key 失败',
    nameRequired: '请输入名称',
    createFailed: '创建失败',
    deleted: '已删除'
  }
}
