// appstore ページ文言
export default {
  pageTitle: 'アプリストア',
  pageSubtitle: '一般的な Web アプリをワンクリックでデプロイ · コンテナ実行 · 環境変数の可視化設定',
  tabs: {
    store: 'アプリストア',
    installed: 'インストール済み',
    sources: 'ソース管理'
  },
  addSource: 'ソースを追加',
  loadSourcesFailed: 'ソース一覧の読み込みに失敗しました',
  sourceTypes: {
    official: '公式',
    community: 'コミュニティ',
    private: 'プライベート'
  },
  form: {
    addTitle: 'ソースを追加',
    editTitle: 'ソースを編集',
    nameLabel: 'ソース名',
    namePlaceholder: 'ソース名を入力してください（例：公式アプリソース）',
    typeLabel: '種類',
    typePlaceholder: 'ソースの種類を選択してください',
    pathLabel: 'ローカルパス',
    enabledLabel: '有効'
  },
  validate: {
    nameRequired: 'ソース名を入力してください',
    communityUrlRequired: 'コミュニティソースの URL を入力してください',
    privatePathRequired: 'プライベートソースのローカルパスを入力してください',
    officialOnlyOnce: '公式ソースは 1 つのみです。既存の公式ソースを変更または削除してください'
  },
  messages: {
    sourceUpdated: 'ソースを更新しました',
    sourceAdded: 'ソースを追加しました',
    actionFailed: '操作に失敗しました',
    sourceDisabled: '「{name}」を無効にしました',
    sourceEnabled: '「{name}」を有効にしました',
    syncDone: '同期完了：追加 {added} / 更新 {updated} / 削除 {deleted}',
    syncFailed: '同期に失敗しました',
    officialNoDelete: '公式ソースは削除できません',
    sourceDeleted: '「{name}」を削除しました',
    deleteFailed: '削除に失敗しました'
  },
  columns: {
    name: 'ソース名',
    type: '種類',
    url: 'URL / パス',
    status: '状態',
    lastSync: '最終同期',
    actions: '操作'
  },
  status: {
    enabled: '有効',
    disabled: '無効'
  },
  actions: {
    sync: '同期',
    enable: '有効化',
    disable: '無効化',
    edit: '編集',
    delete: '削除'
  },
  deleteDialog: {
    title: 'ソースを削除',
    content: 'ソース「{name}」を削除しますか？このソースのアプリは利用できなくなります。この操作は元に戻せません。',
    confirm: '削除'
  },
  categories: {
    website: 'サイト構築',
    devtools: '開発ツール',
    database: 'データベース',
    middleware: 'ミドルウェア',
    monitor: '監視',
    office: 'オフィス',
    storage: 'ストレージ',
    network: 'ネットワーク'
  },
  statuses: {
    installed: 'インストール済み',
    installing: 'インストール中',
    failed: 'インストール失敗',
    not_installed: '未インストール'
  },
  storeTab: {
    allCategories: 'すべて',
    allSources: 'すべてのソース',
    officialGroup: '公式ソース',
    communityGroup: 'コミュニティソース',
    privateGroup: 'プライベートソース',
    loadFailed: 'アプリ一覧の読み込みに失敗しました',
    sourcePlaceholder: 'ソース',
    searchPlaceholder: 'アプリを検索...',
    memory: 'メモリ {value}',
    disk: 'ディスク {value}',
    install: 'インストール',
    noMatch: '一致するアプリがありません',
    noApps: '利用可能なアプリがありません'
  },
  installedTab: {
    loadFailed: 'インストール済みアプリの読み込みに失敗しました',
    noAccessUrl: 'このアプリにはまだアクセス URL がありません',
    noContainer: 'このアプリに関連するコンテナがないため、ログは表示できません',
    noLogs: '（ログはありません）',
    loadLogsFailed: 'ログの読み込みに失敗しました：{error}',
    upgraded: '「{name}」を {version} にアップグレードしました',
    upgradeFailed: 'アップグレードに失敗しました',
    uninstalled: '{name} をアンインストールしました',
    uninstallFailed: 'アンインストールに失敗しました',
    refresh: '更新',
    empty: 'インストール済みアプリはありません。アプリストアを見てみましょう',
    browseApps: 'アプリを見る',
    logTitle: 'アプリログ - {name}',
    uninstallTitle: 'アプリをアンインストール',
    uninstallContent: '「{name}」をアンインストールしますか？コンテナと関連データは削除され、この操作は元に戻せません。',
    uninstallConfirm: 'アンインストール',
    columns: {
      appName: 'アプリ名',
      version: 'バージョン',
      category: 'カテゴリ',
      installedAt: 'インストール日時'
    },
    actions: {
      open: '開く',
      logs: 'ログ',
      upgrade: 'アップグレード',
      uninstall: 'アンインストール'
    }
  },
  installDialog: {
    title: '{name} をインストール',
    titleNoApp: 'アプリをインストール',
    confirmTitle: '情報を確認',
    confirmDesc: 'アプリの詳細を確認',
    configTitle: 'パラメータ設定',
    configDesc: '環境変数を設定',
    installingTitle: 'インストール中',
    installingDesc: 'デプロイ完了を待機',
    version: 'バージョン {version}',
    minResources: '最低リソース要件',
    memory: 'メモリ: {value}',
    disk: 'ディスク: {value}',
    cpu: 'CPU: {value}',
    exposedPorts: '公開ポート:',
    requiredWarning: '必須項目を入力してください：{label}',
    inputPlaceholder: '{label}を入力してください',
    selectPlaceholder: '{label}を選択してください',
    noEnv: 'このアプリは環境変数の設定が不要です。そのままインストールできます',
    logsTitle: 'インストールログ',
    logPreparing: 'インストール環境を準備しています...',
    logPulling: 'イメージ {name}:{version} を取得しています...',
    logConfiguring: 'コンテナとネットワークを設定しています...',
    logDone: 'インストールが完了しました！',
    logError: 'エラー: {error}',
    unknownError: '不明なエラー',
    successTitle: 'インストール成功',
    successMessage: '{name} のインストールに成功しました',
    installFailed: 'インストールに失敗しました',
    accessUrl: 'アクセス URL:',
    initialPassword: '初期パスワード:',
    containerId: 'コンテナ ID:',
    copy: 'コピー',
    passwordCopied: 'パスワードをクリップボードにコピーしました',
    copyFailed: 'コピーに失敗しました。手動でコピーしてください',
    done: '完了',
    cancel: 'キャンセル',
    next: '次へ',
    prev: '戻る',
    start: 'インストール開始',
    viewInstalled: 'インストール済みを見る'
  }
}
