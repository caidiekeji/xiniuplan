// compose page messages
export default {
  cardTitle: 'Compose Stacks',
  buildCardTitle: 'Image Build',
  refresh: 'Refresh',
  createStack: 'New Stack',
  columns: {
    name: 'Stack Name',
    containers: 'Containers',
    status: 'Status',
    actions: 'Actions'
  },
  status: {
    running: 'Running',
    stopped: 'Not Running',
    partial: 'Partially Running'
  },
  actions: {
    start: 'Start',
    stop: 'Stop',
    restart: 'Restart',
    edit: 'Edit',
    delete: 'Delete'
  },
  empty: {
    noStack: 'No Compose stacks yet',
    noStackHint: 'Click "New Stack" and paste your docker-compose.yml content to manage it here',
    noBuildLog: 'No build logs yet',
    noBuildLogHint: 'Fill in the tag and Dockerfile, then click "Start Build" to see logs in real time'
  },
  build: {
    tagLabel: 'Image tag',
    start: 'Start Build',
    abort: 'Abort'
  },
  editor: {
    createTitle: 'New Stack',
    editTitle: 'Edit Stack · {name}',
    nameLabel: 'Stack Name',
    namePlaceholder: 'e.g. blog',
    cancel: 'Cancel',
    save: 'Save'
  },
  deleteDialog: {
    title: 'Delete Stack',
    content: 'Delete stack "{name}"?',
    confirm: 'Delete'
  },
  rules: {
    nameRequired: 'Please enter a stack name',
    tagRequired: 'Please enter an image tag',
    dockerfileRequired: 'Please enter Dockerfile content'
  },
  messages: {
    actionDone: 'Operation completed',
    actionFailed: 'Operation failed',
    loadContentFailed: 'Failed to load stack content',
    saved: 'Saved',
    saveFailed: 'Save failed',
    deleted: 'Deleted',
    deleteFailed: 'Delete failed',
    buildFailed: 'Build failed',
    buildAborted: '[ABORT] Build aborted',
    buildError: 'Build reported errors',
    buildDone: 'Build finished'
  }
}
