// probe page texts
export default {
  pageTitle: 'Web Probe',
  pageSubtitle: 'HTTP(S)/TCP availability probing with automatic failure alerts',
  refresh: 'Refresh',
  addProbe: 'Add Probe',
  status: {
    up: 'Up',
    down: 'Down',
    unknown: 'Not checked'
  },
  switchStatus: {
    enabled: 'Enabled',
    disabled: 'Disabled'
  },
  columns: {
    name: 'Name',
    type: 'Type',
    target: 'Target',
    interval: 'Interval',
    status: 'Status',
    latency: 'Latency',
    httpCode: 'HTTP Code',
    lastChecked: 'Last Checked',
    enabled: 'Enabled',
    actions: 'Actions',
    checkedAt: 'Checked At',
    result: 'Result',
    message: 'Message'
  },
  actions: {
    checkNow: 'Check Now',
    edit: 'Edit',
    delete: 'Delete'
  },
  messages: {
    checkDone: 'Check finished',
    checkFailed: 'Check failed',
    updated: 'Probe updated',
    created: 'Probe created',
    saveFailed: 'Save failed',
    deleted: 'Probe deleted',
    deleteFailed: 'Deletion failed'
  },
  rules: {
    nameRequired: 'Please enter the probe name',
    targetRequired: 'Please enter the check target'
  },
  empty: {
    title: 'No probes yet',
    desc: 'Once added, probes keep checking target availability and push failures through the Alert Notification channels'
  },
  historyButton: 'View history of the first probe',
  form: {
    addTitle: 'Add Probe',
    editTitle: 'Edit Probe',
    nameLabel: 'Probe Name',
    namePlaceholder: 'e.g. Homepage',
    typeLabel: 'Type',
    intervalLabel: 'Check Interval (seconds)',
    targetLabel: 'Check Target',
    keywordLabel: 'Response Keyword (optional, must appear in the body)',
    keywordPlaceholder: 'e.g. ok',
    enabledLabel: 'Enabled'
  },
  results: {
    title: 'Check History · {name}'
  },
  deleteDialog: {
    title: 'Delete Probe',
    content: 'Delete probe "{name}"? Its check history will be removed as well.',
    confirm: 'Delete'
  }
}
