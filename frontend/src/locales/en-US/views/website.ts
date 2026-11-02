// website page texts
export default {
  title: 'Websites',
  subtitle: 'Full lifecycle management of web sites',
  createSite: 'Create Site',
  search: {
    domainLabel: 'Domain',
    domainPlaceholder: 'Search by domain',
    typeLabel: 'Type',
    typePlaceholder: 'All types',
    statusLabel: 'Status',
    statusPlaceholder: 'All statuses'
  },
  types: {
    static: 'Static Site',
    runtime: 'Runtime',
    reverseProxy: 'Reverse Proxy'
  },
  status: {
    running: 'Running',
    stopped: 'Stopped',
    error: 'Error'
  },
  columns: {
    domain: 'Domain',
    type: 'Type',
    runtime: 'Runtime',
    status: 'Status',
    createdAt: 'Created At',
    actions: 'Actions'
  },
  ssl: {
    enabled: 'Enabled',
    disabled: 'Disabled'
  },
  actions: {
    start: 'Start',
    stop: 'Stop',
    restart: 'Restart',
    config: 'Config',
    logs: 'Logs',
    edit: 'Edit',
    delete: 'Delete'
  },
  messages: {
    loadListFailed: 'Failed to load site list',
    startSuccess: 'Site started',
    startFailed: 'Start failed',
    stopSuccess: 'Site stopped',
    stopFailed: 'Stop failed',
    restartSuccess: 'Site restarted',
    restartFailed: 'Restart failed',
    createSuccess: 'Site created successfully',
    createFailed: 'Creation failed',
    updateSuccess: 'Site updated',
    updateFailed: 'Update failed',
    readConfigFailed: 'Failed to read Nginx config',
    configSaved: 'Nginx config saved',
    saveFailed: 'Save failed',
    logsFailed: 'Failed to fetch site logs',
    deleteSuccess: 'Site deleted',
    deleteFailed: 'Delete failed'
  },
  rules: {
    domainRequired: 'Please enter the primary domain',
    domainInvalid: 'Invalid domain format (e.g. example.com or *.example.com)',
    aliasesInvalid: 'Some alias domains have an invalid format',
    rootDirRequired: 'Please enter or select the site root directory',
    runtimeRequired: 'Please select a runtime',
    proxyTargetRequired: 'Please enter the reverse proxy target'
  },
  createModal: {
    title: 'Create Site',
    typeLabel: 'Site Type',
    reverseProxyDesc: 'Forward to backend service',
    domainLabel: 'Primary Domain',
    domainPlaceholder: 'e.g. example.com',
    aliasesLabel: 'Alias Domains',
    aliasesPlaceholder: 'Type an alias and press Enter to add',
    runtimeLabel: 'Runtime',
    runtimePlaceholder: 'Select runtime',
    proxyTargetLabel: 'Reverse Proxy Target',
    proxyTargetPlaceholder: 'e.g. http://127.0.0.1:3000',
    rootDirLabel: 'Site Root Directory',
    rootDirBrowse: 'Browse',
    rootDirHint: 'Enter an absolute path manually, or click "Browse" to pick a directory on the server',
    remarkLabel: 'Remark',
    remarkPlaceholder: 'Optional remark'
  },
  editModal: {
    title: 'Edit Site',
    domainLocked: 'Primary domain cannot be changed after creation; delete and recreate to change it'
  },
  configDrawer: {
    title: 'Nginx Config — {domain}',
    loading: 'Loading...',
    cancel: 'Cancel',
    save: 'Save Config'
  },
  logsModal: {
    title: 'Site Logs — {domain}',
    refresh: 'Refresh'
  },
  logTypes: {
    access: 'Access Log (access)',
    error: 'Error Log (error)'
  },
  deleteDialog: {
    title: 'Delete Site',
    content: 'Are you sure you want to delete site "{domain}"? This action cannot be undone.',
    deleteFiles: 'Also delete files in the site root directory'
  }
}
