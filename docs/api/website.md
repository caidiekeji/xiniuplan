# 网站 API

网站管理模块提供基于 Nginx 的网站生命周期管理，包括站点 CRUD、启停、Nginx 配置编辑、SSL 证书管理和访问日志查看。

## 网站列表

`GET /api/v1/sites`

获取网站列表（分页）。

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| pageSize | int | 10 | 每页条数 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "primary_domain": "example.com",
        "domains": ["www.example.com"],
        "type": "static",
        "runtime": "",
        "root_dir": "/www/wwwroot/example.com",
        "nginx_conf": "/www/server/panel/vhost/nginx/example.com.conf",
        "ssl_enabled": true,
        "ssl_provider": "letsencrypt",
        "ssl_expire_at": "2026-12-09T00:00:00Z",
        "ssl_auto_renew": true,
        "status": "running",
        "remark": "企业官网",
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-09-01T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```

### 网站对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 站点 ID |
| primary_domain | string | 主域名 |
| domains | array | 附加域名列表 |
| type | string | 类型：static / runtime / reverse_proxy |
| runtime | string | 运行时类型（runtime 类型有效） |
| root_dir | string | 网站根目录 |
| nginx_conf | string | Nginx 配置文件路径 |
| ssl_enabled | bool | 是否启用 SSL |
| ssl_provider | string | SSL 提供商：manual / letsencrypt |
| ssl_expire_at | string | SSL 证书到期时间 |
| ssl_auto_renew | bool | 是否自动续期 |
| status | string | 状态：running / stopped / error |
| remark | string | 备注 |

## 创建网站

`POST /api/v1/sites`

创建新网站，自动生成 Nginx 配置。

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| primary_domain | string | 是 | 主域名 |
| domains | array | 否 | 附加域名列表 |
| type | string | 是 | 类型：static / runtime / reverse_proxy |
| runtime | string | 否 | 运行时类型 |
| root_dir | string | 否 | 网站根目录 |
| remark | string | 否 | 备注 |

### 请求示例

```json
{
  "primary_domain": "example.com",
  "domains": ["www.example.com"],
  "type": "static",
  "root_dir": "/www/wwwroot/example.com",
  "remark": "企业官网"
}
```

### 响应示例

返回创建的网站对象（格式同网站列表中的 item）。

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 参数错误 |
| 50001 | 创建网站失败 |

## 获取网站详情

`GET /api/v1/sites/:id`

### 路径参数

| 参数 | 说明 |
|------|------|
| id | 站点 ID |

### 响应示例

返回单个网站对象。

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 无效的站点 ID |
| 40401 | 网站不存在 |

## 更新网站

`PUT /api/v1/sites/:id`

更新网站基本信息。

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| primary_domain | string | 否 | 主域名 |
| domains | array | 否 | 附加域名 |
| type | string | 否 | 类型 |
| root_dir | string | 否 | 根目录 |
| remark | string | 否 | 备注 |

### 响应示例

返回更新后的网站对象。

## 删除网站

`DELETE /api/v1/sites/:id`

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| delete_files | bool | false | 是否同时删除网站文件 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "deleted": true,
    "delete_files": false
  }
}
```

## 启动/停止/重启网站

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/sites/:id/start` | 启动网站 |
| POST | `/api/v1/sites/:id/stop` | 停止网站 |
| POST | `/api/v1/sites/:id/restart` | 重启网站 |

均返回操作后的网站对象。

## 获取 Nginx 配置

`GET /api/v1/sites/:id/config`

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "domain": "example.com",
    "content": "server {\n    listen 80;\n    server_name example.com;\n    root /www/wwwroot/example.com;\n    index index.html;\n}"
  }
}
```

## 更新 Nginx 配置

`PUT /api/v1/sites/:id/config`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| content | string | 是 | Nginx 配置内容 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "updated": true
  }
}
```

## 申请 SSL 证书

`POST /api/v1/sites/:id/ssl/apply`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| provider | string | 是 | 提供商：manual / letsencrypt |
| cert | string | 否 | 证书内容（manual 必填） |
| key | string | 否 | 私钥内容（manual 必填） |
| email | string | 否 | 邮箱（letsencrypt 必填） |

### 响应示例（letsencrypt）

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "site": { "id": 1, "status": "running" },
    "status": "issued"
  }
}
```

## 获取 SSL 信息

`GET /api/v1/sites/:id/ssl`

返回 SSL 证书详情（提供商、到期时间、自动续期状态等）。

## 获取网站日志

`GET /api/v1/sites/:id/logs`

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| type | string | access | 日志类型：access / error |
| lines | int | 100 | 返回行数 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "type": "access",
    "lines": 100,
    "content": "192.168.1.1 - - [09/Sep/2026:10:30:00 +0800] \"GET / HTTP/1.1\" 200 1234\n..."
  }
}
```
