// appstore page texts
export default {
  pageTitle: 'App Store',
  pageSubtitle: 'One-click deploy for common web apps · Containerized runtime · Visual env config',
  tabs: {
    store: 'App Store',
    installed: 'Installed',
    sources: 'Sources'
  },
  addSource: 'Add Source',
  loadSourcesFailed: 'Failed to load source list',
  sourceTypes: {
    official: 'Official',
    community: 'Community',
    private: 'Private'
  },
  form: {
    addTitle: 'Add Source',
    editTitle: 'Edit Source',
    nameLabel: 'Source Name',
    namePlaceholder: 'Enter source name, e.g. Official App Source',
    typeLabel: 'Type',
    typePlaceholder: 'Select source type',
    pathLabel: 'Local Path',
    enabledLabel: 'Enabled'
  },
  validate: {
    nameRequired: 'Please enter the source name',
    communityUrlRequired: 'Please enter the community source URL',
    privatePathRequired: 'Please enter the local path of the private source',
    officialOnlyOnce: 'Only one official source is allowed. Modify or delete the existing one first'
  },
  messages: {
    sourceUpdated: 'Source updated',
    sourceAdded: 'Source added',
    actionFailed: 'Operation failed',
    sourceDisabled: 'Disabled "{name}"',
    sourceEnabled: 'Enabled "{name}"',
    syncDone: 'Sync finished: {added} added / {updated} updated / {deleted} deleted',
    syncFailed: 'Sync failed',
    officialNoDelete: 'The official source cannot be deleted',
    sourceDeleted: '"{name}" deleted',
    deleteFailed: 'Deletion failed'
  },
  columns: {
    name: 'Source Name',
    type: 'Type',
    url: 'URL / Path',
    status: 'Status',
    lastSync: 'Last Sync',
    actions: 'Actions'
  },
  status: {
    enabled: 'Enabled',
    disabled: 'Disabled'
  },
  actions: {
    sync: 'Sync',
    enable: 'Enable',
    disable: 'Disable',
    edit: 'Edit',
    delete: 'Delete'
  },
  deleteDialog: {
    title: 'Delete Source',
    content: 'Delete source "{name}"? Apps under this source will become unavailable. This action cannot be undone.',
    confirm: 'Delete'
  },
  categories: {
    website: 'Websites',
    devtools: 'Dev Tools',
    database: 'Databases',
    middleware: 'Middleware',
    monitor: 'Monitoring',
    office: 'Office',
    storage: 'Storage',
    network: 'Network'
  },
  statuses: {
    installed: 'Installed',
    installing: 'Installing',
    failed: 'Install Failed',
    not_installed: 'Not Installed'
  },
  storeTab: {
    allCategories: 'All',
    allSources: 'All Sources',
    officialGroup: 'Official Sources',
    communityGroup: 'Community Sources',
    privateGroup: 'Private Sources',
    loadFailed: 'Failed to load app list',
    sourcePlaceholder: 'Source',
    searchPlaceholder: 'Search apps...',
    memory: 'Memory {value}',
    disk: 'Disk {value}',
    install: 'Install',
    noMatch: 'No matching apps',
    noApps: 'No apps available yet'
  },
  installedTab: {
    loadFailed: 'Failed to load installed apps',
    noAccessUrl: 'This app has no access URL yet',
    noContainer: 'This app has no linked container, so logs are unavailable',
    noLogs: '(No logs yet)',
    loadLogsFailed: 'Failed to load logs: {error}',
    upgraded: '"{name}" upgraded to {version}',
    upgradeFailed: 'Upgrade failed',
    uninstalled: '{name} uninstalled',
    uninstallFailed: 'Uninstall failed',
    refresh: 'Refresh',
    empty: 'No installed apps yet. Browse the App Store',
    browseApps: 'Browse Apps',
    logTitle: 'App Logs - {name}',
    uninstallTitle: 'Uninstall App',
    uninstallContent: 'Uninstall "{name}"? The container and related data will be deleted. This action cannot be undone.',
    uninstallConfirm: 'Uninstall',
    columns: {
      appName: 'App Name',
      version: 'Version',
      category: 'Category',
      installedAt: 'Installed At'
    },
    actions: {
      open: 'Open',
      logs: 'Logs',
      upgrade: 'Upgrade',
      uninstall: 'Uninstall'
    }
  },
  installDialog: {
    title: 'Install {name}',
    titleNoApp: 'Install App',
    confirmTitle: 'Confirm',
    confirmDesc: 'Review app details',
    configTitle: 'Configure',
    configDesc: 'Set environment variables',
    installingTitle: 'Installing',
    installingDesc: 'Waiting for deployment to finish',
    version: 'Version {version}',
    minResources: 'Minimum resource requirements',
    memory: 'Memory: {value}',
    disk: 'Disk: {value}',
    cpu: 'CPU: {value}',
    exposedPorts: 'Exposed ports:',
    requiredWarning: 'Please fill in the required field: {label}',
    inputPlaceholder: 'Enter {label}',
    selectPlaceholder: 'Select {label}',
    noEnv: 'This app needs no environment parameters and can be installed right away',
    logsTitle: 'Install Log',
    logPreparing: 'Preparing install environment...',
    logPulling: 'Pulling image {name}:{version}...',
    logConfiguring: 'Configuring container and network...',
    logDone: 'Install complete!',
    logError: 'Error: {error}',
    unknownError: 'Unknown error',
    successTitle: 'Installed Successfully',
    successMessage: '{name} installed successfully',
    installFailed: 'Install failed',
    accessUrl: 'Access URL:',
    initialPassword: 'Initial password:',
    containerId: 'Container ID:',
    copy: 'Copy',
    passwordCopied: 'Password copied to clipboard',
    copyFailed: 'Copy failed, please copy it manually',
    done: 'Done',
    cancel: 'Cancel',
    next: 'Next',
    prev: 'Back',
    start: 'Start Install',
    viewInstalled: 'View Installed'
  }
}
