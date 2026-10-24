// sessions page texts
export default {
  apiKeyTab: 'API Key',
  title: 'Online Sessions',
  refresh: 'Refresh',
  createKey: 'Create API Key',
  cancel: 'Cancel',
  create: 'Create',
  columns: {
    user: 'User',
    type: 'Type',
    loginAt: 'Login Time',
    lastActive: 'Last Active',
    actions: 'Actions',
    name: 'Name',
    status: 'Status',
    lastUsed: 'Last Used',
    expiresAt: 'Expires At'
  },
  actions: {
    revoke: 'Revoke',
    delete: 'Delete'
  },
  status: {
    active: 'Active',
    inactive: 'Disabled',
    neverUsed: 'Never',
    neverExpires: 'Never'
  },
  form: {
    name: 'Name',
    namePlaceholder: 'e.g. CI deploy script',
    expiresInDays: 'Valid for (days)',
    neverExpiresHint: '0 = never expires'
  },
  secret: {
    title: 'Save Your API Key Now',
    warning: 'The Secret is shown only once and cannot be viewed after closing! Send it via header',
    saved: 'I Have Saved It'
  },
  deleteDialog: {
    title: 'Delete API Key',
    content: 'Delete "{name}"? Programs using this key will stop working immediately.'
  },
  messages: {
    loadSessionsFailed: 'Failed to load sessions',
    revoked: 'Session revoked',
    revokeFailed: 'Revoke failed',
    loadKeysFailed: 'Failed to load API Keys',
    nameRequired: 'Please enter a name',
    createFailed: 'Creation failed',
    deleted: 'Deleted'
  }
}
