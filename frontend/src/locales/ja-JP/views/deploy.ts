// deploy ページ文言
export default {
  pageTitle: 'Git デプロイ',
  pageSubtitle: 'リポジトリの Webhook でプル・ビルド・公開をトリガー、パネルからの手動デプロイも可能',
  refresh: '更新',
  addApp: 'アプリを追加',
  buildTypes: {
    static: '静的サイト',
    node: 'Node ビルド',
    dockerfile: 'Dockerfile'
  },
  status: {
    success: '成功',
    failed: '失敗',
    deploying: 'デプロイ中'
  },
  columns: {
    name: 'アプリ名',
    repo: 'リポジトリ URL',
    branch: 'ブランチ',
    buildType: 'ビルド種別',
    destPath: '対象ディレクトリ',
    lastDeploy: '前回のデプロイ',
    actions: '操作'
  },
  notDeployed: '未デプロイ',
  actions: {
    deploy: 'デプロイ',
    logs: 'ログ',
    edit: '編集',
    delete: '削除'
  },
  messages: {
    deployDone: 'アプリ「{name}」のデプロイが完了しました',
    deployFailed: 'デプロイに失敗しました',
    updated: 'アプリを更新しました',
    created: 'アプリを作成しました',
    saveFailed: '保存に失敗しました',
    deleted: 'アプリを削除しました',
    deleteFailed: '削除に失敗しました'
  },
  rules: {
    nameRequired: 'アプリ名を入力してください',
    repoRequired: 'リポジトリ URL を入力してください',
    destPathRequired: 'デプロイ先ディレクトリを入力してください'
  },
  empty: {
    title: 'デプロイアプリはまだありません',
    desc: 'Git リポジトリを追加すると手動デプロイ、または Webhook を設定して自動公開できます'
  },
  webhook: {
    tipPrefix: 'Webhook URL 形式：',
    tipSuffix: '（署名はアプリの Secret と一致します）'
  },
  form: {
    addTitle: 'アプリを追加',
    editTitle: 'アプリを編集',
    nameLabel: 'アプリ名',
    namePlaceholder: '例：blog',
    branchLabel: 'ブランチ',
    repoLabel: 'リポジトリ URL',
    repoPlaceholder: "https://github.com/user/repo.git（認証情報 user:token{'@'} を含められます）",
    buildTypeLabel: 'ビルド種別',
    containerPortLabel: 'コンテナポート',
    outDirLabel: '成果物ディレクトリ',
    destPathLabel: 'デプロイ先ディレクトリ',
    destPathPlaceholder: '/var/www/blog（static/node 種別はこのディレクトリに公開されます）',
    hint: "Webhook トリガー URL：/api/v1/deploy/webhook/{'{id}'}?sign=Secret（リポジトリの Webhook 設定で設定してください）"
  },
  logs: {
    title: 'デプロイログ · {name}',
    empty: 'デプロイ履歴はまだありません',
    expand: '展開',
    collapse: '折りたたむ'
  },
  deleteDialog: {
    title: 'アプリを削除',
    content: 'デプロイアプリ「{name}」を削除しますか？ビルド成果物のディレクトリは削除されません。この操作は元に戻せません。',
    confirm: '削除'
  }
}
