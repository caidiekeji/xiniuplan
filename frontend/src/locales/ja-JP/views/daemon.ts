// daemon ページ文言
export default {
  pageTitle: 'プロセスガード',
  pageSubtitle: '管理プロセスのクラッシュを自動で再起動、パネル再起動後も設定どおりに復元',
  refresh: '更新',
  addDaemon: 'ガードを追加',
  columns: {
    name: '名前',
    command: 'コマンド',
    workDir: '作業ディレクトリ',
    status: '状態',
    restarts: '再起動回数',
    lastExitCode: '前回の終了コード',
    autoStart: '自動起動',
    actions: '操作'
  },
  status: {
    running: '実行中',
    stopped: '停止中'
  },
  autoStart: {
    on: 'オン',
    off: 'オフ'
  },
  actions: {
    start: '起動',
    stop: '停止',
    logs: 'ログ',
    edit: '編集',
    delete: '削除'
  },
  messages: {
    started: 'プロセスを起動しました',
    startFailed: '起動に失敗しました',
    stopped: 'プロセスを停止しました',
    stopFailed: '停止に失敗しました',
    updated: 'ガード設定を更新しました',
    created: 'ガードプロセスを作成しました',
    saveFailed: '保存に失敗しました',
    deleted: 'ガードプロセスを削除しました',
    deleteFailed: '削除に失敗しました'
  },
  rules: {
    nameRequired: 'プロセス名を入力してください',
    commandRequired: '起動コマンドを入力してください'
  },
  empty: {
    title: 'ガードプロセスはまだありません',
    desc: '追加すると異常終了時に自動再起動され、「自動起動」を有効にするとパネル起動時に復元されます'
  },
  form: {
    addTitle: 'ガードプロセスを追加',
    editTitle: 'ガードプロセスを編集',
    nameLabel: 'プロセス名',
    namePlaceholder: '例：node-api',
    commandLabel: '起動コマンド',
    commandPlaceholder: "例：node /opt/app/server.js または bash -c 'cd /opt/app && ./start.sh'",
    commandHint: '任意の shell コマンドに対応、複雑な起動ロジックは bash -c で囲めます',
    workDirLabel: '作業ディレクトリ',
    workDirPlaceholder: '例：/opt/app（任意）',
    autoStartLabel: 'クラッシュ時の自動再起動 / パネル起動時の自動復元',
    enabledLabel: '有効'
  },
  logs: {
    title: 'プロセスログ · {name}',
    loading: '読み込み中…',
    empty: 'ログ出力はまだありません'
  },
  deleteDialog: {
    title: 'ガードプロセスを削除',
    content: 'ガードプロセス「{name}」を削除しますか？実行中のプロセスは終了されます。',
    confirm: '削除'
  }
}
