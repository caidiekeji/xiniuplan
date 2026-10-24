// roleList 页面文案
export default {
  title: '角色权限',
  subtitle: '管理系统角色及其权限分配',
  createRole: '创建角色',
  editRole: '编辑角色',
  columns: {
    name: '角色名称',
    code: '角色编码',
    description: '描述',
    permissions: '权限数量',
    type: '类型',
    actions: '操作'
  },
  status: {
    builtIn: '内置',
    custom: '自定义'
  },
  actions: {
    edit: '编辑',
    delete: '删除'
  },
  labels: {
    name: '角色名称',
    code: '角色编码',
    description: '描述',
    permissions: '权限分配'
  },
  hints: {
    name: '例如：运维工程师',
    code: '例如：ops_engineer',
    description: '可选描述信息',
    nameRequired: '请输入角色名称',
    codeRequired: '请输入角色编码'
  },
  matrix: {
    selectAll: '全选',
    selected: '已选 {selected} / {total}'
  },
  permissions: {
    modules: {
      site: '网站',
      database: '数据库',
      file: '文件',
      terminal: '终端',
      container: '容器',
      appstore: '应用商店',
      firewall: '防火墙',
      backup: '备份',
      monitor: '监控',
      user: '用户',
      audit: '审计',
      setting: '设置',
      alert: '告警',
      cron: '计划任务'
    },
    actions: {
      view: '查看',
      create: '创建',
      update: '修改',
      delete: '删除'
    }
  },
  messages: {
    loadError: '加载角色列表失败',
    updated: '角色已更新',
    created: '角色已创建',
    saveFailed: '保存失败',
    deleted: '角色已删除',
    deleteFailed: '删除失败'
  },
  delete: {
    title: '删除角色',
    content: '确定要删除角色「{name}」吗？此操作不可撤销。'
  }
}
