// alert 页面文案
export default {
  metrics: {
    cpu: 'CPU',
    memory: '内存',
    disk: '磁盘',
    network: '网络'
  },
  conditions: {
    gt: '大于',
    lt: '小于',
    eq: '等于'
  },
  channels: {
    email: '邮件',
    dingtalk: '钉钉',
    wechat: '企业微信',
    feishu: '飞书',
    telegram: 'Telegram',
    webhook: 'Webhook'
  },
  ruleStatus: {
    triggered: '触发中',
    recovered: '已恢复',
    all: '全部'
  },
  notifyStatus: {
    pending: '发送中',
    sent: '已发送',
    failed: '失败'
  },
  status: {
    enabled: '启用',
    disabled: '禁用'
  },
  actions: {
    test: '测试',
    edit: '编辑',
    delete: '删除'
  },
  durationUnit: '{duration} 分钟',
  tabs: {
    rules: '告警规则',
    channels: '通知渠道',
    history: '告警历史'
  },
  rules: {
    title: '告警规则',
    subtitle: '配置服务器资源指标的告警触发条件',
    create: '创建规则',
    columns: {
      name: '规则名称',
      metric: '指标类型',
      condition: '条件',
      duration: '持续时间',
      channels: '通知渠道',
      status: '状态',
      actions: '操作'
    },
    form: {
      createTitle: '创建规则',
      editTitle: '编辑规则',
      name: '规则名称',
      namePlaceholder: '例如：CPU 使用率过高',
      metric: '指标类型',
      condition: '条件',
      threshold: '阈值',
      thresholdPlaceholder: '单位 {unit}',
      duration: '持续时间（分钟）',
      durationPlaceholder: '默认 5 分钟',
      channels: '通知渠道',
      noChannels: '暂无可用渠道，请先在「通知渠道」中添加',
      remark: '备注',
      remarkPlaceholder: '可选备注信息'
    },
    messages: {
      enabled: '规则已启用',
      enableFailed: '启用失败',
      disabled: '规则已禁用',
      disableFailed: '禁用失败',
      tested: '测试消息已发送',
      testFailed: '测试失败',
      updated: '规则已更新',
      created: '规则已创建',
      saveFailed: '保存失败',
      deleted: '规则已删除',
      deleteFailed: '删除失败',
      nameRequired: '请输入规则名称',
      thresholdRequired: '请输入阈值'
    },
    deleteConfirm: {
      title: '删除规则',
      content: '确定要删除规则「{name}」吗？此操作不可撤销。',
      confirm: '删除'
    }
  },
  channelsTab: {
    title: '通知渠道',
    subtitle: '管理告警消息的发送渠道（邮件 / 钉钉 / 企微 / 飞书 / Telegram / Webhook）',
    create: '添加渠道',
    columns: {
      name: '渠道名称',
      type: '类型',
      status: '状态',
      createdAt: '创建时间',
      actions: '操作'
    },
    form: {
      createTitle: '添加渠道',
      editTitle: '编辑渠道',
      name: '渠道名称',
      namePlaceholder: '例如：运维告警邮件',
      type: '渠道类型',
      emailRecipients: '收件人（多个邮箱用逗号分隔）',
      emailHint: 'SMTP 服务器配置在系统设置中管理',
      dingtalkWebhook: 'Webhook URL',
      dingtalkSecret: '加签密钥（可选）',
      wechatWebhook: 'Webhook URL',
      feishuWebhook: 'Webhook URL',
      telegramBotToken: 'Bot Token',
      telegramChatId: 'Chat ID',
      telegramChatIdPlaceholder: "{'@'}channelname 或数字 ID",
      webhookUrl: 'URL',
      webhookHeaders: '请求头（JSON 格式，可选）',
      enabled: '启用'
    },
    messages: {
      tested: '测试消息已发送',
      testFailed: '测试失败',
      updated: '渠道已更新',
      created: '渠道已创建',
      saveFailed: '保存失败',
      deleted: '渠道已删除',
      deleteFailed: '删除失败',
      nameRequired: '请输入渠道名称'
    },
    deleteConfirm: {
      title: '删除渠道',
      content: '确定要删除渠道「{name}」吗？关联的规则将无法通过此渠道发送通知。',
      confirm: '删除'
    }
  },
  history: {
    title: '告警历史',
    subtitle: '查看所有告警触发与恢复记录',
    filterRule: '规则',
    filterRulePlaceholder: '全部规则',
    allRules: '全部规则',
    filterStatus: '状态',
    filterStatusPlaceholder: '全部状态',
    reset: '重置',
    empty: '暂无告警记录',
    emptyHint: '当配置的规则被触发时，记录会显示在这里',
    errorLabel: '通知错误：',
    columns: {
      triggeredAt: '触发时间',
      ruleName: '规则名称',
      metric: '指标类型',
      triggerValue: '触发值',
      threshold: '阈值',
      status: '状态',
      notifyStatus: '通知状态',
      recoveredAt: '恢复时间'
    }
  }
}
