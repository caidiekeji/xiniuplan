// plugins 页面文案
export default {
  title: '插件管理',
  subtitle: '已安装 {total} 个插件 · 已启用 {enabled} 个',
  refresh: '刷新',
  status: {
    running: '运行中',
    enabled: '已启用',
    disabled: '已禁用'
  },
  toggle: {
    enable: '点击启用',
    disable: '点击禁用'
  },
  actions: {
    config: '配置'
  },
  empty: '暂无已安装插件',
  configModal: {
    title: '插件配置',
    titleWithPlugin: '配置插件 - {name}',
    fieldLabel: '插件配置 (JSON)',
    hint: '配置以 JSON 格式存储，保存后若插件正在运行将热更新配置。'
  },
  messages: {
    loadFailed: '加载插件列表失败',
    enabled: '插件「{name}」已启用',
    disabled: '插件「{name}」已禁用',
    actionFailed: '操作失败',
    loadConfigFailed: '加载配置失败',
    jsonInvalid: 'JSON 格式错误，请检查配置内容',
    configSaved: '配置已保存',
    saveConfigFailed: '保存配置失败'
  }
}
