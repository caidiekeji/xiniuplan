// audit page texts
export default {
  resultOptions: {
    all: 'All',
    success: 'Success',
    failed: 'Failed'
  },
  actionOptions: {
    all: 'All actions',
    login: 'Login',
    createSite: 'Create site',
    deleteSite: 'Delete site',
    createDatabase: 'Create database',
    installApp: 'Install app',
    restartService: 'Restart service',
    fileOperation: 'File operation',
    systemSettings: 'System settings',
    firewallRule: 'Firewall rule'
  },
  columns: {
    time: 'Time',
    operator: 'Operator',
    actionType: 'Action',
    resource: 'Resource',
    result: 'Result',
    username: 'Username',
    subject: 'Subject',
    count: 'Count',
    lastSeen: 'Last Seen',
    detail: 'Detail',
    type: 'Type'
  },
  messages: {
    loadAuditFailed: 'Failed to load audit logs',
    loadLoginFailed: 'Failed to load login logs',
    exportSuccess: 'Logs exported successfully',
    exportMock: 'Export task created (mock)',
    detectFailed: 'Detection failed'
  },
  anomalyTypes: {
    bruteForceIp: 'Brute force',
    bruteForceUser: 'Credential stuffing',
    newIpLogin: 'New device login'
  },
  tabs: {
    operations: 'Operations',
    anomalies: 'Anomalies',
    logins: 'Logins'
  },
  audit: {
    title: 'Operations Log',
    subtitle: 'Complete audit records of all panel operations',
    userId: 'User ID',
    userIdPlaceholder: 'User ID',
    actionType: 'Action',
    result: 'Result',
    timeRange: 'Time Range'
  },
  anomalies: {
    title: 'Anomaly Login Detection',
    subtitle: 'Last 24 hours: brute force / credential stuffing / new device login (S-012)',
    refresh: 'Re-detect'
  },
  logins: {
    title: 'Login Log',
    subtitle: 'Panel login records and anomaly login tracking',
    exportCsv: 'Export CSV'
  }
}
