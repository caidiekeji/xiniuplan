# 备份 API

备份恢复模块提供网站、数据库、应用和系统配置的备份与恢复功能。

## 备份列表

`GET /api/v1/backups`

获取备份记录列表（分页）。

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| pageSize | int | 10 | 每页条数 |
| type | string | - | 按类型筛选：site / database / app / system |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "type": "site",
        "target": "example.com",
        "file_path": "/data/backups/site_example.com_20260909.tar.gz",
        "size": 104857600,
        "status": "success",
        "remark": "每日自动备份",
        "error_msg": "",
        "created_at": "2026-09-09T03:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```

### 备份对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 备份 ID |
| type | string | 类型：site / database / app / system |
| target | string | 备份目标（网站域名、数据库名等） |
| file_path | string | 备份文件路径 |
| size | int64 | 文件大小（字节） |
| status | string | 状态：pending / running / success / failed |
| remark | string | 备注 |
| error_msg | string | 错误信息（失败时） |
| created_at | string | 创建时间 |

## 创建备份

`POST /api/v1/backups`

创建新的备份任务。

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | 备份类型：site / database / app / system |
| target | string | 是 | 备份目标（网站域名、数据库实例:库名等） |
| remark | string | 否 | 备注 |

### 请求示例

```json
{
  "type": "site",
  "target": "example.com",
  "remark": "改版前备份"
}
```

### 响应示例

返回创建的备份对象（状态为 pending 或 running）。

### 备份过程

1. 创建备份记录，状态设为 pending
2. 开始执行，状态变为 running
3. 根据类型执行对应备份操作：
   - site：压缩网站根目录
   - database：执行 mysqldump/pg_dump
   - app：压缩应用数据目录
   - system：导出面板配置和数据库
4. 保存文件到备份目录
5. 状态更新为 success 或 failed

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 参数错误 |
| 50001 | 创建备份失败 |

## 获取备份详情

`GET /api/v1/backups/:id`

### 路径参数

| 参数 | 说明 |
|------|------|
| id | 备份 ID |

### 响应示例

返回单个备份对象。

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 无效的备份 ID |
| 40401 | 备份不存在 |

## 删除备份

`DELETE /api/v1/backups/:id`

删除备份记录和对应的备份文件。

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "deleted": true
  }
}
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 50001 | 删除备份失败 |

## 恢复备份

`POST /api/v1/backups/:id/restore`

使用指定备份恢复数据。

### 路径参数

| 参数 | 说明 |
|------|------|
| id | 备份 ID |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "restored": true
  }
}
```

### 恢复操作说明

| 备份类型 | 恢复操作 |
|----------|----------|
| site | 解压备份文件覆盖网站根目录，恢复 Nginx 配置 |
| database | 执行 SQL 文件导入到目标数据库 |
| app | 解压备份文件覆盖应用数据目录 |
| system | 导入面板配置和数据库 |

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40401 | 备份不存在 |
| 50001 | 恢复失败 |

> **警告**：恢复操作会覆盖当前数据，不可撤销。恢复前务必备份当前数据。

## 下载备份

`GET /api/v1/backups/:id/download`

直接返回备份文件流下载。

### 响应头

```
Content-Type: application/octet-stream
Content-Disposition: attachment; filename=site_example.com_20260909.tar.gz
Content-Transfer-Encoding: binary
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40401 | 备份不存在 |
| 50001 | 下载失败（文件不存在等） |

## 注意事项

1. **备份存储**：备份文件存储在配置的备份目录中，确保磁盘空间充足
2. **大文件备份**：大型网站或数据库备份可能需要较长时间，期间可正常使用其他功能
3. **恢复风险**：数据库恢复会清空现有数据后导入备份，请谨慎操作
4. **备份保留**：建议配置备份保留策略，定期清理旧备份以节省空间
5. **异地备份**：重要备份建议下载到异地存储，防止服务器故障导致备份丢失
