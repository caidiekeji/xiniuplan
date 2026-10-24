// backup page texts
export default {
  title: 'Backup & Restore',
  subtitle: 'Website, database and system data backup with one-click restore',
  createBackup: 'Create Backup',
  remoteStorage: 'Remote Storage',
  types: {
    site: 'Website',
    database: 'Database',
    system: 'System',
    app: 'Application'
  },
  status: {
    pending: 'Pending',
    completed: 'Completed',
    running: 'In Progress',
    failed: 'Failed'
  },
  unknown: 'Unknown',
  search: {
    type: 'Type',
    allTypes: 'All types',
    keyword: 'Search',
    keywordPlaceholder: 'Search by target or remark'
  },
  columns: {
    type: 'Backup Type',
    target: 'Target',
    size: 'File Size',
    status: 'Status',
    createdAt: 'Created At',
    actions: 'Actions'
  },
  actions: {
    restore: 'Restore',
    download: 'Download',
    push: 'Push',
    delete: 'Delete'
  },
  create: {
    title: 'Create Backup',
    typeLabel: 'Backup Type',
    siteTarget: 'Select Site',
    sitePlaceholder: 'Select the site to back up',
    dbTarget: 'Select Database',
    dbPlaceholder: 'Select the database instance to back up',
    systemLabel: 'System Backup',
    systemHint: 'Performs a full backup of the entire system, including configuration, application data and system files',
    remark: 'Remark',
    remarkPlaceholder: 'Optional remark'
  },
  rules: {
    targetRequired: 'Please select or enter a backup target'
  },
  messages: {
    loadFailed: 'Failed to load backup list',
    enterTarget: 'Please enter a backup target',
    createSuccess: 'Backup task created and now in progress...',
    createFailed: 'Failed to create backup',
    restoreSuccess: 'Restore started; a snapshot was created automatically beforehand',
    restoreFailed: 'Restore failed',
    downloadStart: 'Download started',
    downloadFailed: 'Download failed',
    deleteSuccess: 'Backup deleted',
    deleteFailed: 'Failed to delete backup',
    loadRemoteFailed: 'Failed to load remote storages',
    selectRemote: 'Please select a remote storage',
    pushSuccess: 'Push completed',
    pushFailed: 'Push failed',
    enterName: 'Please enter a name',
    remoteCreateSuccess: 'Remote storage created',
    remoteCreateFailed: 'Failed to create',
    testSuccess: 'Connection OK',
    testFailed: 'Connection failed'
  },
  restore: {
    title: 'Restore Backup',
    content: 'Restore the backup for "{name}"? This will overwrite the current data; a snapshot will be created automatically before restoring.',
    confirm: 'Restore'
  },
  remove: {
    title: 'Delete Backup',
    content: 'Delete the backup for "{name}"? This action cannot be undone.',
    confirm: 'Delete'
  },
  push: {
    title: 'Push Backup to Remote Storage',
    placeholder: 'Select remote storage',
    cancel: 'Cancel',
    confirm: 'Push'
  },
  remote: {
    title: 'Remote Backup Storage (S3 / WebDAV / SFTP / Panel)',
    enabled: 'Enabled',
    disabled: 'Disabled',
    test: 'Test',
    delete: 'Delete',
    close: 'Close',
    create: 'New'
  },
  remoteCreate: {
    title: 'New Remote Storage',
    name: 'Name',
    type: 'Type',
    types: {
      s3: 'S3-Compatible Storage',
      webdav: 'WebDAV',
      sftp: 'SFTP',
      panel: 'Another Panel'
    },
    address: 'Address / Host',
    username: 'Username',
    password: 'Password',
    cancel: 'Cancel',
    create: 'Create'
  }
}
