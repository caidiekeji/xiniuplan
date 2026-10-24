// plugins page texts
export default {
  title: 'Plugins',
  subtitle: '{total} installed · {enabled} enabled',
  refresh: 'Refresh',
  status: {
    running: 'Running',
    enabled: 'Enabled',
    disabled: 'Disabled'
  },
  toggle: {
    enable: 'Click to enable',
    disable: 'Click to disable'
  },
  actions: {
    config: 'Config'
  },
  empty: 'No installed plugins',
  configModal: {
    title: 'Plugin Config',
    titleWithPlugin: 'Config Plugin - {name}',
    fieldLabel: 'Plugin Config (JSON)',
    hint: 'Config is stored as JSON. If the plugin is running, changes take effect after saving.'
  },
  messages: {
    loadFailed: 'Failed to load plugin list',
    enabled: 'Plugin "{name}" enabled',
    disabled: 'Plugin "{name}" disabled',
    actionFailed: 'Operation failed',
    loadConfigFailed: 'Failed to load config',
    jsonInvalid: 'Invalid JSON, please check the config content',
    configSaved: 'Config saved',
    saveConfigFailed: 'Failed to save config'
  }
}
