import { request } from './request'

/* ==================== 类型定义 ==================== */

export interface PgDatabase {
  name: string
  owner?: string
  encoding?: string
  collate?: string
  size?: string
}

export interface CreatePgDatabaseData {
  name: string
  encoding?: string
  collate?: string
  owner?: string
}

export interface PgTable {
  name: string
  rows_estimate?: number
  size?: string
}

export interface PgColumn {
  name: string
  type: string
  nullable: boolean
  default: string | null
  is_primary: boolean
}

export interface PgIndex {
  name: string
  definition: string
}

export interface PgTableSchema {
  columns: PgColumn[]
  indexes: PgIndex[]
}

export interface PgTableDataResult {
  columns: string[]
  rows: any[][]
  total: number
}

export interface PgUser {
  username: string
  attributes: string[]
  valid_until?: string
}

export interface CreatePgUserData {
  username: string
  password: string
  attributes?: string[]
}

export interface PgGrantData {
  db: string
  privileges: string[]
}

export interface PgSQLData {
  db: string
  sql: string
  allow_destructive?: boolean
}

export interface PgSQLResult {
  columns?: string[]
  rows?: any[][]
  affected_rows?: number
  duration_ms?: number
  error?: string
}

/* ==================== 数据库管理 ==================== */

export function getDatabases(id: number | string) {
  return request<PgDatabase[]>({
    url: `/postgres/${id}/dbs`,
    method: 'get'
  })
}

export function createDatabase(id: number | string, data: CreatePgDatabaseData) {
  return request<PgDatabase>({
    url: `/postgres/${id}/dbs`,
    method: 'post',
    data
  })
}

export function dropDatabase(id: number | string, name: string) {
  return request({
    url: `/postgres/${id}/dbs/${encodeURIComponent(name)}`,
    method: 'delete'
  })
}

/* ==================== 表管理 ==================== */

export function getTables(id: number | string, db: string) {
  return request<PgTable[]>({
    url: `/postgres/${id}/dbs/${encodeURIComponent(db)}/tables`,
    method: 'get'
  })
}

export function getTableSchema(id: number | string, db: string, table: string) {
  return request<PgTableSchema>({
    url: `/postgres/${id}/dbs/${encodeURIComponent(db)}/tables/${encodeURIComponent(table)}`,
    method: 'get'
  })
}

export function getTableData(
  id: number | string,
  db: string,
  table: string,
  page = 1,
  pageSize = 20
) {
  return request<PgTableDataResult>({
    url: `/postgres/${id}/dbs/${encodeURIComponent(db)}/tables/${encodeURIComponent(table)}/data`,
    method: 'get',
    params: { page, pageSize }
  })
}

/* ==================== SQL 执行 ==================== */

export function execSQL(id: number | string, data: PgSQLData) {
  return request<PgSQLResult>({
    url: `/postgres/${id}/sql`,
    method: 'post',
    data
  })
}

/* ==================== 用户管理 ==================== */

export function getUsers(id: number | string) {
  return request<PgUser[]>({
    url: `/postgres/${id}/users`,
    method: 'get'
  })
}

export function createUser(id: number | string, data: CreatePgUserData) {
  return request<PgUser>({
    url: `/postgres/${id}/users`,
    method: 'post',
    data
  })
}

export function dropUser(id: number | string, name: string) {
  return request({
    url: `/postgres/${id}/users/${encodeURIComponent(name)}`,
    method: 'delete'
  })
}

export function grantUser(id: number | string, name: string, data: PgGrantData) {
  return request({
    url: `/postgres/${id}/users/${encodeURIComponent(name)}/grant`,
    method: 'post',
    data
  })
}

// 回收 PostgreSQL 用户权限
export function revokePgUser(id: number | string, name: string, dbName: string) {
  return request({
    url: `/postgres/${id}/users/` + encodeURIComponent(name) + `/revoke`,
    method: 'post',
    params: { db_name: dbName }
  })
}
