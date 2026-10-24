// alert page texts
export default {
  metrics: {
    cpu: 'CPU',
    memory: 'Memory',
    disk: 'Disk',
    network: 'Network'
  },
  conditions: {
    gt: 'Greater than',
    lt: 'Less than',
    eq: 'Equal to'
  },
  channels: {
    email: 'Email',
    dingtalk: 'DingTalk',
    wechat: 'WeCom',
    feishu: 'Feishu',
    telegram: 'Telegram',
    webhook: 'Webhook'
  },
  ruleStatus: {
    triggered: 'Triggered',
    recovered: 'Recovered',
    all: 'All'
  },
  notifyStatus: {
    pending: 'Sending',
    sent: 'Sent',
    failed: 'Failed'
  },
  status: {
    enabled: 'Enabled',
    disabled: 'Disabled'
  },
  actions: {
    test: 'Test',
    edit: 'Edit',
    delete: 'Delete'
  },
  durationUnit: '{duration} min',
  tabs: {
    rules: 'Alert Rules',
    channels: 'Notification Channels',
    history: 'Alert History'
  },
  rules: {
    title: 'Alert Rules',
    subtitle: 'Configure alert trigger conditions for server resource metrics',
    create: 'Create Rule',
    columns: {
      name: 'Rule Name',
      metric: 'Metric',
      condition: 'Condition',
      duration: 'Duration',
      channels: 'Channels',
      status: 'Status',
      actions: 'Actions'
    },
    form: {
      createTitle: 'Create Rule',
      editTitle: 'Edit Rule',
      name: 'Rule Name',
      namePlaceholder: 'e.g. High CPU usage',
      metric: 'Metric',
      condition: 'Condition',
      threshold: 'Threshold',
      thresholdPlaceholder: 'Unit {unit}',
      duration: 'Duration (minutes)',
      durationPlaceholder: 'Default 5 minutes',
      channels: 'Notification Channels',
      noChannels: 'No channels available. Please add one in "Notification Channels" first.',
      remark: 'Remark',
      remarkPlaceholder: 'Optional remark'
    },
    messages: {
      enabled: 'Rule enabled',
      enableFailed: 'Enable failed',
      disabled: 'Rule disabled',
      disableFailed: 'Disable failed',
      tested: 'Test message sent',
      testFailed: 'Test failed',
      updated: 'Rule updated',
      created: 'Rule created',
      saveFailed: 'Save failed',
      deleted: 'Rule deleted',
      deleteFailed: 'Delete failed',
      nameRequired: 'Please enter a rule name',
      thresholdRequired: 'Please enter a threshold'
    },
    deleteConfirm: {
      title: 'Delete Rule',
      content: 'Are you sure to delete rule "{name}"? This action cannot be undone.',
      confirm: 'Delete'
    }
  },
  channelsTab: {
    title: 'Notification Channels',
    subtitle: 'Manage notification channels (Email / DingTalk / WeCom / Feishu / Telegram / Webhook)',
    create: 'Add Channel',
    columns: {
      name: 'Channel Name',
      type: 'Type',
      status: 'Status',
      createdAt: 'Created At',
      actions: 'Actions'
    },
    form: {
      createTitle: 'Add Channel',
      editTitle: 'Edit Channel',
      name: 'Channel Name',
      namePlaceholder: 'e.g. Ops alert email',
      type: 'Channel Type',
      emailRecipients: 'Recipients (comma separated)',
      emailHint: 'SMTP server configuration is managed in System Settings',
      dingtalkWebhook: 'Webhook URL',
      dingtalkSecret: 'Signing Secret (optional)',
      wechatWebhook: 'Webhook URL',
      feishuWebhook: 'Webhook URL',
      telegramBotToken: 'Bot Token',
      telegramChatId: 'Chat ID',
      telegramChatIdPlaceholder: "{'@'}channelname or numeric ID",
      webhookUrl: 'URL',
      webhookHeaders: 'Headers (JSON, optional)',
      enabled: 'Enabled'
    },
    messages: {
      tested: 'Test message sent',
      testFailed: 'Test failed',
      updated: 'Channel updated',
      created: 'Channel created',
      saveFailed: 'Save failed',
      deleted: 'Channel deleted',
      deleteFailed: 'Delete failed',
      nameRequired: 'Please enter a channel name'
    },
    deleteConfirm: {
      title: 'Delete Channel',
      content: 'Are you sure to delete channel "{name}"? Related rules will no longer be able to send notifications through it.',
      confirm: 'Delete'
    }
  },
  history: {
    title: 'Alert History',
    subtitle: 'View all alert trigger and recovery records',
    filterRule: 'Rule',
    filterRulePlaceholder: 'All rules',
    allRules: 'All rules',
    filterStatus: 'Status',
    filterStatusPlaceholder: 'All statuses',
    reset: 'Reset',
    empty: 'No alert records',
    emptyHint: 'Records will appear here when configured rules are triggered',
    errorLabel: 'Notification error: ',
    columns: {
      triggeredAt: 'Triggered At',
      ruleName: 'Rule Name',
      metric: 'Metric',
      triggerValue: 'Trigger Value',
      threshold: 'Threshold',
      status: 'Status',
      notifyStatus: 'Notify Status',
      recoveredAt: 'Recovered At'
    }
  }
}
