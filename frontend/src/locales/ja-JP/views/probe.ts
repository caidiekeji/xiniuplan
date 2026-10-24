// probe ページ文言
export default {
  pageTitle: 'Web 監視',
  pageSubtitle: 'HTTP(S)/TCP の可用性を監視し、障害時は自動でアラート',
  refresh: '更新',
  addProbe: '監視プローブを追加',
  status: {
    up: '正常',
    down: '障害',
    unknown: '未チェック'
  },
  switchStatus: {
    enabled: '有効',
    disabled: '無効'
  },
  columns: {
    name: '名前',
    type: '種類',
    target: '監視対象',
    interval: '間隔',
    status: '状態',
    latency: '遅延',
    httpCode: 'HTTP コード',
    lastChecked: '最終チェック',
    enabled: '有効',
    actions: '操作',
    checkedAt: 'チェック日時',
    result: '結果',
    message: '説明'
  },
  actions: {
    checkNow: '今すぐチェック',
    edit: '編集',
    delete: '削除'
  },
  messages: {
    checkDone: 'チェックが完了しました',
    checkFailed: 'チェックに失敗しました',
    updated: 'プローブを更新しました',
    created: 'プローブを作成しました',
    saveFailed: '保存に失敗しました',
    deleted: 'プローブを削除しました',
    deleteFailed: '削除に失敗しました'
  },
  rules: {
    nameRequired: '監視プローブ名を入力してください',
    targetRequired: 'チェック対象を入力してください'
  },
  empty: {
    title: '監視プローブはまだありません',
    desc: '追加すると対象の可用性を継続的にチェックし、障害時は「アラート通知」チャネルへ送信します'
  },
  historyButton: '先頭のプローブのチェック履歴を表示',
  form: {
    addTitle: '監視プローブを追加',
    editTitle: '監視プローブを編集',
    nameLabel: 'プローブ名',
    namePlaceholder: '例：公式サイトのトップページ',
    typeLabel: '種類',
    intervalLabel: 'チェック間隔（秒）',
    targetLabel: 'チェック対象',
    keywordLabel: '応答キーワード（任意、レスポンスに含める必要があります）',
    keywordPlaceholder: '例：ok',
    enabledLabel: '有効'
  },
  results: {
    title: 'チェック履歴 · {name}'
  },
  deleteDialog: {
    title: '監視プローブを削除',
    content: '監視プローブ「{name}」を削除しますか？チェック履歴も一緒に削除されます。',
    confirm: '削除'
  }
}
