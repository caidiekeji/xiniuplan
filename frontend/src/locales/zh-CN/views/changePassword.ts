// changePassword 页面文案
export default {
  title: '修改密码',
  subtitle: '出于安全考虑，首次登录（或密码被重置后）必须修改密码后才能继续使用面板',
  placeholders: {
    oldPassword: '原密码（初始密码见服务启动日志）',
    newPassword: '新密码（至少 8 位，含字母和数字）',
    confirmPassword: '确认新密码'
  },
  rules: {
    oldPasswordRequired: '请输入原密码',
    newPasswordRequired: '请输入新密码',
    newPasswordMin: '新密码至少 8 位',
    newPasswordComplexity: '新密码需同时包含字母和数字',
    newPasswordSameAsOld: '新密码不能与原密码相同',
    confirmPasswordRequired: '请再次输入新密码',
    passwordMismatch: '两次输入的密码不一致'
  },
  submit: '确认修改',
  success: '密码修改成功，请使用新密码重新登录',
  failed: '密码修改失败'
}
