// backup ページ文案
export default {
  title: 'バックアップと復元',
  subtitle: 'サイト / データベース / システムデータのバックアップとワンクリック復元',
  createBackup: 'バックアップを作成',
  remoteStorage: 'リモートストレージ',
  types: {
    site: 'サイト',
    database: 'データベース',
    system: 'システム',
    app: 'アプリ'
  },
  status: {
    pending: '待機中',
    completed: '完了',
    running: '実行中',
    failed: '失敗'
  },
  unknown: '不明',
  search: {
    type: '種類',
    allTypes: 'すべての種類',
    keyword: '検索',
    keywordPlaceholder: '対象・備考で検索'
  },
  columns: {
    type: 'バックアップ種類',
    target: '対象',
    size: 'ファイルサイズ',
    status: 'ステータス',
    createdAt: '作成日時',
    actions: '操作'
  },
  actions: {
    restore: '復元',
    download: 'ダウンロード',
    push: 'プッシュ',
    delete: '削除'
  },
  create: {
    title: 'バックアップを作成',
    typeLabel: 'バックアップ種類',
    siteTarget: 'サイトを選択',
    sitePlaceholder: 'バックアップするサイトを選択してください',
    dbTarget: 'データベースを選択',
    dbPlaceholder: 'バックアップするデータベースインスタンスを選択してください',
    systemLabel: 'システムバックアップ',
    systemHint: '設定・アプリケーションデータ・システムファイルを含むシステム全体のフルバックアップを行います',
    remark: '備考',
    remarkPlaceholder: '任意の備考'
  },
  rules: {
    targetRequired: 'バックアップ対象を選択または入力してください'
  },
  messages: {
    loadFailed: 'バックアップ一覧の取得に失敗しました',
    enterTarget: 'バックアップ対象を入力してください',
    createSuccess: 'バックアップタスクを作成しました。実行中です...',
    createFailed: 'バックアップの作成に失敗しました',
    restoreSuccess: '復元を開始しました。復元前にスナップショットを自動作成しました',
    restoreFailed: '復元に失敗しました',
    downloadStart: 'ダウンロードを開始しました',
    downloadFailed: 'ダウンロードに失敗しました',
    deleteSuccess: 'バックアップを削除しました',
    deleteFailed: '削除に失敗しました',
    loadRemoteFailed: 'リモートストレージの取得に失敗しました',
    selectRemote: 'リモートストレージを選択してください',
    pushSuccess: 'プッシュが完了しました',
    pushFailed: 'プッシュに失敗しました',
    enterName: '名称を入力してください',
    remoteCreateSuccess: 'リモートストレージを作成しました',
    remoteCreateFailed: '作成に失敗しました',
    testSuccess: '接続は正常です',
    testFailed: '接続に失敗しました'
  },
  restore: {
    title: 'バックアップの復元',
    content: 'バックアップ「{name}」を復元しますか？復元では現在のデータが上書きされ、復元前にスナップショットが自動作成されます。',
    confirm: '復元する'
  },
  remove: {
    title: 'バックアップの削除',
    content: 'バックアップ「{name}」を削除しますか？この操作は取り消せません。',
    confirm: '削除'
  },
  push: {
    title: 'バックアップをリモートストレージへプッシュ',
    placeholder: 'リモートストレージを選択',
    cancel: 'キャンセル',
    confirm: 'プッシュ'
  },
  remote: {
    title: 'リモートバックアップストレージ（S3 / WebDAV / SFTP / パネル）',
    enabled: '有効',
    disabled: '無効',
    test: 'テスト',
    delete: '削除',
    close: '閉じる',
    create: '新規作成'
  },
  remoteCreate: {
    title: 'リモートストレージを作成',
    name: '名称',
    type: '種類',
    types: {
      s3: 'S3 互換ストレージ',
      webdav: 'WebDAV',
      sftp: 'SFTP',
      panel: '別のパネル'
    },
    address: 'アドレス / ホスト',
    username: 'ユーザー名',
    password: 'パスワード',
    cancel: 'キャンセル',
    create: '作成'
  }
}
