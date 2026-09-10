# 容器 API

容器管理模块提供 Docker 容器和镜像的查看与管理功能，包括容器列表、启停、日志、资源统计和镜像列表。

## 容器列表

`GET /api/v1/containers`

获取所有 Docker 容器列表。

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| all | bool | true | 是否包含已停止的容器 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": "abc123def456789",
      "name": "mysql-8.0",
      "image": "mysql:8.0",
      "status": "Up 2 hours",
      "state": "running",
      "ports": [
        { "ip": "0.0.0.0", "private_port": 3306, "public_port": 3306, "type": "tcp" }
      ],
      "created": "2026-01-01T00:00:00Z",
      "labels": {
        "managed_by": "cloudpivot",
        "db_type": "mysql"
      }
    }
  ]
}
```

### 容器对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 容器 ID |
| name | string | 容器名称 |
| image | string | 使用的镜像 |
| status | string | 状态文本（如 "Up 2 hours"） |
| state | string | 状态：running / exited / created 等 |
| ports | array | 端口映射列表 |
| created | string | 创建时间 |
| labels | object | 容器标签 |

## 获取容器详情

`GET /api/v1/containers/:id`

### 路径参数

| 参数 | 说明 |
|------|------|
| id | 容器 ID（支持短 ID） |

### 响应示例

返回单个容器对象。

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40401 | 容器不存在 |

## 启动容器

`POST /api/v1/containers/:id/start`

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 50001 | 启动容器失败 |

## 停止容器

`POST /api/v1/containers/:id/stop`

发送停止信号，容器优雅退出。

## 重启容器

`POST /api/v1/containers/:id/restart`

先停止再启动容器。

## 删除容器

`DELETE /api/v1/containers/:id`

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| force | bool | false | 是否强制删除（运行中的容器需 force=true） |

### 错误码

| 错误码 | 说明 |
|--------|------|
| 50001 | 删除容器失败（如容器正在运行且未指定 force） |

## 容器日志

`GET /api/v1/containers/:id/logs`

获取容器的标准输出和标准错误日志。

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| tail | int | 100 | 返回最后多少行 |
| follow | bool | false | 是否实时流式输出（SSE） |

### 非流式响应示例（follow=false）

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "content": "2026-09-09T10:00:00Z [Note] mysqld: ready for connections.\n..."
  }
}
```

### 流式响应（follow=true）

通过 SSE（Server-Sent Events）实时推送日志行：

```
data: 2026-09-09T10:00:01Z New connection from 192.168.1.100

data: 2026-09-09T10:00:02Z Query SELECT * FROM users

```

- Content-Type: `text/event-stream`
- 每行日志以 `data: ` 前缀发送
- 适用于实时调试正在运行的应用

## 容器资源统计

`GET /api/v1/containers/:id/stats`

获取容器实时资源使用情况。

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "cpu_percent": 12.5,
    "memory_usage": 524288000,
    "memory_limit": 1073741824,
    "memory_percent": 48.8
  }
}
```

### 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| cpu_percent | float | CPU 使用率百分比 |
| memory_usage | int64 | 内存使用量（字节） |
| memory_limit | int64 | 内存限制（字节） |
| memory_percent | float | 内存使用率百分比 |

## 镜像列表

`GET /api/v1/images`

获取本地所有 Docker 镜像。

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": "sha256:abc123def456",
      "tags": ["mysql:8.0", "mysql:latest"],
      "size": 562000000,
      "created": "2026-01-01T00:00:00Z"
    },
    {
      "id": "sha256:def789abc012",
      "tags": ["nginx:alpine"],
      "size": 42000000,
      "created": "2026-02-01T00:00:00Z"
    }
  ]
}
```

### 镜像对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 镜像 ID（含 sha256: 前缀） |
| tags | array | 镜像标签列表 |
| size | int64 | 镜像大小（字节） |
| created | string | 创建时间 |

## 注意事项

1. **Docker 依赖**：所有容器操作需要服务器已安装并运行 Docker
2. **权限**：运行面板的用户需要 Docker 访问权限（加入 docker 用户组）
3. **镜像管理**：当前版本镜像管理为只读，拉取和删除镜像需通过 Web 终端执行 docker 命令
4. **日志流式**：follow=true 时使用 SSE，客户端需支持 EventSource 或手动解析 SSE 格式
