# 应用商店 API

应用商店模块提供应用浏览、一键安装、已安装应用管理和应用源管理功能。

## 应用列表

`GET /api/v1/apps`

获取所有可用应用列表，支持分类筛选、关键词搜索和应用源筛选。

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| category | string | 否 | 分类筛选 |
| keyword | string | 否 | 关键词搜索（匹配名称和描述） |
| source_id | string | 否 | 应用源 ID 筛选 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "key": "wordpress",
        "name": "WordPress",
        "description": "最流行的开源博客和内容管理系统",
        "category": "blog",
        "version": "6.4",
        "icon": "https://example.com/icons/wordpress.png",
        "installed": false,
        "source_id": "default"
      },
      {
        "key": "nextcloud",
        "name": "Nextcloud",
        "description": "自建私有云存储和协作平台",
        "category": "cloud",
        "version": "28.0",
        "icon": "https://example.com/icons/nextcloud.png",
        "installed": true,
        "source_id": "default"
      }
    ],
    "total": 2
  }
}
```

### 应用对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| key | string | 应用唯一标识 |
| name | string | 应用名称 |
| description | string | 应用描述 |
| category | string | 应用分类 |
| version | string | 应用版本 |
| icon | string | 应用图标 URL |
| installed | bool | 是否已安装 |
| source_id | string | 所属应用源 ID |

## 获取应用详情

`GET /api/v1/apps/:key`

### 路径参数

| 参数 | 说明 |
|------|------|
| key | 应用唯一标识 |

### 响应示例

返回应用详细信息，包括：
- 基本信息（名称、描述、版本、分类）
- 安装要求（端口、依赖、环境变量说明）
- 安装步骤说明
- 默认配置

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40401 | 应用不存在 |

## 已安装应用列表

`GET /api/v1/apps/installed`

获取所有已安装的应用。

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "key": "nextcloud",
      "name": "Nextcloud",
      "version": "28.0",
      "install_path": "/data/apps/nextcloud",
      "status": "running",
      "container_status": "running",
      "env_params": {
        "MYSQL_HOST": "127.0.0.1",
        "MYSQL_DATABASE": "nextcloud"
      },
      "created_at": "2026-01-01 00:00:00",
      "updated_at": "2026-09-01 00:00:00"
    }
  ]
}
```

### 已安装应用字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 安装记录 ID |
| key | string | 应用标识 |
| name | string | 应用名称 |
| version | string | 安装版本 |
| install_path | string | 安装路径 |
| status | string | 应用状态：running / stopped / error |
| container_status | string | 容器状态 |
| env_params | object | 环境变量配置 |
| created_at | string | 安装时间 |
| updated_at | string | 更新时间 |

## 安装应用

`POST /api/v1/apps/:key/install`

### 路径参数

| 参数 | 说明 |
|------|------|
| key | 应用唯一标识 |

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| source_id | string | 否 | 应用源 ID（默认使用应用所在源） |
| env_params | object | 否 | 环境变量配置（键值对） |

### 请求示例

```json
{
  "env_params": {
    "MYSQL_HOST": "127.0.0.1",
    "MYSQL_DATABASE": "wordpress",
    "MYSQL_USER": "wp_user",
    "MYSQL_PASSWORD": "secure_password"
  }
}
```

### 响应示例

返回已安装应用对象（格式同已安装应用列表中的 item）。

### 安装过程

1. 验证应用是否存在
2. 拉取 Docker 镜像（如未存在）
3. 创建数据目录
4. 使用指定环境变量创建并启动容器
5. 记录安装信息到数据库

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40401 | 应用不存在 |
| 50001 | 安装失败（镜像拉取失败、容器启动失败等） |

## 卸载应用

`POST /api/v1/apps/:key/uninstall`

### 路径参数

| 参数 | 说明 |
|------|------|
| key | 应用唯一标识 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": null
}
```

### 卸载过程

1. 停止并删除应用容器
2. 删除安装记录
3. 应用数据目录默认保留（如需彻底删除请手动删除）

## 获取应用安装状态

`GET /api/v1/apps/:key/status`

获取指定应用的安装状态和运行信息。

### 响应示例

返回应用安装状态对象（含容器状态、端口映射等）。

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40401 | 应用未安装或不存在 |

## 应用源管理

应用源管理接口由 `RegisterAppSourceRoutes` 注册，提供应用源的 CRUD 操作：

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/app-sources` | 应用源列表 |
| POST | `/api/v1/app-sources` | 添加应用源 |
| GET | `/api/v1/app-sources/:id` | 应用源详情 |
| PUT | `/api/v1/app-sources/:id` | 更新应用源 |
| DELETE | `/api/v1/app-sources/:id` | 删除应用源 |
| POST | `/api/v1/app-sources/:id/sync` | 同步应用源 |

### 应用源对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 源 ID |
| name | string | 源名称 |
| url | string | 源地址（应用清单 JSON URL） |
| status | string | 状态：active / error |
| app_count | int | 应用数量 |
| last_sync_at | string | 最后同步时间 |

## 注意事项

1. **Docker 依赖**：应用安装基于 Docker 容器，需要 Docker 服务正常运行
2. **端口冲突**：安装前确保应用所需端口未被占用
3. **数据保留**：卸载应用不会删除数据目录，升级应用时数据可保留
4. **应用源**：默认应用源由系统内置，可添加第三方应用源扩展应用列表
5. **安装时间**：首次安装需要拉取镜像，大镜像可能需要较长时间
