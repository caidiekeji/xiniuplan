// website ページ文言
export default {
  title: 'サイト管理',
  subtitle: 'Web サイトのライフサイクル全体を管理',
  createSite: 'サイトを作成',
  search: {
    domainLabel: 'ドメイン',
    domainPlaceholder: 'ドメインで検索',
    typeLabel: 'タイプ',
    typePlaceholder: 'すべてのタイプ',
    statusLabel: 'ステータス',
    statusPlaceholder: 'すべてのステータス'
  },
  types: {
    static: '静的サイト',
    runtime: '実行環境',
    reverseProxy: 'リバースプロキシ'
  },
  status: {
    running: '稼働中',
    stopped: '停止中',
    error: '異常'
  },
  columns: {
    domain: 'ドメイン',
    type: 'タイプ',
    runtime: '実行環境',
    status: 'ステータス',
    createdAt: '作成日時',
    actions: '操作'
  },
  ssl: {
    enabled: '有効',
    disabled: '無効'
  },
  actions: {
    start: '起動',
    stop: '停止',
    restart: '再起動',
    config: '設定',
    logs: 'ログ',
    edit: '編集',
    delete: '削除'
  },
  messages: {
    loadListFailed: 'サイト一覧の取得に失敗しました',
    startSuccess: 'サイトを起動しました',
    startFailed: '起動に失敗しました',
    stopSuccess: 'サイトを停止しました',
    stopFailed: '停止に失敗しました',
    restartSuccess: 'サイトを再起動しました',
    restartFailed: '再起動に失敗しました',
    createSuccess: 'サイトを作成しました',
    createFailed: '作成に失敗しました',
    updateSuccess: 'サイトを更新しました',
    updateFailed: '更新に失敗しました',
    readConfigFailed: 'Nginx 設定の読み込みに失敗しました',
    configSaved: 'Nginx 設定を保存しました',
    saveFailed: '保存に失敗しました',
    logsFailed: 'サイトログの取得に失敗しました',
    deleteSuccess: 'サイトを削除しました',
    deleteFailed: '削除に失敗しました'
  },
  rules: {
    domainRequired: 'メインドメインを入力してください',
    rootDirRequired: 'サイトのルートディレクトリを入力してください',
    runtimeRequired: '実行環境を選択してください',
    proxyTargetRequired: 'リバースプロキシの送信先を入力してください'
  },
  createModal: {
    title: 'サイトを作成',
    typeLabel: 'サイトタイプ',
    reverseProxyDesc: 'バックエンドサービスへ転送',
    domainLabel: 'メインドメイン',
    domainPlaceholder: '例: example.com',
    aliasesLabel: 'バインドドメイン',
    aliasesPlaceholder: 'サブドメインを入力して Enter で追加',
    runtimeLabel: '実行環境',
    runtimePlaceholder: '実行環境を選択',
    proxyTargetLabel: 'リバースプロキシ送信先',
    proxyTargetPlaceholder: '例: http://127.0.0.1:3000',
    rootDirLabel: 'サイトのルートディレクトリ',
    remarkLabel: '備考',
    remarkPlaceholder: '任意の備考'
  },
  editModal: {
    title: 'サイトを編集',
    domainLocked: 'メインドメインは作成後に変更できません。変更する場合は削除して再作成してください'
  },
  configDrawer: {
    title: 'Nginx 設定 — {domain}',
    loading: '読み込み中...',
    cancel: 'キャンセル',
    save: '設定を保存'
  },
  logsModal: {
    title: 'サイトログ — {domain}',
    refresh: '更新'
  },
  logTypes: {
    access: 'アクセスログ (access)',
    error: 'エラーログ (error)'
  },
  deleteDialog: {
    title: 'サイトを削除',
    content: 'サイト「{domain}」を削除してもよろしいですか？この操作は元に戻せません。',
    deleteFiles: 'サイトのルートディレクトリ内のファイルも削除する'
  }
}
