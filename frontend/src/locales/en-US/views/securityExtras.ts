// securityExtras page texts
export default {
  score: {
    title: 'Security Score',
    subtitle: 'Overall assessment of password, firewall, SSH, SSL and other security status'
  },
  refresh: 'Refresh',
  updates: {
    title: 'System Updates',
    subtitle: 'Detect and install OS security updates (apt / yum / dnf)',
    check: 'Check for Updates',
    installAll: 'Install All ({count})',
    installConfirm: 'Install all system updates? This may take several minutes.',
    emptyHint: 'Click "Check for Updates" to detect upgradable system packages',
    found: '{count} updatable packages found',
    checkFailed: 'Failed to check for updates (non-Linux environment or no package manager)',
    installDone: 'System updates completed',
    installFailed: 'Installation failed'
  },
  ntp: {
    title: 'Time Sync (NTP)',
    subtitle: 'Automatic system time sync (SY-006)',
    sync: 'NTP Sync',
    clockSynced: 'Clock Synced',
    yes: 'Yes',
    no: 'No',
    timezone: 'Timezone',
    localTime: 'Local Time',
    enabled: 'NTP sync enabled',
    disabled: 'NTP sync disabled',
    opFailed: 'Operation failed'
  }
}
