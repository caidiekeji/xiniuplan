// securityExtras 页面文案
export default {
  score: {
    title: '安全评分',
    subtitle: '综合评估密码、防火墙、SSH、SSL 等安全状态'
  },
  refresh: '刷新',
  updates: {
    title: '系统更新',
    subtitle: '检测并安装操作系统安全更新（apt / yum / dnf）',
    check: '检查更新',
    installAll: '一键安装（{count}）',
    installConfirm: '确定安装全部系统更新？过程可能持续数分钟。',
    emptyHint: '点击「检查更新」检测可升级的系统软件包',
    found: '发现 {count} 个可更新软件包',
    checkFailed: '检查更新失败（非 Linux 环境或缺少包管理器）',
    installDone: '系统更新完成',
    installFailed: '安装失败'
  },
  ntp: {
    title: '时间同步 (NTP)',
    subtitle: '系统时间自动同步（SY-006）',
    sync: 'NTP 同步',
    clockSynced: '时钟已同步',
    yes: '是',
    no: '否',
    timezone: '时区',
    localTime: '本地时间',
    enabled: '已开启 NTP 同步',
    disabled: '已关闭 NTP 同步',
    opFailed: '操作失败'
  }
}
