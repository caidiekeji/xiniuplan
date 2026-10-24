// userList 页面文案
export default {
  title: '用户列表',
  subtitle: '管理系统用户、角色分配与账户状态',
  createUser: '创建用户',
  editUser: '编辑用户',
  searchPlaceholder: '搜索用户名...',
  roles: {
    superAdmin: '超级管理员',
    admin: '管理员',
    user: '普通用户'
  },
  columns: {
    username: '用户名',
    role: '角色',
    status: '状态',
    lastLogin: '最后登录',
    createdAt: '创建时间',
    actions: '操作'
  },
  status: {
    enabled: '已启用',
    disabled: '未启用',
    active: '正常',
    inactive: '已禁用',
    on: '启用',
    off: '禁用'
  },
  actions: {
    edit: '编辑',
    password: '改密',
    enable: '启用',
    disable: '禁用',
    delete: '删除'
  },
  labels: {
    username: '用户名',
    password: '密码',
    passwordOptional: '密码（留空不修改）',
    confirmPassword: '确认密码',
    role: '角色',
    status: '状态',
    newPassword: '新密码',
    confirmNewPassword: '确认新密码'
  },
  hints: {
    enterUsername: '请输入用户名',
    enterPassword: '请输入密码',
    reenterPassword: '请再次输入密码',
    confirmPassword: '请确认密码',
    selectRole: '请选择角色',
    enterNewPassword: '请输入新密码',
    confirmNewPassword: '请确认新密码',
    reenterNewPassword: '请再次输入新密码',
    passwordMismatch: '两次输入的密码不一致'
  },
  strength: {
    label: '密码强度：{level}',
    weak: '弱',
    medium: '中',
    strong: '强'
  },
  loadError: '加载用户列表失败',
  enableSuccess: '用户已启用',
  enableFailed: '启用失败',
  disableSuccess: '用户已禁用',
  disableFailed: '禁用失败',
  messages: {
    saveFailed: '保存失败',
    updated: '用户已更新',
    created: '用户已创建',
    passwordChanged: '密码已修改',
    passwordFailed: '修改密码失败',
    deleted: '用户已删除',
    deleteFailed: '删除失败'
  },
  pwdTitle: '修改密码 — {username}',
  delete: {
    title: '删除用户',
    content: '确定要删除用户「{name}」吗？此操作不可撤销。'
  }
}
