// securityExtras ページ文言
export default {
  score: {
    title: 'セキュリティスコア',
    subtitle: 'パスワード、ファイアウォール、SSH、SSL などのセキュリティ状態を総合評価します'
  },
  refresh: '更新',
  updates: {
    title: 'システム更新',
    subtitle: 'OS のセキュリティ更新を検出してインストールします（apt / yum / dnf）',
    check: '更新を確認',
    installAll: 'すべてインストール（{count}）',
    installConfirm: 'すべてのシステム更新をインストールしますか？数分かかる場合があります。',
    emptyHint: '「更新を確認」をクリックしてアップグレード可能なシステムパッケージを検出します',
    found: '{count} 個の更新可能なパッケージが見つかりました',
    checkFailed: '更新の確認に失敗しました（Linux 環境以外、またはパッケージマネージャがありません）',
    installDone: 'システム更新が完了しました',
    installFailed: 'インストールに失敗しました'
  },
  ntp: {
    title: '時間同期 (NTP)',
    subtitle: 'システム時刻の自動同期（SY-006）',
    sync: 'NTP 同期',
    clockSynced: '時計は同期済み',
    yes: 'はい',
    no: 'いいえ',
    timezone: 'タイムゾーン',
    localTime: 'ローカル時刻',
    enabled: 'NTP 同期を有効にしました',
    disabled: 'NTP 同期を無効にしました',
    opFailed: '操作に失敗しました'
  }
}
