// dockerResources 页面文案
export default {
  title: '网络与存储卷',
  subtitle: 'Docker 网络与存储卷管理（C-008 / C-009）',
  refresh: '刷新',
  tabs: {
    networks: '网络',
    volumes: '存储卷'
  },
  createNetwork: '创建网络',
  createVolume: '创建存储卷',
  dialogs: {
    createNetworkTitle: '创建 Docker 网络',
    createVolumeTitle: '创建 Docker 存储卷'
  },
  form: {
    nameLabel: '名称',
    namePlaceholderNetwork: '如: app-net',
    namePlaceholderVolume: '如: app-data',
    driverLabel: '驱动',
    subnetLabel: '子网',
    subnetPlaceholder: '可选，如 172.30.0.0/16',
    gatewayLabel: '网关',
    gatewayPlaceholder: '可选',
    cancel: '取消',
    confirm: '创建'
  },
  columns: {
    name: '名称',
    driver: '驱动',
    subnet: '子网',
    containers: '容器',
    inUse: '使用中',
    usedBy: '被谁使用',
    actions: '操作'
  },
  yes: '是',
  no: '否',
  delete: '删除',
  deleteDialog: {
    cancel: '取消'
  },
  deleteNetwork: {
    title: '删除网络',
    content: '确定删除网络「{name}」？'
  },
  deleteVolume: {
    title: '删除存储卷',
    content: '确定删除存储卷「{name}」？数据将不可恢复。'
  },
  messages: {
    loadNetworksFailed: '加载网络失败',
    loadVolumesFailed: '加载存储卷失败',
    networkCreated: '网络已创建',
    volumeCreated: '存储卷已创建',
    createFailed: '创建失败',
    deleted: '已删除',
    volumeInUse: '存储卷正在被容器使用，无法删除'
  }
}
