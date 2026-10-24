// userList ページ文言
export default {
  title: 'ユーザー一覧',
  subtitle: 'システムユーザー、ロール割り当て、アカウント状態を管理します',
  createUser: 'ユーザーを作成',
  editUser: 'ユーザーを編集',
  searchPlaceholder: 'ユーザー名で検索...',
  roles: {
    superAdmin: 'スーパー管理者',
    admin: '管理者',
    user: '一般ユーザー'
  },
  columns: {
    username: 'ユーザー名',
    role: 'ロール',
    status: '状態',
    lastLogin: '最終ログイン',
    createdAt: '作成日時',
    actions: '操作'
  },
  status: {
    enabled: '有効',
    disabled: '無効',
    active: '正常',
    inactive: '停止中',
    on: '有効',
    off: '無効'
  },
  actions: {
    edit: '編集',
    password: 'パスワード変更',
    enable: '有効化',
    disable: '無効化',
    delete: '削除'
  },
  labels: {
    username: 'ユーザー名',
    password: 'パスワード',
    passwordOptional: 'パスワード（空欄で変更なし）',
    confirmPassword: 'パスワード確認',
    role: 'ロール',
    status: '状態',
    newPassword: '新しいパスワード',
    confirmNewPassword: '新しいパスワードの確認'
  },
  hints: {
    enterUsername: 'ユーザー名を入力してください',
    enterPassword: 'パスワードを入力してください',
    reenterPassword: 'パスワードを再入力してください',
    confirmPassword: 'パスワードを確認してください',
    selectRole: 'ロールを選択してください',
    enterNewPassword: '新しいパスワードを入力してください',
    confirmNewPassword: '新しいパスワードを確認してください',
    reenterNewPassword: '新しいパスワードを再入力してください',
    passwordMismatch: '2 回入力したパスワードが一致しません'
  },
  strength: {
    label: 'パスワード強度：{level}',
    weak: '弱い',
    medium: '普通',
    strong: '強い'
  },
  loadError: 'ユーザー一覧の読み込みに失敗しました',
  enableSuccess: 'ユーザーを有効化しました',
  enableFailed: '有効化に失敗しました',
  disableSuccess: 'ユーザーを無効化しました',
  disableFailed: '無効化に失敗しました',
  messages: {
    saveFailed: '保存に失敗しました',
    updated: 'ユーザーを更新しました',
    created: 'ユーザーを作成しました',
    passwordChanged: 'パスワードを変更しました',
    passwordFailed: 'パスワードの変更に失敗しました',
    deleted: 'ユーザーを削除しました',
    deleteFailed: '削除に失敗しました'
  },
  pwdTitle: 'パスワード変更 — {username}',
  delete: {
    title: 'ユーザーの削除',
    content: 'ユーザー「{name}」を削除しますか？この操作は元に戻せません。'
  }
}
