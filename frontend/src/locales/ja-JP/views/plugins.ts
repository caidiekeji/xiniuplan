// plugins ページ文言
export default {
  title: 'プラグイン管理',
  subtitle: '{total} 個インストール済み · {enabled} 個有効',
  refresh: '更新',
  status: {
    running: '稼働中',
    enabled: '有効',
    disabled: '無効'
  },
  toggle: {
    enable: 'クリックで有効化',
    disable: 'クリックで無効化'
  },
  actions: {
    config: '設定'
  },
  empty: 'インストール済みプラグインはありません',
  configModal: {
    title: 'プラグイン設定',
    titleWithPlugin: 'プラグイン設定 - {name}',
    fieldLabel: 'プラグイン設定 (JSON)',
    hint: '設定は JSON 形式で保存されます。プラグインが実行中の場合は保存後に設定がホットリロードされます。'
  },
  messages: {
    loadFailed: 'プラグイン一覧の取得に失敗しました',
    enabled: 'プラグイン「{name}」を有効にしました',
    disabled: 'プラグイン「{name}」を無効にしました',
    actionFailed: '操作に失敗しました',
    loadConfigFailed: '設定の読み込みに失敗しました',
    jsonInvalid: 'JSON 形式が正しくありません。設定内容を確認してください',
    configSaved: '設定を保存しました',
    saveConfigFailed: '設定の保存に失敗しました'
  }
}
