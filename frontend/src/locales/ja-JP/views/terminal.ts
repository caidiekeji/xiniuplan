// terminal ページ文言
export default {
  title: 'Web ターミナル',
  status: {
    connected: '接続済み',
    connecting: '接続中',
    disconnected: '切断済み'
  },
  rules: {
    hostRequired: 'ホストアドレスを入力してください',
    portRange: 'ポートは 1-65535 の範囲で指定してください',
    usernameRequired: 'ユーザー名を入力してください'
  },
  buttons: {
    copy: 'コピー',
    copyTitle: 'コピー (Ctrl+Shift+C)',
    paste: '貼り付け',
    pasteTitle: '貼り付け (Ctrl+Shift+V)',
    clear: 'クリア',
    clearTitle: '画面をクリア',
    closeTab: 'タブを閉じる',
    closeTabTitle: '現在のタブを閉じる',
    quickCommands: 'クイックコマンド',
    connect: '接続',
    newTab: '新しいターミナルタブ'
  },
  form: {
    host: 'ホスト',
    port: 'ポート',
    username: 'ユーザー名',
    password: 'パスワード',
    passwordPlaceholder: 'パスワード'
  },
  quickBarLabel: 'よく使うコマンド',
  messages: {
    noTab: '先にターミナルタブを開いてください',
    noSelection: '選択されたテキストがありません',
    copied: 'クリップボードにコピーしました',
    copyFailed: 'コピーに失敗しました',
    clipboardDenied: 'クリップボードを読み取れません。Ctrl+Shift+V を使用してください'
  },
  termOutput: {
    connected: 'CloudPivot ターミナル · 接続済み',
    error: 'エラー',
    serverClosed: 'サーバーが接続を閉じました',
    reconnectFailed: '再接続に失敗しました（最大回数に達しました）',
    reconnecting: '{seconds} 秒後に {attempt} 回目の再接続…'
  }
}
