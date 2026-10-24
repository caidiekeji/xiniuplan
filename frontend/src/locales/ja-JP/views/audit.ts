// audit ページ文言
export default {
  resultOptions: {
    all: 'すべて',
    success: '成功',
    failed: '失敗'
  },
  actionOptions: {
    all: 'すべての操作',
    login: 'ログイン',
    createSite: 'サイト作成',
    deleteSite: 'サイト削除',
    createDatabase: 'データベース作成',
    installApp: 'アプリインストール',
    restartService: 'サービス再起動',
    fileOperation: 'ファイル操作',
    systemSettings: 'システム設定',
    firewallRule: 'ファイアウォールルール'
  },
  columns: {
    time: '日時',
    operator: '操作者',
    actionType: '操作タイプ',
    resource: 'リソース',
    result: '結果',
    username: 'ユーザー名',
    subject: '対象',
    count: '回数',
    lastSeen: '最終発生',
    detail: '詳細',
    type: 'タイプ'
  },
  messages: {
    loadAuditFailed: '操作ログの読み込みに失敗しました',
    loadLoginFailed: 'ログインログの読み込みに失敗しました',
    exportSuccess: 'ログのエクスポートに成功しました',
    exportMock: 'エクスポートタスクを作成しました（mock）',
    detectFailed: '検出に失敗しました'
  },
  anomalyTypes: {
    bruteForceIp: '総当たり攻撃',
    bruteForceUser: 'リスト攻撃',
    newIpLogin: '新規デバイスログイン'
  },
  tabs: {
    operations: '操作ログ',
    anomalies: '異常検出',
    logins: 'ログインログ'
  },
  audit: {
    title: '操作ログ',
    subtitle: 'すべてのパネル操作の完全な監査記録',
    userId: 'ユーザーID',
    userIdPlaceholder: 'ユーザーID',
    actionType: '操作タイプ',
    result: '結果',
    timeRange: '期間'
  },
  anomalies: {
    title: '異常ログイン検出',
    subtitle: '直近 24 時間：総当たり / リスト攻撃 / 新規デバイスログイン（S-012）',
    refresh: '再検出'
  },
  logins: {
    title: 'ログインログ',
    subtitle: 'パネルログイン記録と異常ログイン追跡',
    exportCsv: 'CSV エクスポート'
  }
}
