// compose 页面文案
export default {
  cardTitle: 'Compose 栈',
  buildCardTitle: '镜像构建',
  refresh: '刷新',
  createStack: '新建栈',
  columns: {
    name: '栈名',
    containers: '容器',
    status: '状态',
    actions: '操作'
  },
  status: {
    running: '运行中',
    stopped: '未运行',
    partial: '部分运行'
  },
  actions: {
    start: '启动',
    stop: '停止',
    restart: '重启',
    edit: '编辑',
    delete: '删除'
  },
  empty: {
    noStack: '暂无 Compose 栈',
    noStackHint: '点击「新建栈」粘贴 docker-compose.yml 内容即可纳管',
    noBuildLog: '暂无构建日志',
    noBuildLogHint: '填写 tag 与 Dockerfile 后点击「开始构建」，日志将实时输出'
  },
  build: {
    tagLabel: '镜像 tag',
    start: '开始构建',
    abort: '中止'
  },
  editor: {
    createTitle: '新建栈',
    editTitle: '编辑栈 · {name}',
    nameLabel: '栈名',
    namePlaceholder: '例如：blog',
    cancel: '取消',
    save: '保存'
  },
  deleteDialog: {
    title: '删除栈',
    content: '确定要删除栈「{name}」吗？',
    confirm: '删除'
  },
  rules: {
    nameRequired: '请输入栈名',
    tagRequired: '请输入镜像 tag',
    dockerfileRequired: '请输入 Dockerfile 内容'
  },
  messages: {
    actionDone: '操作完成',
    actionFailed: '操作失败',
    loadContentFailed: '加载栈内容失败',
    saved: '已保存',
    saveFailed: '保存失败',
    deleted: '已删除',
    deleteFailed: '删除失败',
    buildFailed: '构建失败',
    buildAborted: '[ABORT] 构建已中止',
    buildError: '构建出现错误',
    buildDone: '构建完成'
  }
}
