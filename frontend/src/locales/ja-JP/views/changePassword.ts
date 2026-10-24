// changePassword ページ文案
export default {
  title: 'パスワード変更',
  subtitle: 'セキュリティのため、初回ログイン時（またはパスワードリセット後）はパスワードを変更してからパネルをご利用ください',
  placeholders: {
    oldPassword: '現在のパスワード（初期パスワードはサーバー起動ログを確認）',
    newPassword: '新しいパスワード（8 文字以上、英字と数字を含む）',
    confirmPassword: '新しいパスワード（確認用）'
  },
  rules: {
    oldPasswordRequired: '現在のパスワードを入力してください',
    newPasswordRequired: '新しいパスワードを入力してください',
    newPasswordMin: '新しいパスワードは 8 文字以上にしてください',
    newPasswordComplexity: '新しいパスワードには英字と数字の両方を含めてください',
    newPasswordSameAsOld: '新しいパスワードは現在のパスワードと異なる必要があります',
    confirmPasswordRequired: '新しいパスワードを再入力してください',
    passwordMismatch: '入力されたパスワードが一致しません'
  },
  submit: '変更を確認',
  success: 'パスワードを変更しました。新しいパスワードでログインしてください',
  failed: 'パスワードの変更に失敗しました'
}
