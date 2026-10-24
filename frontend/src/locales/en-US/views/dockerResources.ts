// dockerResources page messages
export default {
  title: 'Networks & Volumes',
  subtitle: 'Docker networks and volumes management (C-008 / C-009)',
  refresh: 'Refresh',
  tabs: {
    networks: 'Networks',
    volumes: 'Volumes'
  },
  createNetwork: 'Create Network',
  createVolume: 'Create Volume',
  dialogs: {
    createNetworkTitle: 'Create Docker Network',
    createVolumeTitle: 'Create Docker Volume'
  },
  form: {
    nameLabel: 'Name',
    namePlaceholderNetwork: 'e.g. app-net',
    namePlaceholderVolume: 'e.g. app-data',
    driverLabel: 'Driver',
    subnetLabel: 'Subnet',
    subnetPlaceholder: 'Optional, e.g. 172.30.0.0/16',
    gatewayLabel: 'Gateway',
    gatewayPlaceholder: 'Optional',
    cancel: 'Cancel',
    confirm: 'Create'
  },
  columns: {
    name: 'Name',
    driver: 'Driver',
    subnet: 'Subnet',
    containers: 'Containers',
    inUse: 'In Use',
    usedBy: 'Used By',
    actions: 'Actions'
  },
  yes: 'Yes',
  no: 'No',
  delete: 'Delete',
  deleteDialog: {
    cancel: 'Cancel'
  },
  deleteNetwork: {
    title: 'Delete Network',
    content: 'Delete network "{name}"?'
  },
  deleteVolume: {
    title: 'Delete Volume',
    content: 'Delete volume "{name}"? The data cannot be recovered.'
  },
  messages: {
    loadNetworksFailed: 'Failed to load networks',
    loadVolumesFailed: 'Failed to load volumes',
    networkCreated: 'Network created',
    volumeCreated: 'Volume created',
    createFailed: 'Creation failed',
    deleted: 'Deleted',
    volumeInUse: 'The volume is in use by a container and cannot be deleted'
  }
}
