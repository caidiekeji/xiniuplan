// dockerResources ページ文言
export default {
  title: 'ネットワークとボリューム',
  subtitle: 'Dockerネットワークとボリュームの管理（C-008 / C-009）',
  refresh: '更新',
  tabs: {
    networks: 'ネットワーク',
    volumes: 'ボリューム'
  },
  createNetwork: 'ネットワークを作成',
  createVolume: 'ボリュームを作成',
  dialogs: {
    createNetworkTitle: 'Dockerネットワークを作成',
    createVolumeTitle: 'Dockerボリュームを作成'
  },
  form: {
    nameLabel: '名前',
    namePlaceholderNetwork: '例: app-net',
    namePlaceholderVolume: '例: app-data',
    driverLabel: 'ドライバー',
    subnetLabel: 'サブネット',
    subnetPlaceholder: '任意。例: 172.30.0.0/16',
    gatewayLabel: 'ゲートウェイ',
    gatewayPlaceholder: '任意',
    cancel: 'キャンセル',
    confirm: '作成'
  },
  columns: {
    name: '名前',
    driver: 'ドライバー',
    subnet: 'サブネット',
    containers: 'コンテナ',
    inUse: '使用中',
    usedBy: '使用元',
    actions: '操作'
  },
  yes: 'はい',
  no: 'いいえ',
  delete: '削除',
  deleteDialog: {
    cancel: 'キャンセル'
  },
  deleteNetwork: {
    title: 'ネットワークを削除',
    content: 'ネットワーク「{name}」を削除しますか？'
  },
  deleteVolume: {
    title: 'ボリュームを削除',
    content: 'ボリューム「{name}」を削除しますか？データは復元できません。'
  },
  messages: {
    loadNetworksFailed: 'ネットワーク一覧の読み込みに失敗しました',
    loadVolumesFailed: 'ボリューム一覧の読み込みに失敗しました',
    networkCreated: 'ネットワークを作成しました',
    volumeCreated: 'ボリュームを作成しました',
    createFailed: '作成に失敗しました',
    deleted: '削除しました',
    volumeInUse: 'ボリュームはコンテナで使用中のため削除できません'
  }
}
