// resetPassword 页面文案
export default {
  title: '重置密码',
  subtitle: '请设置新的登录密码',
  invalidToken: '重置链接无效：缺少令牌参数，请从邮件中的链接进入',
  placeholders: {
    newPassword: '新密码（至少 6 位）',
    confirmPassword: '确认新密码'
  },
  rules: {
    newPasswordRequired: '请输入新密码',
    newPasswordMin: '新密码至少 6 位',
    confirmPasswordRequired: '请再次输入新密码',
    passwordMismatch: '两次输入的密码不一致'
  },
  submit: '重置密码',
  backToLogin: '返回登录',
  success: '密码重置成功，即将跳转到登录页…',
  failed: '密码重置失败'
}
