// appstore 页面文案
export default {
  pageTitle: '应用商店',
  pageSubtitle: '一键部署常用 Web 应用 · 容器化运行 · 环境参数可视化配置',
  tabs: {
    store: '应用商店',
    installed: '已安装',
    sources: '源管理'
  },
  addSource: '添加源',
  loadSourcesFailed: '加载源列表失败',
  sourceTypes: {
    official: '官方',
    community: '社区',
    private: '私有'
  },
  form: {
    addTitle: '添加源',
    editTitle: '编辑源',
    nameLabel: '源名称',
    namePlaceholder: '请输入源名称，如：官方应用源',
    typeLabel: '类型',
    typePlaceholder: '请选择源类型',
    pathLabel: '本地路径',
    enabledLabel: '启用'
  },
  validate: {
    nameRequired: '请输入源名称',
    communityUrlRequired: '请输入社区源 URL',
    privatePathRequired: '请输入私有源本地路径',
    officialOnlyOnce: '官方源只能存在一个，请先修改或删除现有官方源'
  },
  messages: {
    sourceUpdated: '源已更新',
    sourceAdded: '源已添加',
    actionFailed: '操作失败',
    sourceDisabled: '已禁用「{name}」',
    sourceEnabled: '已启用「{name}」',
    syncDone: '同步完成：新增 {added} / 更新 {updated} / 删除 {deleted}',
    syncFailed: '同步失败',
    officialNoDelete: '官方源不可删除',
    sourceDeleted: '「{name}」已删除',
    deleteFailed: '删除失败'
  },
  columns: {
    name: '源名称',
    type: '类型',
    url: 'URL / 路径',
    status: '状态',
    lastSync: '最后同步',
    actions: '操作'
  },
  status: {
    enabled: '启用',
    disabled: '禁用'
  },
  actions: {
    sync: '同步',
    enable: '启用',
    disable: '禁用',
    edit: '编辑',
    delete: '删除'
  },
  deleteDialog: {
    title: '删除源',
    content: '确定要删除源「{name}」吗？该源下的应用将不再可用，此操作不可恢复。',
    confirm: '删除'
  },
  categories: {
    website: '建站',
    devtools: '开发工具',
    database: '数据库',
    middleware: '中间件',
    monitor: '监控',
    office: '办公',
    storage: '存储',
    network: '网络'
  },
  statuses: {
    installed: '已安装',
    installing: '安装中',
    failed: '安装失败',
    not_installed: '未安装'
  },
  storeTab: {
    allCategories: '全部',
    allSources: '全部来源',
    officialGroup: '官方源',
    communityGroup: '社区源',
    privateGroup: '私有源',
    loadFailed: '加载应用列表失败',
    sourcePlaceholder: '来源',
    searchPlaceholder: '搜索应用...',
    memory: '内存 {value}',
    disk: '磁盘 {value}',
    install: '安装',
    noMatch: '没有匹配的应用',
    noApps: '暂无可用应用'
  },
  installedTab: {
    loadFailed: '加载已安装应用失败',
    noAccessUrl: '该应用暂无访问地址',
    noContainer: '该应用没有关联的容器，无法查看日志',
    noLogs: '（暂无日志）',
    loadLogsFailed: '加载日志失败：{error}',
    upgraded: '「{name}」已升级到 {version}',
    upgradeFailed: '升级失败',
    uninstalled: '{name} 已卸载',
    uninstallFailed: '卸载失败',
    refresh: '刷新',
    empty: '暂无已安装应用，去应用商店看看吧',
    browseApps: '浏览应用',
    logTitle: '应用日志 - {name}',
    uninstallTitle: '卸载应用',
    uninstallContent: '确定要卸载「{name}」吗？容器及相关数据将被删除，此操作不可恢复。',
    uninstallConfirm: '卸载',
    columns: {
      appName: '应用名',
      version: '版本',
      category: '分类',
      installedAt: '安装时间'
    },
    actions: {
      open: '打开',
      logs: '日志',
      upgrade: '升级',
      uninstall: '卸载'
    }
  },
  installDialog: {
    title: '安装 {name}',
    titleNoApp: '安装应用',
    confirmTitle: '确认信息',
    confirmDesc: '查看应用详情',
    configTitle: '配置参数',
    configDesc: '设置环境变量',
    installingTitle: '安装中',
    installingDesc: '等待部署完成',
    version: '版本 {version}',
    minResources: '最低资源要求',
    memory: '内存: {value}',
    disk: '磁盘: {value}',
    cpu: 'CPU: {value}',
    exposedPorts: '暴露端口:',
    requiredWarning: '请填写必填项: {label}',
    inputPlaceholder: '请输入{label}',
    selectPlaceholder: '请选择{label}',
    noEnv: '该应用无需配置环境参数，可直接安装',
    logsTitle: '安装日志',
    logPreparing: '正在准备安装环境...',
    logPulling: '正在拉取镜像 {name}:{version}...',
    logConfiguring: '正在配置容器和网络...',
    logDone: '安装完成！',
    logError: '错误: {error}',
    unknownError: '未知错误',
    successTitle: '安装成功',
    successMessage: '{name} 安装成功',
    installFailed: '安装失败',
    accessUrl: '访问地址:',
    initialPassword: '初始密码:',
    containerId: '容器ID:',
    copy: '复制',
    passwordCopied: '密码已复制到剪贴板',
    copyFailed: '复制失败，请手动复制',
    done: '完成',
    cancel: '取消',
    next: '下一步',
    prev: '上一步',
    start: '开始安装',
    viewInstalled: '查看已安装'
  }
}
