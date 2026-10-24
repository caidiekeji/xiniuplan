// audit 页面文案
export default {
  resultOptions: {
    all: '全部',
    success: '成功',
    failed: '失败'
  },
  actionOptions: {
    all: '全部操作',
    login: '登录',
    createSite: '创建站点',
    deleteSite: '删除站点',
    createDatabase: '创建数据库',
    installApp: '安装应用',
    restartService: '重启服务',
    fileOperation: '文件操作',
    systemSettings: '系统设置',
    firewallRule: '防火墙规则'
  },
  columns: {
    time: '时间',
    operator: '操作人',
    actionType: '操作类型',
    resource: '资源',
    result: '结果',
    username: '用户名',
    subject: '对象',
    count: '次数',
    lastSeen: '最近发生',
    detail: '详情',
    type: '类型'
  },
  messages: {
    loadAuditFailed: '加载操作日志失败',
    loadLoginFailed: '加载登录日志失败',
    exportSuccess: '日志导出成功',
    exportMock: '日志导出任务已创建（mock）',
    detectFailed: '检测失败'
  },
  anomalyTypes: {
    bruteForceIp: '暴力破解',
    bruteForceUser: '撞库攻击',
    newIpLogin: '新设备登录'
  },
  tabs: {
    operations: '操作日志',
    anomalies: '异常检测',
    logins: '登录日志'
  },
  audit: {
    title: '操作日志',
    subtitle: '所有面板操作的完整审计记录',
    userId: '用户ID',
    userIdPlaceholder: '用户ID',
    actionType: '操作类型',
    result: '结果',
    timeRange: '时间范围'
  },
  anomalies: {
    title: '异常登录检测',
    subtitle: '最近 24 小时：暴力破解 / 撞库 / 新设备登录（S-012）',
    refresh: '重新检测'
  },
  logins: {
    title: '登录日志',
    subtitle: '面板登录记录与异常登录追踪',
    exportCsv: '导出 CSV'
  }
}
