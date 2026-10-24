// SSH page texts
export default {
  tabs: {
    config: 'SSH Config',
    keys: 'Key Management',
    logs: 'Login Logs',
    fail2ban: 'Fail2ban Protection'
  },
  config: {
    title: 'SSH Service Configuration',
    subtitle: 'Manage SSH server security parameters',
    apiUnavailable: 'Unable to connect to the SSH service in the current environment (no SSH service in the Windows development environment); the default configuration is displayed below.'
  },
  form: {
    port: 'SSH Port',
    portHint: 'After changing the port, make sure the firewall allows the new port',
    permitRootLogin: 'Allow Root Login',
    permitPasswordLogin: 'Allow Password Login',
    maxAuthTries: 'Max Authentication Tries',
    clientAliveInterval: 'Client Alive Interval (seconds)',
    clientAliveHint: '0 means no keep-alive packets are sent',
    allowUsers: 'Allowed Users',
    allowUsersPlaceholder: 'Separate with commas; leave empty to allow all users',
    maxRetry: 'Max Retry Count',
    banDuration: 'Ban Duration (minutes)',
    findWindow: 'Detection Window (minutes)'
  },
  status: {
    safe: 'Safe',
    risk: 'Risky',
    safeKeyOnly: 'Safe (key only)',
    success: 'Success',
    failed: 'Failed'
  },
  filter: {
    all: 'All'
  },
  columns: {
    type: 'Key Type',
    comment: 'Comment/Name',
    createdAt: 'Added At',
    actions: 'Actions',
    ip: 'IP Address'
  },
  keys: {
    title: 'SSH Public Keys',
    subtitle: 'Manage public keys authorized to log in',
    apiUnavailable: 'Unable to read the SSH key list in the current environment (no SSH service in the Windows development environment).',
    modalTitle: 'Add SSH Public Key',
    keyLabel: 'Public Key',
    keyPlaceholder: 'ssh-rsa AAAA... or ssh-ed25519 AAAA...',
    commentLabel: 'Comment',
    commentPlaceholder: 'e.g. My laptop',
    pasteHintPrefix: 'Paste the contents of the ',
    pasteHintOr: ' or ',
    pasteHintSuffix: ' file here',
    enterPublicKey: 'Please enter the public key content',
    invalidFormat: 'Incorrect public key format; it should start with ssh-rsa / ssh-ed25519 / ecdsa-sha2-',
    addSuccess: 'Public key added',
    addFailed: 'Add failed',
    deleteSuccess: 'Key deleted',
    deleteFailed: 'Delete failed'
  },
  logs: {
    title: 'SSH Login Logs',
    subtitle: 'View SSH login records and failed attempts',
    apiUnavailable: 'Unable to read SSH logs in the current environment (no SSH service in the Windows development environment); sample data is shown below.',
    loadError: 'SSH login log API is unavailable',
    statusPlaceholder: 'All statuses',
    columns: {
      time: 'Time',
      user: 'User',
      ip: 'IP Address',
      method: 'Method',
      status: 'Result'
    }
  },
  fail2ban: {
    notInstalledHeader: 'Fail2ban Not Installed',
    installHint: 'Please install Fail2ban first: ',
    installOr: 'or',
    apiUnavailable: 'Unable to connect to the Fail2ban service in the current environment (Windows development environment); sample data is shown below.',
    loadError: 'Fail2ban status API is unavailable (fail2ban may not be installed)',
    saveSuccess: 'Fail2ban configuration saved',
    configTitle: 'Ban Configuration',
    configSubtitle: 'Configure Fail2ban auto-ban policy',
    manualTitle: 'Manual Ban',
    manualSubtitle: 'Proactively ban suspicious IP addresses',
    ipPlaceholder: 'Enter an IP address, e.g. 192.168.1.100',
    reasonPlaceholder: 'Ban reason (optional)',
    listTitle: 'Banned IP List',
    listSubtitle: 'IPs currently banned by Fail2ban',
    defaultReason: 'Brute force',
    stats: {
      protectionStatus: 'Protection Status',
      enabled: 'Enabled',
      disabled: 'Disabled',
      notInstalled: 'Not installed',
      bannedIps: 'Banned IPs',
      banDuration: 'Ban Duration',
      unitTimes: 'times',
      unitMinutes: 'minutes'
    },
    columns: {
      jail: 'Jail',
      bannedAt: 'Banned At',
      reason: 'Reason'
    }
  },
  message: {
    saveConfigSuccess: 'SSH configuration saved',
    saveFailed: 'Save failed',
    enterBanIp: 'Please enter the IP address to ban',
    banSuccess: 'IP {ip} banned',
    banFailed: 'Ban failed',
    unbanSuccess: 'IP {ip} unbanned',
    unbanFailed: 'Unban failed'
  },
  actions: {
    saveConfig: 'Save Config',
    addKey: 'Add Key',
    ban: 'Ban',
    unban: 'Unban'
  },
  confirm: {
    saveTitle: 'Save SSH Configuration',
    saveContent: 'Changing the SSH configuration may prevent you from connecting to the server. Make sure the new configuration has been tested. Continue?',
    saveConfirmText: 'Confirm Save',
    deleteKeyTitle: 'Delete Public Key',
    deleteKeyContent: 'Delete the public key ({name})? After deletion, this key can no longer be used to log in.',
    deleteKeyConfirmText: 'Delete',
    unbanTitle: 'Unban IP',
    unbanContent: 'Unban IP {ip}? After unbanning, this IP can try to log in again.',
    unbanConfirmText: 'Unban'
  }
}
