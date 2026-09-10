# 云枢面板 - Docker 安装部署文档

本文档介绍如何使用 Docker 和 Docker Compose 快速部署云枢面板。

---

## 目录

1. [前置要求](#1-前置要求)
2. [快速启动](#2-快速启动)
3. [访问与登录](#3-访问与登录)
4. [环境变量说明](#4-环境变量说明)
5. [数据卷说明](#5-数据卷说明)
6. [服务配置说明](#6-服务配置说明)
7. [仅启动核心服务 vs 启动数据库服务](#7-仅启动核心服务-vs-启动数据库服务)
8. [数据备份与恢复](#8-数据备份与恢复)
9. [升级方式](#9-升级方式)
10. [查看日志](#10-查看日志)
11. [停止与卸载](#11-停止与卸载)
12. [常见问题](#12-常见问题)

---

## 1. 前置要求

### 1.1 软件要求

| 软件 | 最低版本 | 说明 |
|------|---------|------|
| Docker | 20.10+ | 容器运行时 |
| Docker Compose | 2.0+ | 容器编排（`docker compose` 子命令） |

### 1.2 检查安装

```bash
# 检查 Docker 版本
docker --version

# 检查 Docker Compose 版本
docker compose version
```

### 1.3 端口要求

确保以下端口未被占用：

| 端口 | 服务 | 说明 |
|------|------|------|
| 8080 | frontend | 前端 Web 界面（Nginx） |
| 8888 | backend | 后端 API 服务（Gin） |

> 如需修改端口，编辑 `docker-compose.yml` 中对应服务的 `ports` 映射。

---

## 2. 快速启动

### 2.1 克隆项目

```bash
git clone <仓库地址>
cd fuwuqi
```

### 2.2 （可选）配置环境变量

在项目根目录创建 `.env` 文件，覆盖默认配置：

```bash
# 管理员初始密码（首次启动生效）
ADMIN_PASSWORD=your_secure_password

# MySQL root 密码（仅启动 database profile 时使用）
MYSQL_ROOT_PASSWORD=your_mysql_password

# 时区
TZ=Asia/Shanghai
```

> 不创建 `.env` 文件时，将使用 `docker-compose.yml` 中的默认值。

### 2.3 构建并启动

```bash
# 构建镜像并后台启动所有核心服务
docker compose up -d --build
```

首次构建需要下载基础镜像和编译前后端，可能需要 5-15 分钟，请耐心等待。

### 2.4 查看启动状态

```bash
# 查看容器运行状态
docker compose ps

# 实时查看日志（确认服务启动成功）
docker compose logs -f
```

当看到后端日志输出 `[云枢面板] 服务启动，监听 :8888` 时，表示启动成功。

---

## 3. 访问与登录

### 3.1 访问地址

| 环境 | 地址 |
|------|------|
| 本地部署 | http://localhost:8080 |
| 服务器部署 | http://<服务器IP>:8080 |

### 3.2 默认账号

| 项目 | 值 |
|------|-----|
| 用户名 | `admin` |
| 密码 | `admin123`（或 `.env` 中 `ADMIN_PASSWORD` 设置的值） |

> **安全提示**：首次登录后请立即修改默认密码。进入「系统设置 → 用户管理」修改。

---

## 4. 环境变量说明

在项目根目录 `.env` 文件中可配置以下变量：

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `ADMIN_PASSWORD` | `admin123` | 管理员初始密码。仅在数据库中不存在 admin 用户时（首次启动）生效；后续修改以数据库为准。 |
| `MYSQL_ROOT_PASSWORD` | `cloudpivot` | MySQL root 用户密码。仅在使用 `--profile database` 启动 MySQL 时生效。 |
| `TZ` | `Asia/Shanghai` | 容器时区。影响日志时间戳和定时任务执行时间。 |

### 4.1 .env 示例

```bash
# ===== 安全配置 =====
ADMIN_PASSWORD=MyStr0ngP@ssw0rd!

# ===== 数据库配置（可选）=====
MYSQL_ROOT_PASSWORD=MySqlR00tP@ss

# ===== 时区 =====
TZ=Asia/Shanghai
```

---

## 5. 数据卷说明

Docker Compose 使用命名卷（Named Volume）持久化数据。数据不会因容器删除而丢失。

| 数据卷 | 容器内路径 | 内容说明 |
|--------|-----------|---------|
| `backend-data` | `/opt/cloudpivot/data` | SQLite 数据库文件（`cloudpivot.db`）、JWT RSA 密钥对 |
| `backend-store` | `/opt/cloudpivot/store` | 应用商店定义文件（`apps/` 目录） |
| `www-data` | `/www` | 网站文件根目录（网站管理功能使用） |
| `mysql-data` | `/var/lib/mysql` | MySQL 数据文件（仅 database profile） |
| `redis-data` | `/data` | Redis 持久化数据（仅 database profile） |

### 5.1 查看数据卷

```bash
# 列出所有数据卷
docker volume ls | grep cloudpivot

# 查看数据卷详情
docker volume inspect fuwuqi_backend-data
```

### 5.2 数据卷位置

Docker 命名卷默认存储在宿主机的 `/var/lib/docker/volumes/` 目录下（Linux）。不建议直接操作该目录，应通过 `docker run --rm -v ...` 方式备份。

---

## 6. 服务配置说明

### 6.1 后端服务（backend）

- **镜像**：多阶段构建，Go 编译为静态二进制，运行在 `alpine:3.20` 上
- **端口**：`8888`
- **用户**：非 root 用户 `cloudpivot` 运行
- **健康检查**：每 30 秒请求 `GET /health`，连续 3 次失败标记为 unhealthy
- **Docker Socket**：挂载 `/var/run/docker.sock`，用于容器管理、数据库实例创建等功能

### 6.2 前端服务（frontend）

- **镜像**：多阶段构建，Node.js 编译后由 `nginx:alpine` 提供静态服务
- **端口**：`8080`（映射到容器内 80）
- **API 代理**：`/api/` 路径反向代理到 `backend:8888`
- **WebSocket**：支持 WebSocket 升级（终端、实时日志功能）
- **SPA 路由**：所有非文件路径回退到 `index.html`（Vue Router history 模式）
- **静态缓存**：`/assets/` 路径设置 30 天缓存

### 6.3 网络

所有服务连接到 `cloudpivot` 桥接网络，服务间通过容器名（`backend`、`frontend`、`mysql`、`redis`）互相访问。

---

## 7. 仅启动核心服务 vs 启动数据库服务

`docker-compose.yml` 中定义了两组服务：

### 7.1 核心服务（默认）

包含 `backend` 和 `frontend`，满足面板基本运行需求。

```bash
# 启动核心服务
docker compose up -d
```

### 7.2 含数据库服务（可选）

额外启动 `mysql` 和 `redis`，可作为面板中创建数据库实例的本地依赖。

```bash
# 启动核心服务 + MySQL + Redis
docker compose --profile database up -d
```

> **说明**：MySQL 和 Redis 使用 `profiles: [database]` 标记，默认不启动。面板的数据库管理功能也可以连接外部数据库，不依赖这两个容器。

### 7.3 仅启动数据库服务

```bash
# 仅启动 MySQL 和 Redis（不启动面板）
docker compose --profile database up -d mysql redis
```

---

## 8. 数据备份与恢复

### 8.1 备份后端数据

后端核心数据存储在 `backend-data` 卷中，包含 SQLite 数据库和 JWT 密钥。

```bash
# 创建备份目录
mkdir -p ./backup

# 备份 backend-data 卷到 tar 包
docker run --rm \
  -v fuwuqi_backend-data:/data \
  -v $(pwd)/backup:/backup \
  alpine tar czf /backup/backend-data-$(date +%Y%m%d_%H%M%S).tar.gz -C /data .
```

### 8.2 备份网站文件

```bash
docker run --rm \
  -v fuwuqi_www-data:/data \
  -v $(pwd)/backup:/backup \
  alpine tar czf /backup/www-data-$(date +%Y%m%d_%H%M%S).tar.gz -C /data .
```

### 8.3 一键备份脚本

```bash
#!/bin/bash
# backup.sh - 一键备份所有云枢面板数据
BACKUP_DIR="./backup/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

for vol in backend-data backend-store www-data; do
  echo "备份 $vol ..."
  docker run --rm \
    -v "fuwuqi_${vol}:/data" \
    -v "$(pwd)/${BACKUP_DIR}:/backup" \
    alpine tar czf "/backup/${vol}.tar.gz" -C /data .
done

echo "备份完成：$BACKUP_DIR"
```

### 8.4 恢复数据

```bash
# 停止服务
docker compose down

# 恢复 backend-data（替换为实际备份文件路径）
docker run --rm \
  -v fuwuqi_backend-data:/data \
  -v $(pwd)/backup/backend-data-20250101_120000.tar.gz:/backup/restore.tar.gz \
  alpine sh -c "cd /data && tar xzf /backup/restore.tar.gz"

# 重新启动
docker compose up -d
```

> **警告**：恢复操作会覆盖当前数据，请谨慎操作，恢复前建议先备份当前数据。

---

## 9. 升级方式

### 9.1 从源码升级

```bash
# 1. 拉取最新代码
git pull

# 2. 重新构建镜像并启动
docker compose up -d --build

# 3. 查看启动日志确认无误
docker compose logs -f --tail=50
```

### 9.2 升级后清理旧镜像

```bash
# 清理未使用的镜像（释放磁盘空间）
docker image prune -f

# 清理构建缓存（可选，释放更多空间）
docker builder prune -f
```

### 9.3 回滚到之前版本

```bash
# 1. 查看 git 历史
git log --oneline -10

# 2. 切换到目标版本
git checkout <commit-hash>

# 3. 重新构建并启动
docker compose up -d --build
```

> **注意**：跨大版本升级时，数据库结构可能发生变化。建议升级前先备份数据（见第 8 节）。

---

## 10. 查看日志

### 10.1 查看所有服务日志

```bash
# 实时跟踪所有服务日志
docker compose logs -f

# 查看最近 100 行
docker compose logs --tail=100
```

### 10.2 查看单个服务日志

```bash
# 后端日志
docker compose logs -f backend

# 前端日志
docker compose logs -f frontend

# MySQL 日志（如已启动）
docker compose logs -f mysql
```

### 10.3 带时间戳查看

```bash
docker compose logs -f --timestamps backend
```

---

## 11. 停止与卸载

### 11.1 停止服务（保留数据）

```bash
# 停止并删除容器（数据卷保留）
docker compose down

# 仅停止，不删除容器
docker compose stop
```

### 11.2 重启服务

```bash
# 重启所有服务
docker compose restart

# 重启单个服务
docker compose restart backend
```

### 11.3 完全卸载（删除数据）

> **警告**：以下操作会删除所有数据，包括数据库、配置和网站文件，且不可恢复！

```bash
# 1. 停止并删除容器、网络
docker compose down

# 2. 删除所有数据卷（危险！数据将永久丢失）
docker volume rm \
  fuwuqi_backend-data \
  fuwuqi_backend-store \
  fuwuqi_www-data \
  fuwuqi_mysql-data \
  fuwuqi_redis-data

# 3. 删除构建的镜像（可选）
docker rmi cloudpivot-backend:latest cloudpivot-frontend:latest
```

---

## 12. 常见问题

### Q1: 启动后访问 http://localhost:8080 显示 502 Bad Gateway

**原因**：后端服务尚未启动完成，或启动失败。

**解决**：
```bash
# 查看后端日志
docker compose logs backend

# 查看后端容器状态
docker compose ps backend
```
常见原因：端口 8888 被占用、配置文件格式错误。

---

### Q2: 后端日志显示 "加载配置失败"

**原因**：`config.yaml` 未正确复制到镜像中，或格式错误。

**解决**：
```bash
# 进入容器检查配置文件
docker compose exec backend cat /opt/cloudpivot/config.yaml

# 确认文件存在
docker compose exec backend ls -la /opt/cloudpivot/
```

---

### Q3: 容器管理功能无法使用，提示连接 Docker 失败

**原因**：Docker Socket 未正确挂载，或非 root 用户无权限访问。

**解决**：
1. 确认 `docker-compose.yml` 中已挂载 `/var/run/docker.sock`
2. 检查宿主机 Docker Socket 权限：
   ```bash
   ls -la /var/run/docker.sock
   ```
3. 如权限不足，可临时在 `docker-compose.yml` 的 backend 服务中添加 `user: root` 运行（不推荐生产环境），或调整宿主机 docker 组 GID 与容器内一致。

---

### Q4: WebSocket 终端连接失败

**原因**：Nginx 未正确配置 WebSocket 升级，或后端服务异常。

**解决**：
1. 确认 `frontend/nginx.conf` 中包含 `Upgrade` 和 `Connection` 头配置
2. 确认后端服务健康：`docker compose ps backend`
3. 检查浏览器控制台是否有 WebSocket 连接错误

---

### Q5: 如何修改前端访问端口？

**解决**：编辑 `docker-compose.yml`，修改 frontend 服务的端口映射：
```yaml
ports:
  - "9090:80"   # 将宿主机 9090 端口映射到容器 80 端口
```
然后执行 `docker compose up -d` 重启。

---

### Q6: 如何让后端不暴露 8888 端口？

**原因**：生产环境中后端 API 应仅通过前端 Nginx 代理访问，不直接暴露。

**解决**：注释或删除 `docker-compose.yml` 中 backend 服务的 `ports` 配置：
```yaml
# ports:
#   - "8888:8888"
```
前端容器仍可通过内部网络 `backend:8888` 访问后端。

---

### Q7: 构建时 npm ci 失败

**原因**：`package-lock.json` 与 `package.json` 不同步，或网络问题。

**解决**：
```bash
# 在本地重新生成 lock 文件
cd frontend
rm -rf node_modules package-lock.json
npm install
cd ..

# 重新构建
docker compose build --no-cache frontend
```

---

### Q8: 数据卷名称为什么带前缀？

**原因**：Docker Compose 默认以 `项目目录名_` 作为数据卷前缀。本项目目录为 `fuwuqi`，因此数据卷名为 `fuwuqi_backend-data` 等。

**解决**：如需自定义前缀，在项目根目录创建 `.env` 文件并设置：
```bash
COMPOSE_PROJECT_NAME=cloudpivot
```
此后数据卷名将变为 `cloudpivot_backend-data`。

---

### Q9: 如何进入后端容器调试？

```bash
# 进入后端容器（sh，alpine 无 bash）
docker compose exec backend sh

# 进入前端容器
docker compose exec frontend sh
```

---

### Q10: 服务器重启后服务会自动启动吗？

**答**：会。`backend` 和 `frontend` 配置了 `restart: always`，Docker 守护进程启动后会自动拉起这些容器。MySQL 和 Redis 为 `restart: unless-stopped`，在手动停止后不会自动重启。

---

## 附录：文件清单

| 文件路径 | 说明 |
|---------|------|
| `Dockerfile.backend` | 后端多阶段构建 Dockerfile（项目根目录） |
| `frontend/Dockerfile` | 前端多阶段构建 Dockerfile |
| `frontend/nginx.conf` | Nginx 反向代理配置 |
| `frontend/.dockerignore` | 前端构建上下文忽略规则 |
| `docker-compose.yml` | Docker Compose 编排文件 |
| `.dockerignore` | 后端构建上下文忽略规则（项目根目录） |
| `docs/docker-install.md` | 本文档 |
