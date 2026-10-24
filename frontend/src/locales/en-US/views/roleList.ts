// roleList page texts
export default {
  title: 'Roles & Permissions',
  subtitle: 'Manage system roles and their permission assignments',
  createRole: 'Create Role',
  editRole: 'Edit Role',
  columns: {
    name: 'Role Name',
    code: 'Role Code',
    description: 'Description',
    permissions: 'Permissions',
    type: 'Type',
    actions: 'Actions'
  },
  status: {
    builtIn: 'Built-in',
    custom: 'Custom'
  },
  actions: {
    edit: 'Edit',
    delete: 'Delete'
  },
  labels: {
    name: 'Role Name',
    code: 'Role Code',
    description: 'Description',
    permissions: 'Permissions'
  },
  hints: {
    name: 'e.g. Operations Engineer',
    code: 'e.g. ops_engineer',
    description: 'Optional description',
    nameRequired: 'Please enter a role name',
    codeRequired: 'Please enter a role code'
  },
  matrix: {
    selectAll: 'Select all',
    selected: 'Selected {selected} / {total}'
  },
  permissions: {
    modules: {
      site: 'Website',
      database: 'Database',
      file: 'Files',
      terminal: 'Terminal',
      container: 'Containers',
      appstore: 'App Store',
      firewall: 'Firewall',
      backup: 'Backup',
      monitor: 'Monitor',
      user: 'Users',
      audit: 'Audit',
      setting: 'Settings',
      alert: 'Alerts',
      cron: 'Cron Jobs'
    },
    actions: {
      view: 'View',
      create: 'Create',
      update: 'Update',
      delete: 'Delete'
    }
  },
  messages: {
    loadError: 'Failed to load roles',
    updated: 'Role updated',
    created: 'Role created',
    saveFailed: 'Save failed',
    deleted: 'Role deleted',
    deleteFailed: 'Delete failed'
  },
  delete: {
    title: 'Delete Role',
    content: 'Delete role "{name}"? This action cannot be undone.'
  }
}
