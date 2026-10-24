import { request } from './request'

/* ==================== 类型定义 ==================== */

export type DBType = 'mysql' | 'postgres' | 'redis'
export type InstanceStatus = 'running' | 'stopped' | 'error' | 'pending'

/* ==================== 类型常量 ==================== */

export const DB_VERSION_OPTIONS: Record<DBType, { label: string; value: string }[]> = {
  mysql: [
    { label: 'MySQL 8.0', value: '8.0' },
    { label: 'MySQL 5.7', value: '5.7' }
  ],
  postgres: [
    { label: 'PostgreSQL 16', value: '16' },
    { label: 'PostgreSQL 15', value: '15' },
    { label: 'PostgreSQL 14', value: '14' }
  ],
  redis: [
    { label: 'Redis 7', value: '7' },
    { label: 'Redis 6', value: '6' }
  ]
}

export const DB_DEFAULT_PORT: Record<DBType, number> = {
  mysql: 3306,
  postgres: 5432,
  redis: 6379
}

export const DB_TYPE_LABELS: Record<DBType, string> = {
  mysql: 'MySQL',
  postgres: 'PostgreSQL',
  redis: 'Redis'
}

export interface DBInstance {
  id: number
  name?: string
  type: DBType
  version: string
  port: number
  status: InstanceStatus
  container_id: string
  root_password?: string
  remark?: string
  created_at: string
}

export interface InstanceListParams {
  page?: number
  pageSize?: number
}

export interface InstanceListData {
  items: DBInstance[]
  total: number
}

export interface CreateInstanceData {
  type: DBType
  version: string
  port?: number
  root_password?: string
  remark?: string
}

export interface DatabaseInfo {
  name: string
  charset: string
  collation: string
}

export interface CreateDatabaseData {
  name: string
  charset: string
  collation: string
}

export interface TableInfo {
  name: string
  engine?: string
  rows?: number
  collation?: string
  created_at?: string
}

export interface TableColumn {
  field: string
  type: string
  null: string
  key: string
  default: string | null
  extra: string
}

export interface TableStructure {
  columns: TableColumn[]
  indexes: any[]
}

export interface TableDataResult {
  columns: string[]
  rows: any[][]
  total: number
}

export interface DBUser {
  username: string
  host: string
}

export interface CreateUserData {
  username: string
  password: string
  host: string
}

export interface GrantUserData {
  host: string
  db_name: string
  privileges: string[]
}

export interface ExecuteSQLData {
  db_name: string
  sql: string
  allow_destructive: boolean
}

export interface SQLResult {
  columns?: string[]
  rows?: any[][]
  affected_rows?: number
  execution_time_ms?: number
  duration_ms?: number
  error?: string
}

export interface ImportSQLData {
  db_name: string
  content: string
}

export interface ExportSQLParams {
  db_name: string
  table_name?: string
}

/* ==================== 实例管理 ==================== */

export function getInstances(params?: InstanceListParams) {
  return request<InstanceListData>({
    url: '/databases',
    method: 'get',
    params
  })
}

export function createInstance(data: CreateInstanceData) {
  return request<DBInstance>({
    url: '/databases',
    method: 'post',
    data
  })
}

export function getInstance(id: number | string) {
  return request<DBInstance>({
    url: `/databases/${id}`,
    method: 'get'
  })
}

export function deleteInstance(id: number | string) {
  return request({
    url: `/databases/${id}`,
    method: 'delete'
  })
}

export function startInstance(id: number | string) {
  return request({
    url: `/databases/${id}/start`,
    method: 'post'
  })
}

export function stopInstance(id: number | string) {
  return request({
    url: `/databases/${id}/stop`,
    method: 'post'
  })
}

export function restartInstance(id: number | string) {
  return request({
    url: `/databases/${id}/restart`,
    method: 'post'
  })
}

/* ==================== 数据库管理 ==================== */

export function getDatabases(id: number | string) {
  return request<DatabaseInfo[]>({
    url: `/databases/${id}/dbs`,
    method: 'get'
  })
}

