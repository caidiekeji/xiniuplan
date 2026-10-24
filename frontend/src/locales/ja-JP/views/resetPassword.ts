// resetPassword ページ文案
export default {
  title: 'パスワードリセット',
  subtitle: '新しいログインパスワードを設定してください',
  invalidToken: 'リセットリンクが無効です：トークンパラメータがありません。メール内のリンクからアクセスしてください',
  placeholders: {
    newPassword: '新しいパスワード（6 文字以上）',
    confirmPassword: '新しいパスワード（確認用）'
  },
  rules: {
    newPasswordRequired: '新しいパスワードを入力してください',
    newPasswordMin: '新しいパスワードは 6 文字以上にしてください',
    confirmPasswordRequired: '新しいパスワードを再入力してください',
    passwordMismatch: '入力されたパスワードが一致しません'
  },
  submit: 'パスワードをリセット',
  backToLogin: 'ログインに戻る',
  success: 'パスワードをリセットしました。まもなくログインページに移動します…',
  failed: 'パスワードのリセットに失敗しました'
}
