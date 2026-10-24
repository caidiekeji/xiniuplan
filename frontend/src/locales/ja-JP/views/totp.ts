// totp ページ文言
export default {
  title: '二要素認証 (TOTP)',
  subtitle: '時間基準のワンタイムパスワードでアカウントに追加の保護を提供します',
  statusLabel: '現在の状態：',
  status: {
    enabled: '有効',
    disabled: '無効'
  },
  boundAt: '紐付け日時：',
  compatDesc:
    'TOTP（時間基準のワンタイムパスワード）は、Google Authenticator、Microsoft Authenticator、Authy などの認証器アプリと互換性があります。',
  loginDesc:
    '有効にすると、ログイン時にパスワードに加えて認証器アプリが生成する 6 桁のコードの入力が必要になります。',
  bindNow: '今すぐ紐付ける',
  disableBtn: 'TOTP を無効化',
  recoveryNotice: {
    title: 'リカバリーコードについて',
    line1:
      '各リカバリーコードは 1 回のみ使用でき、認証器が使えない場合のログインに使用されます。必ず安全に保管してください。紛失しても復元できません。',
    line2: 'リカバリーコードを再取得するには、先に TOTP を無効化してから再度紐付けてください。'
  },
  loadStatusError: 'TOTP 状態の取得に失敗しました',
  getBindInfoError: '紐付け情報の取得に失敗しました',
  secretCopied: 'シークレットをコピーしました',
  recoveryCopied: 'リカバリーコードをコピーしました',
  copyFailed: 'コピーに失敗しました。手動でコピーしてください',
  bindSuccessMsg: 'TOTP の紐付けに成功しました',
  verifyFailed: '検証に失敗しました。コードを確認してください',
  disableSuccessMsg: 'TOTP を無効化しました',
  disableFailed: '無効化に失敗しました。コードを確認してください',
  validate: {
    codeRequired: '6 桁のコードを入力してください',
    enterDigits: '6 桁の数字を入力してください'
  },
  cancel: 'キャンセル',
  bindModal: {
    title: 'TOTP 二要素認証の紐付け',
    steps: {
      scan: {
        title: 'QR コードをスキャン',
        desc: '認証器アプリでスキャン'
      },
      verify: {
        title: '検証してリカバリーコードを保存',
        desc: 'コードを入力して紐付けを完了'
      }
    },
    qrAlt: 'TOTP QR コード',
    qrTip: '左側の QR コードを認証器アプリでスキャンしてください',
    qrManual: 'スキャンできない場合は、次のシークレットを手動で入力してください：',
    copy: 'コピー',
    codeLabel: '認証器アプリに表示される 6 桁のコードを入力',
    confirm: '検証して紐付ける'
  },
  success: {
    title: '紐付けに成功しました',
    body: 'TOTP 二要素認証が有効になりました。以下のリカバリーコードは必ず安全に保管してください。各コードは 1 回のみ使用できます。'
  },
  recoveryCount: 'リカバリーコード（全 {count} 個）',
  copyAll: 'すべてコピー',
  savedRecovery: 'リカバリーコードを保存しました',
  disableModal: {
    title: 'TOTP 二要素認証の無効化',
    desc: '無効化すると、ログイン時に二要素認証は不要になります。確認のため、現在の 6 桁のコードを入力してください。',
    codeLabel: '6 桁のコード',
    codePlaceholder: '認証器アプリに表示される 6 桁の数字を入力',
    confirm: '無効化を確認'
  }
}
