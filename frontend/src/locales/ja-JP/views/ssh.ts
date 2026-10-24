// SSH ページ文言
export default {
  tabs: {
    config: 'SSH 設定',
    keys: '鍵管理',
    logs: 'ログインログ',
    fail2ban: 'Fail2ban 保護'
  },
  config: {
    title: 'SSH サービス設定',
    subtitle: 'SSH サーバーのセキュリティパラメータを管理',
    apiUnavailable: '現在の環境から SSH サービスに接続できません（Windows 開発環境には SSH サービスがありません）。以下はデフォルト設定の表示です。'
  },
  form: {
    port: 'SSH ポート',
    portHint: 'ポート変更後、ファイアウォールで新しいポートが許可されていることを確認してください',
    permitRootLogin: 'Root ログインを許可',
    permitPasswordLogin: 'パスワードログインを許可',
    maxAuthTries: '最大認証試行回数',
    clientAliveInterval: 'クライアントキープアライブ間隔（秒）',
    clientAliveHint: '0 はキーパイブパケットを送信しないことを意味します',
    allowUsers: 'ログイン許可ユーザー',
    allowUsersPlaceholder: 'カンマ区切り。空欄ですべてのユーザーを許可',
    maxRetry: '最大リトライ回数',
    banDuration: 'BAN 期間（分）',
    findWindow: '検出ウィンドウ（分）'
  },
  status: {
    safe: '安全',
    risk: 'リスク',
    safeKeyOnly: '安全（鍵のみ）',
    success: '成功',
    failed: '失敗'
  },
  filter: {
    all: 'すべて'
  },
  columns: {
    type: '鍵の種類',
    comment: 'コメント/名前',
    createdAt: '追加日時',
    actions: '操作',
    ip: 'IP アドレス'
  },
  keys: {
    title: 'SSH 公開鍵管理',
    subtitle: 'ログインを許可する公開鍵を管理',
    apiUnavailable: '現在の環境から SSH 鍵リストを読み取れません（Windows 開発環境には SSH サービスがありません）。',
    modalTitle: 'SSH 公開鍵を追加',
    keyLabel: '公開鍵の内容',
    keyPlaceholder: 'ssh-rsa AAAA... または ssh-ed25519 AAAA...',
    commentLabel: 'コメント',
    commentPlaceholder: '例：マイノートパソコン',
    pasteHintPrefix: '',
    pasteHintOr: ' または ',
    pasteHintSuffix: ' のファイル内容をここに貼り付けてください',
    enterPublicKey: '公開鍵の内容を入力してください',
    invalidFormat: '公開鍵の形式が正しくありません。ssh-rsa / ssh-ed25519 / ecdsa-sha2- で始まる必要があります',
    addSuccess: '公開鍵を追加しました',
    addFailed: '追加に失敗しました',
    deleteSuccess: '鍵を削除しました',
    deleteFailed: '削除に失敗しました'
  },
  logs: {
    title: 'SSH ログインログ',
    subtitle: 'SSH ログイン記録と失敗した試行を確認',
    apiUnavailable: '現在の環境から SSH ログを読み取れません（Windows 開発環境には SSH サービスがありません）。以下はサンプルデータです。',
    loadError: 'SSH ログインログ API を利用できません',
    statusPlaceholder: 'すべての状態',
    columns: {
      time: '時刻',
      user: 'ユーザー',
      ip: 'IP アドレス',
      method: 'ログイン方式',
      status: '結果'
    }
  },
  fail2ban: {
    notInstalledHeader: 'Fail2ban 未インストール',
    installHint: 'まず Fail2ban をインストールしてください：',
    installOr: 'または',
    apiUnavailable: '現在の環境から Fail2ban サービスに接続できません（Windows 開発環境）。以下はサンプルデータです。',
    loadError: 'Fail2ban 状態 API を利用できません（fail2ban が未インストールの可能性があります）',
    saveSuccess: 'Fail2ban 設定を保存しました',
    configTitle: 'BAN 設定',
    configSubtitle: 'Fail2ban の自動 BAN ポリシーを設定',
    manualTitle: '手動 BAN',
    manualSubtitle: '疑わしい IP アドレスを手動で BAN',
    ipPlaceholder: 'IP アドレスを入力（例：192.168.1.100）',
    reasonPlaceholder: 'BAN 理由（任意）',
    listTitle: 'BAN 一覧',
    listSubtitle: '現在 Fail2ban により BAN されている IP',
    defaultReason: '総当たり攻撃',
    stats: {
      protectionStatus: '保護状態',
      enabled: '有効',
      disabled: '無効',
      notInstalled: '未インストール',
      bannedIps: '現在の BAN IP',
      banDuration: 'BAN 期間',
      unitTimes: '回',
      unitMinutes: '分'
    },
    columns: {
      jail: 'ジェイル',
      bannedAt: 'BAN 日時',
      reason: '理由'
    }
  },
  message: {
    saveConfigSuccess: 'SSH 設定を保存しました',
    saveFailed: '保存に失敗しました',
    enterBanIp: 'BAN する IP アドレスを入力してください',
    banSuccess: 'IP {ip} を BAN しました',
    banFailed: 'BAN に失敗しました',
    unbanSuccess: 'IP {ip} の BAN を解除しました',
    unbanFailed: 'BAN 解除に失敗しました'
  },
  actions: {
    saveConfig: '設定を保存',
    addKey: '鍵を追加',
    ban: 'BAN',
    unban: 'BAN 解除'
  },
  confirm: {
    saveTitle: 'SSH 設定を保存',
    saveContent: 'SSH 設定の変更によりサーバーに接続できなくなる可能性があります。新しい設定をテスト済みであることを確認してください。続行しますか？',
    saveConfirmText: '保存を確認',
    deleteKeyTitle: '公開鍵を削除',
    deleteKeyContent: '公開鍵（{name}）を削除しますか？削除後はこの鍵でログインできなくなります。',
    deleteKeyConfirmText: '削除',
    unbanTitle: 'IP の BAN 解除',
    unbanContent: 'IP {ip} の BAN を解除しますか？解除後、この IP は再びログインを試みることができます。',
    unbanConfirmText: '解除'
  }
}
