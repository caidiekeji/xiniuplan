// daemon page texts
export default {
  pageTitle: 'Process Guardian',
  pageSubtitle: 'Auto-restart crashed managed processes, restore them after panel restart',
  refresh: 'Refresh',
  addDaemon: 'Add Guardian',
  columns: {
    name: 'Name',
    command: 'Command',
    workDir: 'Working Directory',
    status: 'Status',
    restarts: 'Restarts',
    lastExitCode: 'Last Exit Code',
    autoStart: 'Auto Start',
    actions: 'Actions'
  },
  status: {
    running: 'Running',
    stopped: 'Stopped'
  },
  autoStart: {
    on: 'On',
    off: 'Off'
  },
  actions: {
    start: 'Start',
    stop: 'Stop',
    logs: 'Logs',
    edit: 'Edit',
    delete: 'Delete'
  },
  messages: {
    started: 'Process started',
    startFailed: 'Start failed',
    stopped: 'Process stopped',
    stopFailed: 'Stop failed',
    updated: 'Guardian config updated',
    created: 'Guardian process created',
    saveFailed: 'Save failed',
    deleted: 'Guardian process deleted',
    deleteFailed: 'Deletion failed'
  },
  rules: {
    nameRequired: 'Please enter the process name',
    commandRequired: 'Please enter the start command'
  },
  empty: {
    title: 'No guardian processes yet',
    desc: 'Once added, abnormal exits are restarted automatically; enable Auto Start to restore with the panel'
  },
  form: {
    addTitle: 'Add Guardian Process',
    editTitle: 'Edit Guardian Process',
    nameLabel: 'Process Name',
    namePlaceholder: 'e.g. node-api',
    commandLabel: 'Start Command',
    commandPlaceholder: "e.g. node /opt/app/server.js or bash -c 'cd /opt/app && ./start.sh'",
    commandHint: 'Any shell command is supported; wrap complex start logic with bash -c',
    workDirLabel: 'Working Directory',
    workDirPlaceholder: 'e.g. /opt/app (optional)',
    autoStartLabel: 'Restart on crash / restore on panel restart',
    enabledLabel: 'Enabled'
  },
  logs: {
    title: 'Process Logs · {name}',
    loading: 'Loading…',
    empty: 'No log output yet'
  },
  deleteDialog: {
    title: 'Delete Guardian Process',
    content: 'Delete guardian process "{name}"? Running processes will be terminated.',
    confirm: 'Delete'
  }
}
