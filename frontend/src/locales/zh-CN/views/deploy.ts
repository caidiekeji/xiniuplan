// deploy 页面文案
export default {
  pageTitle: 'Git 部署',
  pageSubtitle: '仓库 Webhook 触发拉取构建发布，也可在面板手动部署',
  refresh: '刷新',
  addApp: '添加应用',
  buildTypes: {
    static: '静态站点',
    node: 'Node 构建',
    dockerfile: 'Dockerfile'
  },
  status: {
    success: '成功',
    failed: '失败',
    deploying: '部署中'
  },
  columns: {
    name: '应用名称',
    repo: '仓库地址',
    branch: '分支',
    buildType: '构建类型',
    destPath: '目标目录',
    lastDeploy: '上次部署',
    actions: '操作'
  },
  notDeployed: '未部署',
  actions: {
    deploy: '部署',
    logs: '日志',
    edit: '编辑',
    delete: '删除'
  },
  messages: {
    deployDone: '应用「{name}」部署完成',
    deployFailed: '部署失败',
    updated: '应用已更新',
    created: '应用已创建',
    saveFailed: '保存失败',
    deleted: '应用已删除',
    deleteFailed: '删除失败'
  },
  rules: {
    nameRequired: '请输入应用名称',
    repoRequired: '请输入仓库地址',
    destPathRequired: '请输入部署目标目录'
  },
  empty: {
    title: '暂无部署应用',
    desc: '添加 Git 仓库后，可手动部署或配置 Webhook 自动发布'
  },
  webhook: {
    tipPrefix: 'Webhook 地址格式：',
    tipSuffix: '，签名与应用 Secret 一致'
  },
  form: {
    addTitle: '添加应用',
    editTitle: '编辑应用',
    nameLabel: '应用名称',
    namePlaceholder: '例如：blog',
    branchLabel: '分支',
    repoLabel: '仓库地址',
    repoPlaceholder: "https://github.com/user/repo.git（可含凭据 user:token{'@'}）",
    buildTypeLabel: '构建类型',
    containerPortLabel: '容器端口',
    outDirLabel: '产物目录',
    destPathLabel: '部署目标目录',
    destPathPlaceholder: '/var/www/blog（static/node 类型发布到此目录）',
    hint: "Webhook 触发地址：/api/v1/deploy/webhook/{'{id}'}?sign=Secret，需在仓库 Webhook 设置中配置"
  },
  logs: {
    title: '部署日志 · {name}',
    empty: '暂无部署记录',
    expand: '展开',
    collapse: '收起'
  },
  deleteDialog: {
    title: '删除应用',
    content: '确定要删除部署应用「{name}」吗？部署产物目录不会被删除。',
    confirm: '删除'
  }
}
