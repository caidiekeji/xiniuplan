// roleList ページ文言
export default {
  title: 'ロールと権限',
  subtitle: 'システムロールとその権限割り当てを管理します',
  createRole: 'ロールを作成',
  editRole: 'ロールを編集',
  columns: {
    name: 'ロール名',
    code: 'ロールコード',
    description: '説明',
    permissions: '権限数',
    type: '種類',
    actions: '操作'
  },
  status: {
    builtIn: '組み込み',
    custom: 'カスタム'
  },
  actions: {
    edit: '編集',
    delete: '削除'
  },
  labels: {
    name: 'ロール名',
    code: 'ロールコード',
    description: '説明',
    permissions: '権限の割り当て'
  },
  hints: {
    name: '例：運用エンジニア',
    code: '例：ops_engineer',
    description: '任意の説明',
    nameRequired: 'ロール名を入力してください',
    codeRequired: 'ロールコードを入力してください'
  },
  matrix: {
    selectAll: 'すべて選択',
    selected: '選択済み {selected} / {total}'
  },
  permissions: {
    modules: {
      site: 'サイト',
      database: 'データベース',
      file: 'ファイル',
      terminal: 'ターミナル',
      container: 'コンテナ',
      appstore: 'アプリストア',
      firewall: 'ファイアウォール',
      backup: 'バックアップ',
      monitor: '監視',
      user: 'ユーザー',
      audit: '監査',
      setting: '設定',
      alert: 'アラート',
      cron: 'スケジュールタスク'
    },
    actions: {
      view: '閲覧',
      create: '作成',
      update: '変更',
      delete: '削除'
    }
  },
  messages: {
    loadError: 'ロール一覧の読み込みに失敗しました',
    updated: 'ロールを更新しました',
    created: 'ロールを作成しました',
    saveFailed: '保存に失敗しました',
    deleted: 'ロールを削除しました',
    deleteFailed: '削除に失敗しました'
  },
  delete: {
    title: 'ロールの削除',
    content: 'ロール「{name}」を削除しますか？この操作は元に戻せません。'
  }
}
