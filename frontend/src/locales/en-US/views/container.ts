// container page messages
export default {
  title: 'Containers',
  subtitle: 'Docker container lifecycle · Live logs · Image browsing',
  refresh: 'Refresh',
  searchPlaceholder: 'Search container name / image / ID',
  showImagesLabel: 'Show images',
  imagesTitle: 'Images',
  pull: {
    button: 'Pull Image',
    title: 'Pull Image',
    imageLabel: 'Image',
    imagePlaceholder: 'e.g. nginx:1.25',
    imageHint: 'Supports official images with mirror rewrite and full private registry paths',
    registryLabel: 'Registry',
    registryPlaceholder: 'Select registry',
    usernamePlaceholder: 'Private registry username',
    passwordPlaceholder: 'Private registry password',
    start: 'Pull',
    cancel: 'Cancel Pull',
    close: 'Close',
    pulling: 'Pulling…',
    complete: 'Image pulled successfully',
    cancelled: 'Pull cancelled',
    imageRequired: 'Image name is required',
    moreLayers: '{n} more layers…'
  },
  unknownError: 'Unknown error',
  execTitle: 'Container Terminal - {name}',
  execDisconnected: '[Disconnected]',
  status: {
    running: 'Running',
    stopped: 'Stopped',
    exited: 'Exited',
    created: 'Created',
    paused: 'Paused',
    restarting: 'Restarting'
  },
  filter: {
    all: 'All'
  },
  columns: {
    name: 'Name',
    image: 'Image',
    status: 'Status',
    ports: 'Port Mapping',
    createdAt: 'Created At',
    actions: 'Actions'
  },
  imageColumns: {
    image: 'Image',
    id: 'ID',
    size: 'Size',
    createdAt: 'Created At'
  },
  actions: {
    start: 'Start',
    stop: 'Stop',
    restart: 'Restart',
    logs: 'Logs',
    terminal: 'Terminal',
    stats: 'Stats',
    delete: 'Delete',
    more: 'More'
  },
  empty: {
    noMatch: 'No matching containers',
    noContainer: 'No containers yet',
    noImages: 'No images yet'
  },
  messages: {
    dockerUnavailable: 'Docker service is unavailable. Please check whether Docker is running',
    started: 'Container {name} started',
    startFailed: 'Start failed',
    stopped: 'Container {name} stopped',
    stopFailed: 'Stop failed',
    restarted: 'Container {name} restarted',
    restartFailed: 'Restart failed',
    deleted: 'Container {name} deleted',
    deleteFailed: 'Delete failed',
    statsFailed: 'Failed to get container resource stats',
    copySuccess: 'Logs copied to clipboard',
    copyFailed: 'Copy failed. Please select the text and copy manually',
    loadImagesFailed: 'Failed to load images',
    imageDeleted: 'Image {name} deleted',
    imageDeleteFailed: 'Failed to delete image'
  },
  logs: {
    title: 'Container Logs - {name}',
    titlePlain: 'Container Logs',
    tail: 'Tail lines',
    tailLines: '{n} lines',
    follow: 'Follow',
    copy: 'Copy',
    empty: '(no log output)',
    loadFailed: 'Failed to load logs: {msg}'
  },
  deleteDialog: {
    title: 'Delete Container',
    content: 'Delete container "{name}"? This cannot be undone and unpersisted data in the container will be lost.',
    confirm: 'Delete'
  },
  imageDeleteDialog: {
    title: 'Delete Image',
    content: 'Delete image "{name}"? This cannot be undone. If any container is using this image, the deletion will fail.',
    confirm: 'Delete'
  },
  stats: {
    title: 'Resource Stats - {name}',
    cpuUsage: 'CPU Usage',
    memoryUsage: 'Memory Usage',
    memoryValue: '{usage} / {limit} ({percent}%)'
  }
}
