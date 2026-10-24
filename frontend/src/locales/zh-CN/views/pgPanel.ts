// pgPanel 页面文案
export default {
  tabs: {
    schema: '库表管理',
    users: '用户权限',
    sql: 'SQL 执行器'
  },
  refresh: '刷新',
  execute: '执行',
  delete: '删除',
  createFailed: '创建失败',
  deleteFailed: '删除失败',
  selectDb: '请选择数据库',
  selectDbLabel: '选择数据库',
  selectDbPlaceholder: '选择数据库',
  loadDatabasesFailed: '加载数据库列表失败',
  loadTablesFailed: '加载表列表失败',
  loadSchemaFailed: '加载表结构失败',
  loadTableDataFailed: '加载表数据失败',
  loadUsersFailed: '加载用户列表失败',
  columns: {
    tableName: '表名',
    rowsEstimate: '行数（估算）',
    size: '大小',
    actions: '操作',
    structure: '结构',
    data: '数据',
    columnName: '列名',
    type: '类型',
    nullable: '可空',
    defaultValue: '默认值',
    primaryKey: '主键',
    username: '用户名',
    attributes: '属性',
    validUntil: '有效时间',
    permanent: '永久',
    grant: '授权',
    revoke: '回收权限'
  },
  sidebarTitle: '数据库',
  noDatabases: '暂无数据库',
  selectDbLeft: '请选择左侧数据库',
  noTables: '该数据库暂无表',
  tableSchema: '表结构: {name}',
  tableData: '表数据: {name}',
  indexes: '索引',
  noUsers: '暂无用户，点击上方按钮创建',
  allowDestructive: '允许危险操作',
  selectDbFirst: '请先选择数据库',
  sql: {
    enterSql: '请输入 SQL 语句',
    executed: '执行完成',
    failed: '执行失败',
    success: '执行成功',
    duration: '耗时 {ms} ms',
    affected: '影响 {rows} 行',
    noResultSet: '无返回结果集（非 SELECT 语句）',
    defaultSql: '-- 在此输入 SQL 语句\nSELECT 1;'
  },
  createDb: {
    title: '创建数据库',
    nameLabel: '数据库名',
    namePlaceholder: '例如: myapp_db',
    encodingLabel: '编码',
    collateLabel: '排序规则',
    ownerPlaceholder: '留空使用当前用户',
    nameRequired: '请输入数据库名',
    success: '数据库创建成功'
  },
  dbDeleted: '数据库 {name} 已删除',
  createUser: {
    title: '创建数据库用户',
    create: '创建用户',
    nameLabel: '用户名',
    namePlaceholder: '例如: app_user',
    passwordLabel: '密码',
    passwordPlaceholder: '请输入密码',
    roleAttributesLabel: '角色属性',
    nameRequired: '请输入用户名',
    passwordRequired: '请输入密码',
    success: '用户创建成功'
  },
  userDeleted: '用户 {name} 已删除',
  grant: {
    title: '授权用户: {name}',
    privilegesLabel: '权限',
    success: '授权成功',
    failed: '授权失败'
  },
  revoke: {
    title: '回收权限: {name}',
    dbPlaceholder: '选择要回收权限的数据库',
    success: '已回收 {user} 在 {db} 上的权限',
    failed: '回收权限失败'
  },
  deleteDb: {
    title: '删除数据库',
    content: '确定要删除数据库「{name}」吗？此操作不可恢复。'
  },
  deleteUser: {
    title: '删除用户',
    content: '确定要删除用户「{name}」吗？'
  }
}
