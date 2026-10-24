// SSL ページ文言
export default {
  title: 'SSL 証明書',
  subtitle: 'HTTPS 証明書の申請・更新・管理',
  provider: {
    manual: '手動アップロード'
  },
  status: {
    valid: '有効',
    expiring: 'まもなく期限切れ',
    expired: '期限切れ',
    pending: '発行中'
  },
  columns: {
    domain: 'ドメイン',
    provider: '証明書プロバイダー',
    expireAt: '有効期限',
    status: '状態',
    actions: '操作'
  },
  expireText: {
    expired: '{days} 日前に期限切れ',
    remaining: '残り {days} 日'
  },
  actions: {
    apply: '証明書を申請',
    detail: '詳細',
    renew: '更新',
    delete: '削除'
  },
  detail: {
    title: '証明書の詳細',
    providerLabel: 'プロバイダー',
    issuer: '発行機関',
    email: '連絡先メール'
  },
  messages: {
    loadListFailed: 'SSL 証明書リストの読み込みに失敗しました',
    renewSuccess: '証明書を更新しました',
    renewPending: '自動更新はまだ実装されていません：サーバーは申請を記録するだけで，实际には証明書を発行しません。手動で証明書をアップロードしてください。',
    emailRequired: '更新の前に申請フォームでメールアドレスを入力してください',
    renewFailed: '更新に失敗しました',
    applySuccess: '証明書の申請を送信しました',
    applyFailed: '申請に失敗しました',
    deleteSuccess: '証明書を削除しました',
    deleteFailed: '削除に失敗しました'
  },
  rules: {
    selectSite: 'サイトを選択してください',
    email: 'メールアドレスを入力してください',
    cert: '証明書の内容を入力してください',
    key: '秘密鍵の内容を入力してください',
    dnsProvider: 'DNS プロバイダーを選択してください'
  },
  apply: {
    title: 'SSL 証明書を申請',
    selectSite: 'サイトを選択',
    sitePlaceholder: '証明書を紐付けるサイトを選択',
    methodLabel: '申請方法',
    letsencryptOption: "Let's Encrypt（自動発行）",
    emailLabel: 'メールアドレス',
    emailPlaceholder: '証明書の期限切れ通知の受信に使用',
    verificationLabel: '検証方法',
    httpOption: 'HTTP 検証',
    dnsOption: 'DNS 検証',
    verificationNoteLabel: '検証の説明',
    httpNote: 'サイトディレクトリに検証ファイルを自動配置し、HTTP-01 検証を完了します',
    dnsProviderLabel: 'DNS プロバイダー',
    dnsProviderPlaceholder: '設定済みの DNS プロバイダーアカウントを選択',
    dnsNote: 'TXT レコードを自動追加して DNS-01 検証を完了します。手動作業は不要です',
    certLabel: '証明書 (PEM)',
    keyLabel: '秘密鍵 (PEM)',
    dnsAccountLabel: '{name}（{type}）'
  },
  applyStatus: {
    dnsAdding: 'DNS TXT レコードを追加中...'
  },
  confirm: {
    deleteTitle: '証明書を削除',
    deleteContent: 'ドメイン「{domain}」の SSL 証明書を削除しますか？削除後、このサイトは HTTPS を使用できなくなります。',
    deleteConfirmText: '削除'
  }
}