export function createDatabase(id: number | string, data: CreateDatabaseData) {
  return request<DatabaseInfo>({
    url: `/databases/${id}/dbs`,
    method: 'post',
    data
  })
}

export function deleteDatabase(id: number | string, dbName: string) {
  return request({
    url: `/databases/${id}/dbs/${encodeURIComponent(dbName)}`,
    method: 'delete'
  })
}

/* ==================== 表管理 ==================== */

export function getTables(id: number | string, dbName: string) {
  return request<TableInfo[]>({
    url: `/databases/${id}/dbs/${encodeURIComponent(dbName)}/tables`,
    method: 'get'
  })
}

export async function getTableStructure(id: number | string, dbName: string, tableName: string) {
  // 后端返回 { columns, indexes }，页面按数组消费，这里统一取 columns
  const res = await request<{ columns?: TableColumn[]; indexes?: unknown[] }>({
    url: `/databases/${id}/dbs/${encodeURIComponent(dbName)}/tables/${encodeURIComponent(tableName)}`,
    method: 'get'
  })
  return { ...res, data: res.data?.columns || [] }
}

export function getTableData(
  id: number | string,
  dbName: string,
  tableName: string,
  page = 1,
  pageSize = 20
) {
  return request<TableDataResult>({
    url: `/databases/${id}/dbs/${encodeURIComponent(dbName)}/tables/${encodeURIComponent(tableName)}/data`,
    method: 'get',
    params: { page, pageSize }
  })
}

/* ==================== 用户管理 ==================== */

export function getUsers(id: number | string) {
  return request<DBUser[]>({
    url: `/databases/${id}/users`,
    method: 'get'
  })
}

export function createUser(id: number | string, data: CreateUserData) {
  return request<DBUser>({
    url: `/databases/${id}/users`,
    method: 'post',
    data
  })
}

export function deleteUser(id: number | string, username: string, host: string) {
  return request({
    url: `/databases/${id}/users/${encodeURIComponent(username)}`,
    method: 'delete',
    params: { host }
  })
}

export function grantUser(id: number | string, username: string, data: GrantUserData) {
  return request({
    url: `/databases/${id}/users/${encodeURIComponent(username)}/grant`,
    method: 'post',
    data
  })
}

/* ==================== SQL 执行 / 导入导出 ==================== */

export function executeSQL(id: number | string, data: ExecuteSQLData) {
  return request<SQLResult>({
    url: `/databases/${id}/sql`,
    method: 'post',
    data
  })
}

export function importSQL(id: number | string, data: ImportSQLData) {
  return request<{ affected_rows: number; execution_time_ms?: number; duration_ms?: number }>({
    url: `/databases/${id}/import`,
    method: 'post',
    data
  })
}

export function exportSQL(id: number | string, params: ExportSQLParams) {
  return request<{ content: string; filename: string }>({
    url: `/databases/${id}/export`,
    method: 'get',
    params
  })
}

/* ==================== 导入已有实例 / 权限回收 ==================== */

export interface DiscoveredDB {
  type: string
  host: string
  port: number
  version: string
  running: boolean
}

export interface ImportInstanceData {
  type: string
  host: string
  port: number
  username?: string
  password?: string
  version?: string
  remark?: string
}

// 自动发现本机已安装的数据库
export function discoverInstances() {
  return request<DiscoveredDB[]>({
    url: '/databases/discover',
    method: 'get'
  })
}

// 导入已有数据库实例（不创建容器）
export function importInstance(data: ImportInstanceData) {
  return request<{ id: number }>({
    url: '/databases/import',
    method: 'post',
    data
  })
}

// 回收 MySQL 用户权限
export function revokeUser(id: number | string, username: string, params: { host?: string; db_name?: string }) {
  return request({
    url: `/databases/${id}/users/${username}/revoke`,
    method: 'post',
    params
  })
}

// 回收 PostgreSQL 用户权限
export function revokePgUser(id: number | string, username: string, dbName: string) {
  return request({
    url: `/postgres/${id}/users/${username}/revoke`,
    method: 'post',
    params: { db_name: dbName }
  })
}

