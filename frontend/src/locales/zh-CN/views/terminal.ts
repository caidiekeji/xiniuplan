// terminal 页面文案
export default {
  title: 'Web 终端',
  status: {
    connected: '已连接',
    connecting: '连接中',
    disconnected: '已断开'
  },
  rules: {
    hostRequired: '请输入主机地址',
    portRange: '端口范围 1-65535',
    usernameRequired: '请输入用户名'
  },
  buttons: {
    copy: '复制',
    copyTitle: '复制 (Ctrl+Shift+C)',
    paste: '粘贴',
    pasteTitle: '粘贴 (Ctrl+Shift+V)',
    clear: '清屏',
    clearTitle: '清屏',
    closeTab: '关闭标签',
    closeTabTitle: '关闭当前标签',
    quickCommands: '快捷命令',
    connect: '连接',
    newTab: '新建终端标签'
  },
  form: {
    host: '主机',
    port: '端口',
    username: '用户名',
    password: '密码',
    passwordPlaceholder: '密码'
  },
  quickBarLabel: '常用命令',
  messages: {
    noTab: '请先打开一个终端标签',
    noSelection: '没有选中的文本',
    copied: '已复制到剪贴板',
    copyFailed: '复制失败',
    clipboardDenied: '无法读取剪贴板，请使用 Ctrl+Shift+V'
  },
  termOutput: {
    connected: '云枢终端 · 已连接',
    error: '错误',
    serverClosed: '服务端已关闭连接',
    reconnectFailed: '重连失败，已达到最大次数',
    reconnecting: '{seconds}s 后第 {attempt} 次重连…'
  }
}
