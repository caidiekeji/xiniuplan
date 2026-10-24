// alert ページ文言
export default {
  metrics: {
    cpu: 'CPU',
    memory: 'メモリ',
    disk: 'ディスク',
    network: 'ネットワーク'
  },
  conditions: {
    gt: 'より大きい',
    lt: 'より小さい',
    eq: '等しい'
  },
  channels: {
    email: 'メール',
    dingtalk: 'DingTalk',
    wechat: 'WeCom',
    feishu: 'Feishu',
    telegram: 'Telegram',
    webhook: 'Webhook'
  },
  ruleStatus: {
    triggered: '発火中',
    recovered: '回復済み',
    all: 'すべて'
  },
  notifyStatus: {
    pending: '送信中',
    sent: '送信済み',
    failed: '失敗'
  },
  status: {
    enabled: '有効',
    disabled: '無効'
  },
  actions: {
    test: 'テスト',
    edit: '編集',
    delete: '削除'
  },
  durationUnit: '{duration} 分',
  tabs: {
    rules: 'アラートルール',
    channels: '通知チャネル',
    history: 'アラート履歴'
  },
  rules: {
    title: 'アラートルール',
    subtitle: 'サーバーリソース指標のアラート発火条件を設定します',
    create: 'ルール作成',
    columns: {
      name: 'ルール名',
      metric: '指標タイプ',
      condition: '条件',
      duration: '継続時間',
      channels: '通知チャネル',
      status: '状態',
      actions: '操作'
    },
    form: {
      createTitle: 'ルール作成',
      editTitle: 'ルール編集',
      name: 'ルール名',
      namePlaceholder: '例：CPU 使用率が高い',
      metric: '指標タイプ',
      condition: '条件',
      threshold: 'しきい値',
      thresholdPlaceholder: '単位 {unit}',
      duration: '継続時間（分）',
      durationPlaceholder: 'デフォルト 5 分',
      channels: '通知チャネル',
      noChannels: '利用可能なチャネルがありません。先に「通知チャネル」で追加してください',
      remark: '備考',
      remarkPlaceholder: '任意の備考'
    },
    messages: {
      enabled: 'ルールを有効にしました',
      enableFailed: '有効化に失敗しました',
      disabled: 'ルールを無効にしました',
      disableFailed: '無効化に失敗しました',
      tested: 'テストメッセージを送信しました',
      testFailed: 'テストに失敗しました',
      updated: 'ルールを更新しました',
      created: 'ルールを作成しました',
      saveFailed: '保存に失敗しました',
      deleted: 'ルールを削除しました',
      deleteFailed: '削除に失敗しました',
      nameRequired: 'ルール名を入力してください',
      thresholdRequired: 'しきい値を入力してください'
    },
    deleteConfirm: {
      title: 'ルール削除',
      content: 'ルール「{name}」を削除しますか？この操作は元に戻せません。',
      confirm: '削除'
    }
  },
  channelsTab: {
    title: '通知チャネル',
    subtitle: '通知メッセージの送信チャネルを管理します（メール / DingTalk / WeCom / Feishu / Telegram / Webhook）',
    create: 'チャネル追加',
    columns: {
      name: 'チャネル名',
      type: 'タイプ',
      status: '状態',
      createdAt: '作成日時',
      actions: '操作'
    },
    form: {
      createTitle: 'チャネル追加',
      editTitle: 'チャネル編集',
      name: 'チャネル名',
      namePlaceholder: '例：運用アラートメール',
      type: 'チャネルタイプ',
      emailRecipients: '受信者（カンマ区切り）',
      emailHint: 'SMTP サーバー設定はシステム設定で管理されます',
      dingtalkWebhook: 'Webhook URL',
      dingtalkSecret: '署名シークレット（任意）',
      wechatWebhook: 'Webhook URL',
      feishuWebhook: 'Webhook URL',
      telegramBotToken: 'Bot Token',
      telegramChatId: 'Chat ID',
      telegramChatIdPlaceholder: "{'@'}channelname または数値 ID",
      webhookUrl: 'URL',
      webhookHeaders: 'リクエストヘッダー（JSON 形式・任意）',
      enabled: '有効'
    },
    messages: {
      tested: 'テストメッセージを送信しました',
      testFailed: 'テストに失敗しました',
      updated: 'チャネルを更新しました',
      created: 'チャネルを作成しました',
      saveFailed: '保存に失敗しました',
      deleted: 'チャネルを削除しました',
      deleteFailed: '削除に失敗しました',
      nameRequired: 'チャネル名を入力してください'
    },
    deleteConfirm: {
      title: 'チャネル削除',
      content: 'チャネル「{name}」を削除しますか？関連ルールはこのチャネルで通知を送信できなくなります。',
      confirm: '削除'
    }
  },
  history: {
    title: 'アラート履歴',
    subtitle: 'すべてのアラート発火と回復の記録を確認します',
    filterRule: 'ルール',
    filterRulePlaceholder: 'すべてのルール',
    allRules: 'すべてのルール',
    filterStatus: '状態',
    filterStatusPlaceholder: 'すべての状態',
    reset: 'リセット',
    empty: 'アラート記録はありません',
    emptyHint: '設定されたルールが発火すると、ここに記録が表示されます',
    errorLabel: '通知エラー：',
    columns: {
      triggeredAt: '発火日時',
      ruleName: 'ルール名',
      metric: '指標タイプ',
      triggerValue: '発火値',
      threshold: 'しきい値',
      status: '状態',
      notifyStatus: '通知状態',
      recoveredAt: '回復日時'
    }
  }
}
