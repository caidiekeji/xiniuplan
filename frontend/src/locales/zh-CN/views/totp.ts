// totp 页面文案
export default {
  title: '二次验证 (TOTP)',
  subtitle: '基于时间的一次性密码，为账户提供额外安全保护',
  statusLabel: '当前状态：',
  status: {
    enabled: '已启用',
    disabled: '未启用'
  },
  boundAt: '绑定时间：',
  compatDesc: 'TOTP（基于时间的一次性密码）兼容 Google Authenticator、Microsoft Authenticator、Authy 等验证器应用。',
  loginDesc: '启用后，登录时除密码外还需输入验证器生成的 6 位动态码。',
  bindNow: '立即绑定',
  disableBtn: '禁用 TOTP',
  recoveryNotice: {
    title: '恢复码说明',
    line1: '每个恢复码只能使用一次，用于在无法使用验证器时登录账户。请妥善保存恢复码，丢失后无法找回。',
    line2: '如需重新获取恢复码，请先禁用 TOTP 再重新绑定。'
  },
  loadStatusError: '获取 TOTP 状态失败',
  getBindInfoError: '获取绑定信息失败',
  secretCopied: '密钥已复制',
  recoveryCopied: '恢复码已复制',
  copyFailed: '复制失败，请手动复制',
  bindSuccessMsg: 'TOTP 绑定成功',
  verifyFailed: '验证失败，请检查动态码',
  disableSuccessMsg: 'TOTP 已禁用',
  disableFailed: '禁用失败，请检查动态码',
  validate: {
    codeRequired: '请输入 6 位动态码',
    enterDigits: '请输入 6 位数字'
  },
  cancel: '取消',
  bindModal: {
    title: '绑定 TOTP 二次验证',
    steps: {
      scan: {
        title: '扫描二维码',
        desc: '使用验证器 App 扫描'
      },
      verify: {
        title: '验证并保存恢复码',
        desc: '输入动态码完成绑定'
      }
    },
    qrAlt: 'TOTP 二维码',
    qrTip: '使用验证器 App 扫描左侧二维码',
    qrManual: '如无法扫描，可手动输入以下密钥：',
    copy: '复制',
    codeLabel: '输入验证器显示的 6 位动态码',
    confirm: '验证并绑定'
  },
  success: {
    title: '绑定成功',
    body: 'TOTP 二次验证已启用。请妥善保存以下恢复码，每个恢复码只能使用一次。'
  },
  recoveryCount: '恢复码（共 {count} 个）',
  copyAll: '全部复制',
  savedRecovery: '我已保存恢复码',
  disableModal: {
    title: '禁用 TOTP 二次验证',
    desc: '禁用后，登录将不再需要二次验证。请输入当前 6 位动态码以确认操作。',
    codeLabel: '6 位动态码',
    codePlaceholder: '请输入验证器显示的 6 位数字',
    confirm: '确认禁用'
  }
}
