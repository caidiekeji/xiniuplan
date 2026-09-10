import service, { request } from './request'

/* ==================== 类型定义 ==================== */

export interface FileItem {
  name: string
  path: string
  is_dir: boolean
  size: number
  mode: string
  mod_time: string
  extension?: string
}

export interface FileListData {
  path: string
  parent: string
  items: FileItem[]
}

export interface SearchResultItem {
  path: string
  name: string
  is_dir: boolean
  size: number
  mod_time: string
  mode?: string
  extension?: string
}

export interface SearchData {
  root: string
  keyword: string
  total: number
  items: SearchResultItem[]
}

export type CompressFormat = 'zip' | 'tar.gz'

/* ==================== 目录 / 文件 ==================== */

export function listFiles(path: string) {
  return request<FileListData>({
    url: '/files',
    method: 'get',
    params: { path }
  })
}

export function createDir(path: string) {
  return request({
    url: '/files/mkdir',
    method: 'post',
    data: { path }
  })
}

export function createFile(path: string, content = '') {
  return request({
    url: '/files/create',
    method: 'post',
    data: { path, content }
  })
}

export function getFileContent(path: string) {
  return request<{ path: string; content: string; size: number }>({
    url: '/files/content',
    method: 'get',
    params: { path }
  })
}

export function writeFileContent(path: string, content: string) {
  return request({
    url: '/files/content',
    method: 'put',
    data: { path, content }
  })
}

export function renameFile(old_path: string, new_path: string) {
  return request({
    url: '/files/rename',
    method: 'post',
    data: { old_path, new_path }
  })
}

export function copyFile(src: string, dst: string) {
  return request({
    url: '/files/copy',
    method: 'post',
    data: { src, dst }
  })
}

export function moveFile(src: string, dst: string) {
  return request({
    url: '/files/move',
    method: 'post',
    data: { src, dst }
  })
}

export function deleteFiles(paths: string[]) {
  return request({
    url: '/files',
    method: 'delete',
    data: { paths }
  })
}

/* ==================== 压缩 / 解压 ==================== */

export function compressFiles(src_paths: string[], dest_path: string, format: CompressFormat) {
  return request({
    url: '/files/compress',
    method: 'post',
    data: { src_paths, dest_path, format }
  })
}

export function decompressFile(archive_path: string, dest_path: string) {
  return request({
    url: '/files/decompress',
    method: 'post',
    data: { archive_path, dest_path }
  })
}

/* ==================== 搜索 ==================== */

export function searchFiles(root: string, keyword: string, regex = false) {
  return request<SearchData>({
    url: '/files/search',
    method: 'get',
    params: { root, keyword, regex }
  })
}

/* ==================== 上传 / 下载 ==================== */

export interface UploadProgress {
  (percent: number): void
}

export function uploadFile(formData: FormData, onProgress?: UploadProgress) {
  return request({
    url: '/files/upload',
    method: 'post',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: (e) => {
      if (onProgress && e.total) {
        onProgress(Math.round((e.loaded / e.total) * 100))
      }
    }
  })
}

/** 下载文件，返回 blob URL（调用方用完后记得 revokeObjectURL） */
export async function downloadFile(path: string): Promise<string> {
  const res = await service.get('/files/download', {
    params: { path },
    responseType: 'blob'
  })
  const blob = res.data as Blob
  return URL.createObjectURL(blob)
}

/** 下载文件并触发浏览器保存（便捷方法） */
export async function downloadFileAs(path: string, filename?: string) {
  const url = await downloadFile(path)
  const a = document.createElement('a')
  a.href = url
  a.download = filename || path.split('/').pop() || 'download'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
