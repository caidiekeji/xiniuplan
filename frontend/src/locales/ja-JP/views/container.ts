// container ページ文言
export default {
  title: 'コンテナ管理',
  subtitle: 'Dockerコンテナのライフサイクル管理 · リアルタイムログ · イメージ閲覧',
  refresh: '更新',
  searchPlaceholder: 'コンテナ名 / イメージ / ID を検索',
  showImagesLabel: 'イメージを表示',
  imagesTitle: 'イメージ一覧',
  unknownError: '不明なエラー',
  execTitle: 'コンテナターミナル - {name}',
  execDisconnected: '[接続が切断されました]',
  status: {
    running: '実行中',
    stopped: '停止済み',
    exited: '終了済み',
    created: '作成済み',
    paused: '一時停止中',
    restarting: '再起動中'
  },
  filter: {
    all: 'すべて'
  },
  columns: {
    name: '名前',
    image: 'イメージ',
    status: 'ステータス',
    ports: 'ポートマッピング',
    createdAt: '作成日時',
    actions: '操作'
  },
  imageColumns: {
    image: 'イメージ',
    id: 'ID',
    size: 'サイズ',
    createdAt: '作成日時'
  },
  actions: {
    start: '起動',
    stop: '停止',
    restart: '再起動',
    logs: 'ログ',
    terminal: 'ターミナル',
    stats: 'リソース',
    delete: '削除',
    more: 'その他'
  },
  empty: {
    noMatch: '一致するコンテナがありません',
    noContainer: 'コンテナがありません',
    noImages: 'イメージがありません'
  },
  messages: {
    dockerUnavailable: 'Dockerサービスを利用できません。Dockerが起動しているか確認してください',
    started: 'コンテナ {name} を起動しました',
    startFailed: '起動に失敗しました',
    stopped: 'コンテナ {name} を停止しました',
    stopFailed: '停止に失敗しました',
    restarted: 'コンテナ {name} を再起動しました',
    restartFailed: '再起動に失敗しました',
    deleted: 'コンテナ {name} を削除しました',
    deleteFailed: '削除に失敗しました',
    statsFailed: 'コンテナのリソース使用状況を取得できませんでした',
    copySuccess: 'ログをクリップボードにコピーしました',
    copyFailed: 'コピーに失敗しました。テキストを選択して手動でコピーしてください',
    loadImagesFailed: 'イメージ一覧の読み込みに失敗しました',
    imageDeleted: 'イメージ {name} を削除しました',
    imageDeleteFailed: 'イメージの削除に失敗しました'
  },
  logs: {
    title: 'コンテナログ - {name}',
    titlePlain: 'コンテナログ',
    tail: '末尾の行数',
    tailLines: '{n} 行',
    follow: 'リアルタイム追従',
    copy: 'コピー',
    empty: '(ログ出力なし)',
    loadFailed: 'ログの取得に失敗しました: {msg}'
  },
  deleteDialog: {
    title: 'コンテナを削除',
    content: 'コンテナ「{name}」を削除しますか？この操作は元に戻せず、コンテナ内の永続化されていないデータは失われます。',
    confirm: '削除'
  },
  imageDeleteDialog: {
    title: 'イメージを削除',
    content: 'イメージ「{name}」を削除しますか？この操作は元に戻せず、使用中のコンテナがある場合は削除に失敗します。',
    confirm: '削除'
  },
  stats: {
    title: 'リソース使用状況 - {name}',
    cpuUsage: 'CPU使用率',
    memoryUsage: 'メモリ使用量',
    memoryValue: '{usage} / {limit}（{percent}%）'
  }
}
