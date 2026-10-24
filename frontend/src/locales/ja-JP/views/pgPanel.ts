// pgPanel ページ文案
export default {
  tabs: {
    schema: 'テーブル管理',
    users: 'ユーザー権限',
    sql: 'SQL 実行'
  },
  refresh: '更新',
  execute: '実行',
  delete: '削除',
  createFailed: '作成に失敗しました',
  deleteFailed: '削除に失敗しました',
  selectDb: 'データベースを選択してください',
  selectDbLabel: 'データベースを選択',
  selectDbPlaceholder: 'データベースを選択',
  loadDatabasesFailed: 'データベース一覧の読み込みに失敗しました',
  loadTablesFailed: 'テーブル一覧の読み込みに失敗しました',
  loadSchemaFailed: 'テーブル構造の読み込みに失敗しました',
  loadTableDataFailed: 'テーブルデータの読み込みに失敗しました',
  loadUsersFailed: 'ユーザー一覧の読み込みに失敗しました',
  columns: {
    tableName: 'テーブル名',
    rowsEstimate: '行数（概算）',
    size: 'サイズ',
    actions: '操作',
    structure: '構造',
    data: 'データ',
    columnName: '列名',
    type: '型',
    nullable: 'NULL 許容',
    defaultValue: 'デフォルト値',
    primaryKey: '主キー',
    username: 'ユーザー名',
    attributes: '属性',
    validUntil: '有効期限',
    permanent: '無期限',
    grant: '権限付与',
    revoke: '権限取り消し'
  },
  sidebarTitle: 'データベース',
  noDatabases: 'データベースがありません',
  selectDbLeft: '左のデータベースを選択してください',
  noTables: 'このデータベースにテーブルはありません',
  tableSchema: 'テーブル構造: {name}',
  tableData: 'テーブルデータ: {name}',
  indexes: 'インデックス',
  noUsers: 'ユーザーがありません。上のボタンから作成してください',
  allowDestructive: '破壊的操作を許可',
  selectDbFirst: '先にデータベースを選択してください',
  sql: {
    enterSql: 'SQL 文を入力してください',
    executed: '実行が完了しました',
    failed: '実行に失敗しました',
    success: '実行に成功しました',
    duration: '実行時間 {ms} ms',
    affected: '{rows} 行に影響',
    noResultSet: '結果セットなし（SELECT 文以外）',
    defaultSql: '-- SQL 文をここに入力\nSELECT 1;'
  },
  createDb: {
    title: 'データベース作成',
    nameLabel: 'データベース名',
    namePlaceholder: '例: myapp_db',
    encodingLabel: 'エンコーディング',
    collateLabel: '照合順序',
    ownerPlaceholder: '空の場合は現在のユーザーを使用',
    nameRequired: 'データベース名を入力してください',
    success: 'データベースを作成しました'
  },
  dbDeleted: 'データベース {name} を削除しました',
  createUser: {
    title: 'データベースユーザー作成',
    create: 'ユーザー作成',
    nameLabel: 'ユーザー名',
    namePlaceholder: '例: app_user',
    passwordLabel: 'パスワード',
    passwordPlaceholder: 'パスワードを入力してください',
    roleAttributesLabel: 'ロール属性',
    nameRequired: 'ユーザー名を入力してください',
    passwordRequired: 'パスワードを入力してください',
    success: 'ユーザーを作成しました'
  },
  userDeleted: 'ユーザー {name} を削除しました',
  grant: {
    title: 'ユーザーに付与: {name}',
    privilegesLabel: '権限',
    success: '権限を付与しました',
    failed: '権限の付与に失敗しました'
  },
  revoke: {
    title: '権限を取り消し: {name}',
    dbPlaceholder: '権限を取り消すデータベースを選択',
    success: '{user} の {db} に対する権限を取り消しました',
    failed: '権限の取り消しに失敗しました'
  },
  deleteDb: {
    title: 'データベースを削除',
    content: 'データベース「{name}」を削除しますか？この操作は元に戻せません。'
  },
  deleteUser: {
    title: 'ユーザーを削除',
    content: 'ユーザー「{name}」を削除しますか？'
  }
}
