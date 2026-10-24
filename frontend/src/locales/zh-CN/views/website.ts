// website 页面文案
export default {
  title: '网站管理',
  subtitle: 'Web 站点全生命周期管理',
  createSite: '创建网站',
  search: {
    domainLabel: '域名',
    domainPlaceholder: '输入域名搜索',
    typeLabel: '类型',
    typePlaceholder: '全部类型',
    statusLabel: '状态',
    statusPlaceholder: '全部状态'
  },
  types: {
    static: '静态网站',
    runtime: '运行环境',
    reverseProxy: '反向代理'
  },
  status: {
    running: '运行中',
    stopped: '已停止',
    error: '异常'
  },
  columns: {
    domain: '域名',
    type: '类型',
    runtime: '运行环境',
    status: '状态',
    createdAt: '创建时间',
    actions: '操作'
  },
  ssl: {
    enabled: '已启用',
    disabled: '未启用'
  },
  actions: {
    start: '启动',
    stop: '停止',
    restart: '重启',
    config: '配置',
    logs: '日志',
    edit: '编辑',
    delete: '删除'
  },
  messages: {
    loadListFailed: '加载网站列表失败',
    startSuccess: '站点已启动',
    startFailed: '启动失败',
    stopSuccess: '站点已停止',
    stopFailed: '停止失败',
    restartSuccess: '站点已重启',
    restartFailed: '重启失败',
    createSuccess: '网站创建成功',
    createFailed: '创建失败',
    updateSuccess: '网站已更新',
    updateFailed: '更新失败',
    readConfigFailed: '读取 Nginx 配置失败',
    configSaved: 'Nginx 配置已保存',
    saveFailed: '保存失败',
    logsFailed: '获取站点日志失败',
    deleteSuccess: '网站已删除',
    deleteFailed: '删除失败'
  },
  rules: {
    domainRequired: '请输入主域名',
    rootDirRequired: '请输入网站根目录',
    runtimeRequired: '请选择运行环境',
    proxyTargetRequired: '请输入反向代理目标'
  },
  createModal: {
    title: '创建网站',
    typeLabel: '网站类型',
    reverseProxyDesc: '转发到后端服务',
    domainLabel: '主域名',
    domainPlaceholder: '例如 example.com',
    aliasesLabel: '绑定域名',
    aliasesPlaceholder: '输入副域名后回车添加',
    runtimeLabel: '运行环境',
    runtimePlaceholder: '选择运行环境',
    proxyTargetLabel: '反向代理目标',
    proxyTargetPlaceholder: '例如 http://127.0.0.1:3000',
    rootDirLabel: '网站根目录',
    remarkLabel: '备注',
    remarkPlaceholder: '可选备注信息'
  },
  editModal: {
    title: '编辑网站',
    domainLocked: '主域名创建后不可修改，如需更换请删除后重建'
  },
  configDrawer: {
    title: 'Nginx 配置 — {domain}',
    loading: '加载中...',
    cancel: '取消',
    save: '保存配置'
  },
  logsModal: {
    title: '站点日志 — {domain}',
    refresh: '刷新'
  },
  logTypes: {
    access: '访问日志 (access)',
    error: '错误日志 (error)'
  },
  deleteDialog: {
    title: '删除网站',
    content: '确定要删除网站「{domain}」吗？此操作不可撤销。',
    deleteFiles: '同时删除网站根目录文件'
  }
}
