# 计划任务 API

计划任务模块提供基于 Cron 表达式的定时任务管理，支持 Shell 命令、备份、URL 请求和应用重启四种任务类型。

## 任务列表

`GET /api/v1/cron/tasks`

获取计划任务列表（分页）。

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
        "name": "每日数据库备份",
        "type": "backup",
        "schedule": "0 3 * * *",
        "command": "",
        "target": "database:1:mydb",
        "remark": "每天凌晨3点备份",
        "status": "enabled",
        "entry_id": 42,
        "last_run_at": "2026-09-09T03:00:00Z",
        "next_run_at": "2026-09-10T03:00:00Z",
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

### 任务对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 任务 ID |
| name | string | 任务名称 |
| type | string | 类型：shell / backup / url / app_restart |
| schedule | string | Cron 表达式（5 段） |
| command | string | Shell 命令（type=shell 时有效） |
| target | string | 目标（type=backup/url/app_restart 时有效） |
| remark | string | 备注 |
| status | string | 状态：enabled / disabled |
| entry_id | int | Cron 调度器条目 ID |
| last_run_at | string | 上次执行时间 |
| next_run_at | string | 下次执行时间（预测） |
| created_at | string | 创建时间 |
| updated_at | string | 更新时间 |

## 创建任务

`POST /api/v1/cron/tasks`

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 任务名称 |
| type | string | 是 | 类型：shell / backup / url / app_restart |
| schedule | string | 是 | Cron 表达式 |
| command | string | 否 | Shell 命令（shell 类型必填） |
| target | string | 否 | 目标（backup/url/app_restart 类型必填） |
| remark | string | 否 | 备注 |

### 请求示例（Shell 任务）

```json
{
  "name": "清理日志",
  "type": "shell",
  "schedule": "0 4 * * 0",
  "command": "find /www/wwwlogs -name '*.log' -mtime +30 -delete",
  "remark": "每周日凌晨4点清理30天前的日志"
}
```

### 请求示例（备份任务）

```json
{
  "name": "每日备份",
  "type": "backup",
  "schedule": "0 3 * * *",
  "target": "database:1:mydb",
  "remark": "每天凌晨3点备份数据库"
}
```

### 响应示例

返回创建的任务对象（status 默认为 enabled）。

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 参数错误（Cron 表达式无效、必填字段缺失等） |
| 50001 | 创建任务失败 |

## 更新任务

`PUT /api/v1/cron/tasks/:id`

### 请求体

字段同创建任务，所有字段可选（只传需要修改的字段）。

### 响应示例

返回更新后的任务对象。

## 删除任务

`DELETE /api/v1/cron/tasks/:id`

删除任务及其执行日志。

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

## 启用任务

`POST /api/v1/cron/tasks/:id/enable`

启用已禁用的任务，重新加入 Cron 调度。

### 响应示例

返回启用后的任务对象。

## 禁用任务

`POST /api/v1/cron/tasks/:id/disable`

禁用任务，从 Cron 调度中移除（保留配置和日志）。

### 响应示例

返回禁用后的任务对象。

## 手动触发任务

`POST /api/v1/cron/tasks/:id/trigger`

立即执行一次任务（不等待调度时间），用于测试任务是否正常工作。

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "triggered": true
  }
}
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 50001 | 触发任务失败 |

## 任务执行日志

`GET /api/v1/cron/tasks/:id/logs`

获取指定任务的执行日志（分页）。

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| pageSize | int | 20 | 每页条数 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 100,
        "task_id": 1,
        "task_name": "每日数据库备份",
        "start_time": "2026-09-09T03:00:00Z",
        "end_time": "2026-09-09T03:02:30Z",
        "output": "mysqldump: Exporting database 'mydb'...\nBackup completed: 100MB\n",
        "result": "success",
        "created_at": "2026-09-09T03:00:00Z"
      }
    ],
    "total": 30,
    "page": 1,
    "pageSize": 20
  }
}
```

### 日志对象字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 日志 ID |
| task_id | uint | 任务 ID |
| task_name | string | 任务名称 |
| start_time | string | 执行开始时间 |
| end_time | string | 执行结束时间 |
| output | string | 命令输出或执行结果 |
| result | string | 结果：running / success / failed |
| created_at | string | 创建时间 |

## Cron 表达式

Cron 表达式格式为 `分 时 日 月 周`，共 5 个字段：

| 字段 | 取值范围 | 特殊字符 |
|------|----------|----------|
| 分 | 0-59 | `*` `,` `-` `/` |
| 时 | 0-23 | `*` `,` `-` `/` |
| 日 | 1-31 | `*` `,` `-` `/` |
| 月 | 1-12 | `*` `,` `-` `/` |
| 周 | 0-7（0和7均为周日） | `*` `,` `-` `/` |

常用示例：

| 表达式 | 说明 |
|--------|------|
| `* * * * *` | 每分钟 |
| `0 * * * *` | 每小时整点 |
| `0 3 * * *` | 每天凌晨 3 点 |
| `0 3 * * 0` | 每周日凌晨 3 点 |
| `*/5 * * * *` | 每 5 分钟 |

## 注意事项

1. **任务引擎**：计划任务引擎在服务启动时自动初始化，服务重启后任务自动恢复调度
2. **并发执行**：同一任务上次执行未完成时，下次调度可能跳过（取决于 Cron 库实现）
3. **Shell 环境**：Shell 任务在面板进程的系统用户环境下执行，注意权限和路径问题
4. **日志保留**：执行日志长期保存，建议定期清理旧日志
5. **任务超时**：长时间运行的任务可能影响调度，建议将耗时操作拆分为多个任务
