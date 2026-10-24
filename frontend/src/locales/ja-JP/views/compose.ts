// compose ページ文言
export default {
  cardTitle: 'Composeスタック',
  buildCardTitle: 'イメージビルド',
  refresh: '更新',
  createStack: 'スタックを作成',
  columns: {
    name: 'スタック名',
    containers: 'コンテナ',
    status: 'ステータス',
    actions: '操作'
  },
  status: {
    running: '実行中',
    stopped: '未実行',
    partial: '一部実行中'
  },
  actions: {
    start: '起動',
    stop: '停止',
    restart: '再起動',
    edit: '編集',
    delete: '削除'
  },
  empty: {
    noStack: 'Composeスタックがありません',
    noStackHint: '「スタックを作成」をクリックして docker-compose.yml の内容を貼り付けると管理できます',
    noBuildLog: 'ビルドログがありません',
    noBuildLogHint: 'タグと Dockerfile を入力し「ビルド開始」をクリックすると、ログがリアルタイムに出力されます'
  },
  build: {
    tagLabel: 'イメージタグ',
    start: 'ビルド開始',
    abort: '中止'
  },
  editor: {
    createTitle: 'スタックを作成',
    editTitle: 'スタックを編集 · {name}',
    nameLabel: 'スタック名',
    namePlaceholder: '例：blog',
    cancel: 'キャンセル',
    save: '保存'
  },
  deleteDialog: {
    title: 'スタックを削除',
    content: 'スタック「{name}」を削除しますか？',
    confirm: '削除'
  },
  rules: {
    nameRequired: 'スタック名を入力してください',
    tagRequired: 'イメージタグを入力してください',
    dockerfileRequired: 'Dockerfileの内容を入力してください'
  },
  messages: {
    actionDone: '操作が完了しました',
    actionFailed: '操作に失敗しました',
    loadContentFailed: 'スタック内容の読み込みに失敗しました',
    saved: '保存しました',
    saveFailed: '保存に失敗しました',
    deleted: '削除しました',
    deleteFailed: '削除に失敗しました',
    buildFailed: 'ビルドに失敗しました',
    buildAborted: '[ABORT] ビルドを中止しました',
    buildError: 'ビルドでエラーが発生しました',
    buildDone: 'ビルドが完了しました'
  }
}
