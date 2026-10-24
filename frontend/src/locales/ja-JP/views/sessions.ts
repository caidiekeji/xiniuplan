// sessions ページ文言
export default {
  apiKeyTab: 'API Key',
  title: 'オンラインセッション',
  refresh: '更新',
  createKey: 'API Key を作成',
  cancel: 'キャンセル',
  create: '作成',
  columns: {
    user: 'ユーザー',
    type: 'タイプ',
    loginAt: 'ログイン日時',
    lastActive: '最終アクティブ',
    actions: '操作',
    name: '名前',
    status: 'ステータス',
    lastUsed: '最終使用',
    expiresAt: '有効期限'
  },
  actions: {
    revoke: '強制終了',
    delete: '削除'
  },
  status: {
    active: '有効',
    inactive: '無効',
    neverUsed: '未使用',
    neverExpires: '無期限'
  },
  form: {
    name: '名前',
    namePlaceholder: '例: CI デプロイスクリプト',
    expiresInDays: '有効期間（日）',
    neverExpiresHint: '0 = 無期限'
  },
  secret: {
    title: 'API Key をすぐ保存してください',
    warning: 'Secret は一度しか表示されず、閉じると再表示できません！送信方法：リクエストヘッダー',
    saved: '保存しました'
  },
  deleteDialog: {
    title: 'API Key を削除',
    content: '「{name}」を削除してもよろしいですか？この Key を使用しているプログラムは直ちに無効になります。'
  },
  messages: {
    loadSessionsFailed: 'セッションの取得に失敗しました',
    revoked: 'セッションを強制終了しました',
    revokeFailed: '強制終了に失敗しました',
    loadKeysFailed: 'API Key の取得に失敗しました',
    nameRequired: '名前を入力してください',
    createFailed: '作成に失敗しました',
    deleted: '削除しました'
  }
}
