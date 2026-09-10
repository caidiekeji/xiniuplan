import { request } from './request'

/* ==================== 类型定义 ==================== */

export interface User {
  id: number
  username: string
  role: string
  role_id: number
  status: 'active' | 'disabled'
  totp_enabled: boolean
  last_login_at: string | null
  created_at: string
}

export interface UserListParams {
  page: number
  pageSize: number
  keyword?: string
}

export interface UserListData {
  items: User[]
  total: number
}

export interface CreateUserData {
  username: string
  password: string
  role_id: number
  status?: 'active' | 'disabled'
}

export interface UpdateUserData {
  username?: string
  role_id?: number
  status?: 'active' | 'disabled'
}

export interface UpdatePasswordData {
  old_password?: string
  new_password: string
}

export interface Role {
  id: number
  name: string
  code: string
  description: string
  permissions: string[]
  is_system: boolean
}

export interface CreateRoleData {
  name: string
  code: string
  description?: string
  permissions: string[]
}

export interface UpdateRoleData {
  name?: string
  description?: string
  permissions?: string[]
}

export interface PermissionItem {
  code: string
  label: string
}

export interface PermissionGroup {
  module: string
  label: string
  permissions: PermissionItem[]
}

/* ==================== 用户管理 ==================== */

export function getUsers(params: UserListParams) {
  return request<UserListData>({
    url: '/users',
    method: 'get',
    params
  })
}

export function createUser(data: CreateUserData) {
  return request<User>({
    url: '/users',
    method: 'post',
    data
  })
}

export function getUser(id: number) {
  return request<User>({
    url: `/users/${id}`,
    method: 'get'
  })
}

export function updateUser(id: number, data: UpdateUserData) {
  return request<User>({
    url: `/users/${id}`,
    method: 'put',
    data
  })
}

export function deleteUser(id: number) {
  return request({
    url: `/users/${id}`,
    method: 'delete'
  })
}

export function updatePassword(id: number, data: UpdatePasswordData) {
  return request({
    url: `/users/${id}/password`,
    method: 'put',
    data
  })
}

export function disableUser(id: number) {
  return request({
    url: `/users/${id}/disable`,
    method: 'post'
  })
}

export function enableUser(id: number) {
  return request({
    url: `/users/${id}/enable`,
    method: 'post'
  })
}

/* ==================== 角色管理 ==================== */

export function getRoles() {
  return request<Role[]>({
    url: '/roles',
    method: 'get'
  })
}

export function createRole(data: CreateRoleData) {
  return request<Role>({
    url: '/roles',
    method: 'post',
    data
  })
}

export function updateRole(id: number, data: UpdateRoleData) {
  return request<Role>({
    url: `/roles/${id}`,
    method: 'put',
    data
  })
}

export function deleteRole(id: number) {
  return request({
    url: `/roles/${id}`,
    method: 'delete'
  })
}

/* ==================== 权限 ==================== */

export function getPermissions() {
  return request<PermissionGroup[]>({
    url: '/permissions',
    method: 'get'
  })
}
