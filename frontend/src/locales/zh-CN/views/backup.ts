// backup 页面文案
export default {
  title: '备份恢复',
  subtitle: '网站 / 数据库 / 系统数据备份与一键恢复',
  createBackup: '创建备份',
  remoteStorage: '远程存储',
  types: {
    site: '网站',
    database: '数据库',
    system: '系统',
    app: '应用'
  },
  status: {
    pending: '等待中',
    completed: '已完成',
    running: '进行中',
    failed: '失败'
  },
  unknown: '未知',
  search: {
    type: '类型',
    allTypes: '全部类型',
    keyword: '搜索',
    keywordPlaceholder: '输入目标/备注搜索'
  },
  columns: {
    type: '备份类型',
    target: '目标',
    size: '文件大小',
    status: '状态',
    createdAt: '创建时间',
    actions: '操作'
  },
  actions: {
    restore: '恢复',
    download: '下载',
    push: '推送',
    delete: '删除'
  },
  create: {
    title: '创建备份',
    typeLabel: '备份类型',
    siteTarget: '选择站点',
    sitePlaceholder: '选择要备份的站点',
    dbTarget: '选择数据库',
    dbPlaceholder: '选择要备份的数据库实例',
    systemLabel: '系统备份',
    systemHint: '将对整个系统进行全量备份，包括配置、应用数据与系统文件',
    remark: '备注',
    remarkPlaceholder: '可选备注信息'
  },
  rules: {
    targetRequired: '请选择或输入备份目标'
  },
  messages: {
    loadFailed: '加载备份列表失败',
    enterTarget: '请输入备份目标',
    createSuccess: '备份任务已创建，正在进行中...',
    createFailed: '创建备份失败',
    restoreSuccess: '恢复任务已启动，恢复前已自动创建快照',
    restoreFailed: '恢复失败',
    downloadStart: '开始下载',
    downloadFailed: '下载失败',
    deleteSuccess: '备份已删除',
    deleteFailed: '删除失败',
    loadRemoteFailed: '加载远程存储失败',
    selectRemote: '请选择远程存储',
    pushSuccess: '推送完成',
    pushFailed: '推送失败',
    enterName: '请输入名称',
    remoteCreateSuccess: '远程存储已创建',
    remoteCreateFailed: '创建失败',
    testSuccess: '连接正常',
    testFailed: '连接失败'
  },
  restore: {
    title: '恢复备份',
    content: '确定要恢复备份「{name}」吗？恢复将覆盖当前数据，系统会在恢复前自动创建快照。',
    confirm: '确认恢复'
  },
  remove: {
    title: '删除备份',
    content: '确定要删除备份「{name}」吗？此操作不可撤销。',
    confirm: '删除'
  },
  push: {
    title: '推送备份到远程存储',
    placeholder: '选择远程存储',
    cancel: '取消',
    confirm: '推送'
  },
  remote: {
    title: '远程备份存储（S3 / WebDAV / SFTP / 面板）',
    enabled: '启用',
    disabled: '停用',
    test: '测试',
    delete: '删除',
    close: '关闭',
    create: '新建'
  },
  remoteCreate: {
    title: '新建远程存储',
    name: '名称',
    type: '类型',
    types: {
      s3: 'S3 兼容存储',
      webdav: 'WebDAV',
      sftp: 'SFTP',
      panel: '另一台面板'
    },
    address: '地址/主机',
    username: '用户名',
    password: '密码',
    cancel: '取消',
    create: '创建'
  }
}
