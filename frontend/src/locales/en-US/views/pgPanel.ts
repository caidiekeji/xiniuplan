// pgPanel page texts
export default {
  tabs: {
    schema: 'Tables',
    users: 'Users & Privileges',
    sql: 'SQL Executor'
  },
  refresh: 'Refresh',
  execute: 'Execute',
  delete: 'Delete',
  createFailed: 'Creation failed',
  deleteFailed: 'Deletion failed',
  selectDb: 'Please select a database',
  selectDbLabel: 'Select database',
  selectDbPlaceholder: 'Select a database',
  loadDatabasesFailed: 'Failed to load database list',
  loadTablesFailed: 'Failed to load table list',
  loadSchemaFailed: 'Failed to load table schema',
  loadTableDataFailed: 'Failed to load table data',
  loadUsersFailed: 'Failed to load user list',
  columns: {
    tableName: 'Table',
    rowsEstimate: 'Rows (est.)',
    size: 'Size',
    actions: 'Actions',
    structure: 'Structure',
    data: 'Data',
    columnName: 'Column',
    type: 'Type',
    nullable: 'Nullable',
    defaultValue: 'Default',
    primaryKey: 'Primary Key',
    username: 'Username',
    attributes: 'Attributes',
    validUntil: 'Valid Until',
    permanent: 'Forever',
    grant: 'Grant',
    revoke: 'Revoke'
  },
  sidebarTitle: 'Databases',
  noDatabases: 'No databases',
  selectDbLeft: 'Please select a database on the left',
  noTables: 'No tables in this database',
  tableSchema: 'Table schema: {name}',
  tableData: 'Table data: {name}',
  indexes: 'Indexes',
  noUsers: 'No users, click the button above to create one',
  allowDestructive: 'Allow destructive operations',
  selectDbFirst: 'Please select a database first',
  sql: {
    enterSql: 'Please enter a SQL statement',
    executed: 'Execution completed',
    failed: 'Execution failed',
    success: 'Execution succeeded',
    duration: 'Took {ms} ms',
    affected: '{rows} rows affected',
    noResultSet: 'No result set (non-SELECT statement)',
    defaultSql: '-- Enter SQL statements here\nSELECT 1;'
  },
  createDb: {
    title: 'Create Database',
    nameLabel: 'Database Name',
    namePlaceholder: 'e.g. myapp_db',
    encodingLabel: 'Encoding',
    collateLabel: 'Collation',
    ownerPlaceholder: 'Leave empty to use current user',
    nameRequired: 'Please enter a database name',
    success: 'Database created successfully'
  },
  dbDeleted: 'Database {name} has been deleted',
  createUser: {
    title: 'Create Database User',
    create: 'Create User',
    nameLabel: 'Username',
    namePlaceholder: 'e.g. app_user',
    passwordLabel: 'Password',
    passwordPlaceholder: 'Please enter a password',
    roleAttributesLabel: 'Role Attributes',
    nameRequired: 'Please enter a username',
    passwordRequired: 'Please enter a password',
    success: 'User created successfully'
  },
  userDeleted: 'User {name} has been deleted',
  grant: {
    title: 'Grant to user: {name}',
    privilegesLabel: 'Privileges',
    success: 'Granted successfully',
    failed: 'Failed to grant privileges'
  },
  revoke: {
    title: 'Revoke privileges: {name}',
    dbPlaceholder: 'Select the database to revoke from',
    success: 'Revoked privileges of {user} on {db}',
    failed: 'Failed to revoke privileges'
  },
  deleteDb: {
    title: 'Delete Database',
    content: 'Are you sure you want to delete database "{name}"? This action cannot be undone.'
  },
  deleteUser: {
    title: 'Delete User',
    content: 'Are you sure you want to delete user "{name}"?'
  }
}
