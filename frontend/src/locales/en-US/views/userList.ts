// userList page texts
export default {
  title: 'User List',
  subtitle: 'Manage system users, role assignments and account status',
  createUser: 'Create User',
  editUser: 'Edit User',
  searchPlaceholder: 'Search username...',
  roles: {
    superAdmin: 'Super Admin',
    admin: 'Admin',
    user: 'User'
  },
  columns: {
    username: 'Username',
    role: 'Role',
    status: 'Status',
    lastLogin: 'Last Login',
    createdAt: 'Created At',
    actions: 'Actions'
  },
  status: {
    enabled: 'Enabled',
    disabled: 'Not enabled',
    active: 'Active',
    inactive: 'Disabled',
    on: 'On',
    off: 'Off'
  },
  actions: {
    edit: 'Edit',
    password: 'Password',
    enable: 'Enable',
    disable: 'Disable',
    delete: 'Delete'
  },
  labels: {
    username: 'Username',
    password: 'Password',
    passwordOptional: 'Password (leave blank to keep unchanged)',
    confirmPassword: 'Confirm Password',
    role: 'Role',
    status: 'Status',
    newPassword: 'New Password',
    confirmNewPassword: 'Confirm New Password'
  },
  hints: {
    enterUsername: 'Please enter a username',
    enterPassword: 'Please enter a password',
    reenterPassword: 'Please enter the password again',
    confirmPassword: 'Please confirm the password',
    selectRole: 'Please select a role',
    enterNewPassword: 'Please enter a new password',
    confirmNewPassword: 'Please confirm the new password',
    reenterNewPassword: 'Please enter the new password again',
    passwordMismatch: 'The two passwords do not match'
  },
  strength: {
    label: 'Password strength: {level}',
    weak: 'Weak',
    medium: 'Medium',
    strong: 'Strong'
  },
  loadError: 'Failed to load users',
  enableSuccess: 'User enabled',
  enableFailed: 'Failed to enable',
  disableSuccess: 'User disabled',
  disableFailed: 'Failed to disable',
  messages: {
    saveFailed: 'Save failed',
    updated: 'User updated',
    created: 'User created',
    passwordChanged: 'Password changed',
    passwordFailed: 'Failed to change password',
    deleted: 'User deleted',
    deleteFailed: 'Delete failed'
  },
  pwdTitle: 'Change Password — {username}',
  delete: {
    title: 'Delete User',
    content: 'Delete user "{name}"? This action cannot be undone.'
  }
}
