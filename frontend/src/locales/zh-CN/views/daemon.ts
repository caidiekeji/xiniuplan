// daemon 页面文案
export default {
  pageTitle: '进程守护',
  pageSubtitle: '托管进程崩溃自动拉起，面板重启后按配置恢复',
  refresh: '刷新',
  addDaemon: '添加守护',
  columns: {
    name: '名称',
    command: '命令',
    workDir: '工作目录',
    status: '状态',
    restarts: '重启次数',
    lastExitCode: '上次退出码',
    autoStart: '自动启动',
    actions: '操作'
  },
  status: {
    running: '运行中',
    stopped: '已停止'
  },
  autoStart: {
    on: '开启',
    off: '关闭'
  },
  actions: {
    start: '启动',
    stop: '停止',
    logs: '日志',
    edit: '编辑',
    delete: '删除'
  },
  messages: {
    started: '进程已启动',
    startFailed: '启动失败',
    stopped: '进程已停止',
    stopFailed: '停止失败',
    updated: '守护配置已更新',
    created: '守护进程已创建',
    saveFailed: '保存失败',
    deleted: '守护进程已删除',
    deleteFailed: '删除失败'
  },
  rules: {
    nameRequired: '请输入进程名称',
    commandRequired: '请输入启动命令'
  },
  empty: {
    title: '暂无守护进程',
    desc: '添加后进程异常退出将被自动重启，开启「自动启动」可随面板启动恢复'
  },
  form: {
    addTitle: '添加守护进程',
    editTitle: '编辑守护进程',
    nameLabel: '进程名称',
    namePlaceholder: '例如：node-api',
    commandLabel: '启动命令',
    commandPlaceholder: "例如：node /opt/app/server.js 或 bash -c 'cd /opt/app && ./start.sh'",
    commandHint: '支持任意 shell 命令，复杂启动逻辑可通过 bash -c 包裹',
    workDirLabel: '工作目录',
    workDirPlaceholder: '例如：/opt/app（可选）',
    autoStartLabel: '崩溃自动拉起 / 面板重启自动启动',
    enabledLabel: '启用'
  },
  logs: {
    title: '进程日志 · {name}',
    loading: '加载中…',
    empty: '暂无日志输出'
  },
  deleteDialog: {
    title: '删除守护进程',
    content: '确定要删除守护进程「{name}」吗？正在运行的进程将被终止。',
    confirm: '删除'
  }
}
