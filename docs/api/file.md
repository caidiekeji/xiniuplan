# 文件 API

文件管理模块提供服务器文件的浏览、创建、编辑、上传下载、压缩解压和搜索功能。文件操作限制在沙箱目录（`./sandbox`）内。

## 目录列表

`GET /api/v1/files`

列出指定目录下的文件和子目录。

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| path | string | / | 目录路径（相对于沙箱根目录） |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "path": "/www",
    "items": [
      {
        "name": "wwwroot",
        "path": "/www/wwwroot",
        "type": "dir",
        "size": 4096,
        "modified_at": "2026-09-01T00:00:00Z",
        "permission": "rwxr-xr-x",
        "owner": "www"
      },
      {
        "name": "index.html",
        "path": "/www/index.html",
        "type": "file",
        "size": 1234,
        "modified_at": "2026-09-09T10:00:00Z",
        "permission": "rw-r--r--",
        "owner": "root"
      }
    ]
  }
}
```

### 文件项字段

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 文件名 |
| path | string | 完整路径 |
| type | string | 类型：file / dir |
| size | int64 | 文件大小（字节） |
| modified_at | string | 修改时间 |
| permission | string | 权限字符串 |
| owner | string | 所有者 |

## 创建目录

`POST /api/v1/files/mkdir`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| path | string | 是 | 目录路径 |

### 响应示例

返回创建的目录信息对象。

## 创建文件或目录

`POST /api/v1/files/create`

统一创建接口，根据 type 决定创建文件还是目录。

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| path | string | 是 | 文件/目录路径 |
| type | string | 是 | 类型：file / dir |
| content | string | 否 | 文件初始内容（type=file 时有效） |

### 响应示例

返回创建的文件/目录信息对象。

## 读取文件内容

`GET /api/v1/files/content`

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| path | string | 是 | 文件路径 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "path": "/www/wwwroot/example.com/index.html",
    "content": "<!DOCTYPE html>\n<html>\n<head>..."
  }
}
```

## 写入文件内容

`PUT /api/v1/files/content`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| path | string | 是 | 文件路径 |
| content | string | 是 | 文件内容 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

## 重命名

`POST /api/v1/files/rename`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| old_path | string | 是 | 原路径 |
| new_path | string | 是 | 新路径 |

## 复制

`POST /api/v1/files/copy`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| src | string | 是 | 源路径 |
| dst | string | 是 | 目标路径 |

## 移动

`POST /api/v1/files/move`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| src | string | 是 | 源路径 |
| dst | string | 是 | 目标路径 |

## 删除

`DELETE /api/v1/files`

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| paths | string | 是 | 要删除的路径，多个用逗号分隔 |

### 示例

```
DELETE /api/v1/files?paths=/tmp/a.txt,/tmp/b.txt
```

## 压缩

`POST /api/v1/files/compress`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| src_paths | array | 是 | 源路径列表 |
| dest_path | string | 是 | 压缩包保存路径 |
| format | string | 否 | 压缩格式，默认 zip |

## 解压

`POST /api/v1/files/decompress`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| archive_path | string | 是 | 压缩包路径 |
| dest_path | string | 是 | 解压目标目录 |

## 搜索

`GET /api/v1/files/search`

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| root | string | / | 搜索起始目录 |
| keyword | string | - | 搜索关键词（必填） |
| regex | bool | false | 是否使用正则表达式 |

### 响应示例

返回匹配的文件路径列表。

## 上传文件

`POST /api/v1/files/upload`

使用 `multipart/form-data` 格式上传文件。

### 表单字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 上传的文件 |
| path | string | 否 | 目标目录，默认 / |

### 响应示例

返回上传后的文件信息对象。

## 下载文件

`GET /api/v1/files/download`

直接返回文件流下载（Content-Disposition: attachment）。

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| path | string | 是 | 文件路径 |

> 仅支持下载单个文件，目录需先压缩再下载。
