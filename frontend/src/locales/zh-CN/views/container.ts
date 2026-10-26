// container 页面文案
export default {
  title: '容器管理',
  subtitle: 'Docker 容器生命周期管理 · 实时日志 · 镜像浏览',
  refresh: '刷新',
  searchPlaceholder: '搜索容器名 / 镜像 / ID',
  showImagesLabel: '显示镜像',
  imagesTitle: '镜像列表',
  pull: {
    button: '拉取镜像',
    title: '拉取镜像',
    imageLabel: '镜像名',
    imagePlaceholder: '例如 nginx:1.25',
    imageHint: '支持官方镜像、加速源自动改写、私有仓库完整地址',
    registryLabel: '拉取源',
    registryPlaceholder: '选择拉取源',
    usernamePlaceholder: '私有仓库用户名',
    passwordPlaceholder: '私有仓库密码',
    start: '开始拉取',
    cancel: '取消拉取',
    close: '关闭',
    pulling: '正在拉取…',
    complete: '镜像拉取完成',
    cancelled: '已取消拉取',
    imageRequired: '请输入镜像名',
    moreLayers: '还有 {n} 个层…'
  },
  unknownError: '未知错误',
  execTitle: '容器终端 - {name}',
  execDisconnected: '[连接已断开]',
  status: {
    running: '运行中',
    stopped: '已停止',
    exited: '已退出',
    created: '已创建',
    paused: '已暂停',
    restarting: '重启中'
  },
  filter: {
    all: '全部'
  },
  columns: {
    name: '名称',
    image: '镜像',
    status: '状态',
    ports: '端口映射',
    createdAt: '创建时间',
    actions: '操作'
  },
  imageColumns: {
    image: '镜像',
    id: 'ID',
    size: '大小',
    createdAt: '创建时间'
  },
  actions: {
    start: '启动',
    stop: '停止',
    restart: '重启',
    logs: '日志',
    terminal: '终端',
    stats: '资源',
    delete: '删除',
    more: '更多'
  },
  empty: {
    noMatch: '没有匹配的容器',
    noContainer: '暂无容器',
    noImages: '暂无镜像'
  },
  messages: {
    dockerUnavailable: 'Docker 服务不可用，请检查 Docker 是否正常运行',
    started: '容器 {name} 已启动',
    startFailed: '启动失败',
    stopped: '容器 {name} 已停止',
    stopFailed: '停止失败',
    restarted: '容器 {name} 已重启',
    restartFailed: '重启失败',
    deleted: '容器 {name} 已删除',
    deleteFailed: '删除失败',
    statsFailed: '获取容器资源统计失败',
    copySuccess: '日志已复制到剪贴板',
    copyFailed: '复制失败，请手动选择复制',
    loadImagesFailed: '加载镜像列表失败',
    imageDeleted: '镜像 {name} 已删除',
    imageDeleteFailed: '删除镜像失败'
  },
  logs: {
    title: '容器日志 - {name}',
    titlePlain: '容器日志',
    tail: '尾部行数',
    tailLines: '{n} 行',
    follow: '实时跟随',
    copy: '复制',
    empty: '(无日志输出)',
    loadFailed: '获取日志失败: {msg}'
  },
  deleteDialog: {
    title: '删除容器',
    content: '确定要删除容器「{name}」吗？此操作不可恢复，容器内未持久化的数据将丢失。',
    confirm: '删除'
  },
  imageDeleteDialog: {
    title: '删除镜像',
    content: '确定要删除镜像「{name}」吗？此操作不可恢复，若有容器正在使用该镜像将删除失败。',
    confirm: '删除'
  },
  stats: {
    title: '资源统计 - {name}',
    cpuUsage: 'CPU 使用率',
    memoryUsage: '内存使用',
    memoryValue: '{usage} / {limit}（{percent}%）'
  }
}
