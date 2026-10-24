// ssl 页面文案
export default {
  title: 'SSL 证书',
  subtitle: 'HTTPS 证书申请、续期与管理',
  provider: {
    manual: '手动上传'
  },
  status: {
    valid: '有效',
    expiring: '即将到期',
    expired: '已过期',
    pending: '签发中'
  },
  columns: {
    domain: '域名',
    provider: '证书提供方',
    expireAt: '到期时间',
    status: '状态',
    actions: '操作'
  },
  expireText: {
    expired: '已过期 {days} 天',
    remaining: '剩余 {days} 天'
  },
  actions: {
    apply: '申请证书',
    detail: '详情',
    renew: '续期',
    delete: '删除'
  },
  detail: {
    title: '证书详情',
    providerLabel: '提供方',
    issuer: '签发机构',
    email: '联系邮箱'
  },
  messages: {
    loadListFailed: '加载 SSL 证书列表失败',
    renewSuccess: '证书续期成功',
    renewPending: '自动续期尚未实现：服务器端只登记了申请，未真正执行签发。可先手动上传证书。',
    emailRequired: '请先在申请表单中填写邮箱，再执行续期',
    renewFailed: '续期失败',
    applySuccess: '证书申请提交成功',
    applyFailed: '申请失败',
    deleteSuccess: '证书已删除',
    deleteFailed: '删除失败'
  },
  rules: {
    selectSite: '请选择站点',
    email: '请输入邮箱地址',
    cert: '请输入证书内容',
    key: '请输入私钥内容',
    dnsProvider: '请选择 DNS 提供商'
  },
  apply: {
    title: '申请 SSL 证书',
    selectSite: '选择站点',
    sitePlaceholder: '选择要绑定证书的站点',
    methodLabel: '申请方式',
    letsencryptOption: "Let's Encrypt（自动签发）",
    emailLabel: '邮箱地址',
    emailPlaceholder: '用于接收证书到期提醒',
    verificationLabel: '验证方式',
    httpOption: 'HTTP 验证',
    dnsOption: 'DNS 验证',
    verificationNoteLabel: '验证说明',
    httpNote: '将自动在站点目录放置验证文件完成 HTTP-01 验证',
    dnsProviderLabel: 'DNS 提供商',
    dnsProviderPlaceholder: '选择已配置的 DNS 提供商账号',
    dnsNote: '将自动添加 TXT 记录完成 DNS-01 验证，无需手动操作',
    certLabel: '证书内容 (PEM)',
    keyLabel: '私钥内容 (PEM)',
    dnsAccountLabel: '{name}（{type}）'
  },
  applyStatus: {
    dnsAdding: '正在添加 DNS TXT 记录...'
  },
  confirm: {
    deleteTitle: '删除证书',
    deleteContent: '确定要删除域名「{domain}」的 SSL 证书吗？删除后该站点将无法使用 HTTPS。',
    deleteConfirmText: '删除'
  }
}
