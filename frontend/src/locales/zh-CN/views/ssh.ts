// ssh 页面文案
export default {
  tabs: {
    config: 'SSH 配置',
    keys: '密钥管理',
    logs: '登录日志',
    fail2ban: 'Fail2ban 防护'
  },
  config: {
    title: 'SSH 服务配置',
    subtitle: '管理 SSH 服务端安全参数',
    apiUnavailable: '当前环境无法连接 SSH 服务（Windows 开发环境无 SSH 服务），以下为默认配置展示。'
  },
  form: {
    port: 'SSH 端口',
    portHint: '修改端口后需确保防火墙已放行新端口',
    permitRootLogin: '允许 Root 登录',
    permitPasswordLogin: '允许密码登录',
    maxAuthTries: '最大认证尝试次数',
    clientAliveInterval: '客户端保活间隔(秒)',
    clientAliveHint: '0 表示不发送保活包',
    allowUsers: '允许登录用户',
    allowUsersPlaceholder: '逗号分隔，留空表示允许全部用户',
    maxRetry: '最大重试次数',
    banDuration: '封禁时长(分钟)',
    findWindow: '检测窗口(分钟)'
  },
  status: {
    safe: '安全',
    risk: '风险',
    safeKeyOnly: '安全（仅密钥）',
    success: '成功',
    failed: '失败'
  },
  filter: {
    all: '全部'
  },
  columns: {
    type: '密钥类型',
    comment: '注释/名称',
    createdAt: '添加时间',
    actions: '操作',
    ip: 'IP 地址'
  },
  keys: {
    title: 'SSH 公钥管理',
    subtitle: '管理授权登录的公钥',
    apiUnavailable: '当前环境无法读取 SSH 密钥列表（Windows 开发环境无 SSH 服务）。',
    modalTitle: '添加 SSH 公钥',
    keyLabel: '公钥内容',
    keyPlaceholder: 'ssh-rsa AAAA... 或 ssh-ed25519 AAAA...',
    commentLabel: '注释',
    commentPlaceholder: '例如：我的笔记本电脑',
    pasteHintPrefix: '将 ',
    pasteHintOr: ' 或 ',
    pasteHintSuffix: ' 文件内容粘贴到此处',
    enterPublicKey: '请输入公钥内容',
    invalidFormat: '公钥格式不正确，应以 ssh-rsa / ssh-ed25519 / ecdsa-sha2- 开头',
    addSuccess: '公钥添加成功',
    addFailed: '添加失败',
    deleteSuccess: '密钥已删除',
    deleteFailed: '删除失败'
  },
  logs: {
    title: 'SSH 登录日志',
    subtitle: '查看 SSH 登录记录与失败尝试',
    apiUnavailable: '当前环境无法读取 SSH 日志（Windows 开发环境无 SSH 服务），以下为示例数据。',
    loadError: 'SSH 登录日志接口不可用',
    statusPlaceholder: '全部状态',
    columns: {
      time: '时间',
      user: '用户',
      ip: 'IP 地址',
      method: '登录方式',
      status: '结果'
    }
  },
  fail2ban: {
    notInstalledHeader: 'Fail2ban 未安装',
    installHint: '请先安装 Fail2ban：',
    installOr: '或',
    apiUnavailable: '当前环境无法连接 Fail2ban 服务（Windows 开发环境），以下为示例数据。',
    loadError: 'Fail2ban 状态接口不可用（可能服务器未安装 fail2ban）',
    saveSuccess: 'Fail2ban 配置已保存',
    configTitle: '封禁配置',
    configSubtitle: '配置 Fail2ban 自动封禁策略',
    manualTitle: '手动封禁',
    manualSubtitle: '主动封禁可疑 IP 地址',
    ipPlaceholder: '输入 IP 地址，例如 192.168.1.100',
    reasonPlaceholder: '封禁原因（可选）',
    listTitle: '封禁 IP 列表',
    listSubtitle: '当前被 Fail2ban 封禁的 IP',
    defaultReason: '暴力破解',
    stats: {
      protectionStatus: '防护状态',
      enabled: '已启用',
      disabled: '已禁用',
      notInstalled: '未安装',
      bannedIps: '当前封禁 IP',
      banDuration: '封禁时长',
      unitTimes: '次',
      unitMinutes: '分钟'
    },
    columns: {
      jail: '监狱',
      bannedAt: '封禁时间',
      reason: '原因'
    }
  },
  message: {
    saveConfigSuccess: 'SSH 配置已保存',
    saveFailed: '保存失败',
    enterBanIp: '请输入要封禁的 IP 地址',
    banSuccess: 'IP {ip} 已封禁',
    banFailed: '封禁失败',
    unbanSuccess: 'IP {ip} 已解封',
    unbanFailed: '解封失败'
  },
  actions: {
    saveConfig: '保存配置',
    addKey: '添加公钥',
    ban: '封禁',
    unban: '解封'
  },
  confirm: {
    saveTitle: '保存 SSH 配置',
    saveContent: '修改 SSH 配置可能导致无法连接服务器，请确保已测试新配置。确认继续？',
    saveConfirmText: '确认保存',
    deleteKeyTitle: '删除公钥',
    deleteKeyContent: '确定要删除公钥（{name}）吗？删除后该密钥将无法用于登录。',
    deleteKeyConfirmText: '删除',
    unbanTitle: '解封 IP',
    unbanContent: '确定要解封 IP {ip} 吗？解封后该 IP 可再次尝试登录。',
    unbanConfirmText: '解封'
  }
}
