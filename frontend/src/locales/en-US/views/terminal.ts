// terminal page texts
export default {
  title: 'Web Terminal',
  status: {
    connected: 'Connected',
    connecting: 'Connecting',
    disconnected: 'Disconnected'
  },
  rules: {
    hostRequired: 'Please enter the host address',
    portRange: 'Port must be between 1-65535',
    usernameRequired: 'Please enter the username'
  },
  buttons: {
    copy: 'Copy',
    copyTitle: 'Copy (Ctrl+Shift+C)',
    paste: 'Paste',
    pasteTitle: 'Paste (Ctrl+Shift+V)',
    clear: 'Clear',
    clearTitle: 'Clear screen',
    closeTab: 'Close Tab',
    closeTabTitle: 'Close current tab',
    quickCommands: 'Quick Commands',
    connect: 'Connect',
    newTab: 'New terminal tab'
  },
  form: {
    host: 'Host',
    port: 'Port',
    username: 'Username',
    password: 'Password',
    passwordPlaceholder: 'Password'
  },
  quickBarLabel: 'Common Commands',
  messages: {
    noTab: 'Please open a terminal tab first',
    noSelection: 'No text selected',
    copied: 'Copied to clipboard',
    copyFailed: 'Copy failed',
    clipboardDenied: 'Cannot read clipboard, please use Ctrl+Shift+V'
  },
  termOutput: {
    connected: 'CloudPivot Terminal · Connected',
    error: 'Error',
    serverClosed: 'Server closed the connection',
    reconnectFailed: 'Reconnect failed, max attempts reached',
    reconnecting: 'Reconnecting in {seconds}s (attempt {attempt})…'
  }
}
