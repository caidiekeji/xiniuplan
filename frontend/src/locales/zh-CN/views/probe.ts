// probe 页面文案
export default {
  pageTitle: '网站拨测',
  pageSubtitle: 'HTTP(S)/TCP 可用性探测，故障自动告警',
  refresh: '刷新',
  addProbe: '添加探针',
  status: {
    up: '正常',
    down: '故障',
    unknown: '未检测'
  },
  switchStatus: {
    enabled: '启用',
    disabled: '停用'
  },
  columns: {
    name: '名称',
    type: '类型',
    target: '目标',
    interval: '间隔',
    status: '状态',
    latency: '延迟',
    httpCode: 'HTTP 码',
    lastChecked: '最后检测',
    enabled: '启用',
    actions: '操作',
    checkedAt: '检测时间',
    result: '结果',
    message: '说明'
  },
  actions: {
    checkNow: '立即检测',
    edit: '编辑',
    delete: '删除'
  },
  messages: {
    checkDone: '检测完成',
    checkFailed: '检测失败',
    updated: '探针已更新',
    created: '探针已创建',
    saveFailed: '保存失败',
    deleted: '探针已删除',
    deleteFailed: '删除失败'
  },
  rules: {
    nameRequired: '请输入探针名称',
    targetRequired: '请输入检测目标'
  },
  empty: {
    title: '暂无拨测探针',
    desc: '添加探针后将持续检测目标可用性，故障时通过「告警通知」渠道推送'
  },
  historyButton: '查看首个探针的检测历史',
  form: {
    addTitle: '添加探针',
    editTitle: '编辑探针',
    nameLabel: '探针名称',
    namePlaceholder: '例如：官网首页',
    typeLabel: '类型',
    intervalLabel: '检测间隔（秒）',
    targetLabel: '检测目标',
    keywordLabel: '响应关键词（可选，响应体需包含）',
    keywordPlaceholder: '例如：ok',
    enabledLabel: '启用'
  },
  results: {
    title: '检测历史 · {name}'
  },
  deleteDialog: {
    title: '删除探针',
    content: '确定要删除探针「{name}」吗？其检测历史将一并移除。',
    confirm: '删除'
  }
}
