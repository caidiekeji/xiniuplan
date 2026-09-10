# 数据库 API

数据库管理模块提供 MySQL、PostgreSQL、Redis 实例的完整管理，包括实例生命周期、库表管理、用户权限和 SQL 执行。所有实例基于 Docker 容器运行。

## 实例管理

### 实例列表

`GET /api/v1/databases`

获取数据库实例列表（分页）。

#### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| page_size | int | 10 | 每页条数 |

#### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "type": "mysql",
        "version": "8.0",
        "port": 3306,
        "container_id": "abc123def456",
        "status": "running",
        "container_status": "running",
        "created_at": "2026-01-01 00:00:00"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```

#### 实例对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 实例 ID |
| type | string | 类型：mysql / postgres / redis |
| version | string | 版本号 |
| port | int | 宿主机映射端口 |
| container_id | string | Docker 容器 ID |
| status | string | 记录状态 |
| container_status | string | 实际容器状态：running / stopped / error |
| created_at | string | 创建时间 |

### 创建实例

`POST /api/v1/databases`

#### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | 类型：mysql / postgres / redis |
| version | string | 是 | 版本，如 "8.0"、"16"、"7" |
| port | int | 是 | 宿主机端口 |
| root_password | string | 否 | root 密码（mysql/postgres） |

#### 响应示例

返回创建的实例对象。

### 获取实例详情

`GET /api/v1/databases/:id`

返回单个实例对象（含真实容器状态）。

### 删除实例

`DELETE /api/v1/databases/:id`

#### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| cleanup | bool | false | 是否同时删除数据卷 |

### 实例启停

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/databases/:id/start` | 启动实例 |
| POST | `/api/v1/databases/:id/stop` | 停止实例 |
| POST | `/api/v1/databases/:id/restart` | 重启实例 |

## 库表管理

### 数据库列表

`GET /api/v1/databases/:id/dbs`

返回实例下所有数据库名称数组。

### 创建数据库

`POST /api/v1/databases/:id/dbs`

#### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 数据库名 |
| charset | string | 否 | 字符集（MySQL，默认 utf8mb4） |
| collation | string | 否 | 排序规则 |

### 删除数据库

`DELETE /api/v1/databases/:id/dbs/:name`

### 表列表

`GET /api/v1/databases/:id/dbs/:name/tables`

返回指定数据库下所有表信息。

### 表结构

`GET /api/v1/databases/:id/dbs/:name/tables/:table`

返回表的字段、索引等结构信息。

### 表数据

`GET /api/v1/databases/:id/dbs/:name/tables/:table/data`

#### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| page_size | int | 10 | 每页条数 |

## 用户权限

### 用户列表

`GET /api/v1/databases/:id/users`

返回数据库用户列表。

### 创建用户

`POST /api/v1/databases/:id/users`

#### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |
| host | string | 否 | 允许连接的主机（MySQL，默认 %） |

### 删除用户

`DELETE /api/v1/databases/:id/users/:username`

#### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| host | string | % | 用户主机（MySQL） |

### 授权权限

`POST /api/v1/databases/:id/users/:username/grant`

#### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| host | string | 否 | 主机 |
| db_name | string | 是 | 数据库名 |
| privileges | array | 是 | 权限列表，如 ["SELECT","INSERT","ALL PRIVILEGES"] |

### 撤销权限

`POST /api/v1/databases/:id/users/:username/revoke`

#### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| host | string | 否 | 主机（默认 %） |
| db_name | string | 是 | 数据库名 |

## SQL 执行

### 执行 SQL

`POST /api/v1/databases/:id/sql`

#### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| db_name | string | 是 | 目标数据库名 |
| sql | string | 是 | SQL 语句 |
| allow_destructive | bool | 否 | 是否允许破坏性操作（DROP/TRUNCATE 等），默认 false |

#### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "affected_rows": 5,
    "columns": ["id", "name", "email"],
    "rows": [
      [1, "Alice", "alice@example.com"],
      [2, "Bob", "bob@example.com"]
    ]
  }
}
```

### 导入 SQL

`POST /api/v1/databases/:id/import`

#### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| db_name | string | 是 | 目标数据库名 |
| content | string | 是 | SQL 文件内容 |

### 导出 SQL

`GET /api/v1/databases/:id/export`

直接返回 `.sql` 文件下载（Content-Disposition: attachment）。

#### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| db_name | string | 是 | 数据库名 |
| table_name | string | 否 | 表名（不填则导出整库） |

## PostgreSQL 和 Redis 专用 API

PostgreSQL 和 Redis 还有各自的专用管理接口，路由注册在主路由组中：

- PostgreSQL：`RegisterPostgresRoutes` — 提供 PG 特有的管理功能
- Redis：`RegisterRedisRoutes` — 提供 Redis Key 管理、信息查看等功能

具体接口请参考后端代码 `backend/internal/api/postgres.go` 和 `backend/internal/api/redis.go`。
