# 快速开始

本指南将帮助您在本地快速启动云枢面板开发环境，以及在生产服务器上完成部署。

## 环境要求

| 组件 | 最低版本 | 说明 |
|------|----------|------|
| Go | 1.22+ | 后端编译运行 |
| Node.js | 18+ | 前端构建，推荐使用 LTS 版本 |
| npm | 9+ | 随 Node.js 安装 |
| Docker | 20.10+ | 容器管理、数据库实例、应用商店功能依赖 |
| Git | 2.0+ | 代码克隆 |

> **注意**：云枢面板设计运行于 Linux 服务器。Windows/macOS 可用于前端和后端开发，但容器管理、Nginx 配置、SSH 等系统级功能需要 Linux 环境才能完整运行。

## 克隆项目

```bash
git clone <repository-url> fuwuqi
cd fuwuqi
```

## 后端启动

### 1. 进入后端目录

```bash
cd backend
```

### 2. 下载依赖

```bash
go mod download
```

### 3. 配置文件

后端默认读取 `config.yaml`，首次运行可使用默认配置：

```yaml
server:
  port: 8888
  mode: debug          # debug | release
  read_timeout: 60
  write_timeout: 60

database:
  path: ./cloudpivot.db

jwt:
  key_dir: ./runtime/keys

monitor:
  interval: 60         # 监控采集间隔（秒）
```

### 4. 启动后端

```bash
go run cmd/agent/main.go
```

启动成功后将看到以下日志：

```
[云枢面板] 配置加载成功，端口: 8888, 模式: debug
[云枢面板] 数据库初始化成功: ./cloudpivot.db
[云枢面板] JWT RSA 密钥对就绪
[云枢面板] 计划任务引擎已启动
[云枢面板] 告警引擎已启动
[云枢面板] 服务启动，监听 :8888
```

后端 API 地址：`http://localhost:8888/api/v1`

健康检查：`http://localhost:8888/health`

## 前端启动

### 1. 进入前端目录

```bash
cd frontend
```

### 2. 安装依赖

```bash
npm install
```

### 3. 启动开发服务器

```bash
npm run dev
```

启动成功后访问：`http://localhost:5173`

前端开发服务器已配置代理，将 `/api` 请求转发到后端 `http://localhost:8888`。

## 默认账号密码

| 项目 | 值 |
|------|-----|
| 用户名 | `admin` |
| 密码 | `admin123` |

> **安全提醒**：首次登录后请立即修改默认密码，并在系统设置中启用 TOTP 二次验证。

## 生产部署

### 使用 install.sh 一键部署

项目提供 `install.sh` 脚本用于在 Linux 服务器上一键部署：

```bash
# 下载并执行安装脚本
chmod +x install.sh
./install.sh
```

安装脚本将自动完成以下操作：

1. 检查系统环境（Go、Node、Docker）
2. 编译后端二进制文件
3. 构建前端静态资源
4. 创建运行目录和配置文件
5. 注册 systemd 服务（`cloudpivot.service`）
6. 启动服务并设置开机自启

### 服务管理

```bash
# 查看服务状态
systemctl status cloudpivot

# 启动/停止/重启
systemctl start cloudpivot
systemctl stop cloudpivot
systemctl restart cloudpivot

# 查看日志
journalctl -u cloudpivot -f
```

### 手动部署

如果需要手动部署，步骤如下：

```bash
# 1. 编译后端
cd backend
go build -o cloudpivot cmd/agent/main.go

# 2. 构建前端
cd ../frontend
npm run build
# 产物在 dist/ 目录

# 3. 部署到服务器
# 将后端二进制、前端 dist、config.yaml 上传到服务器
# 配置 Nginx 反向代理到后端 8888 端口
```

## 常见问题

### Q: 后端启动报错 "端口已被占用"

A: 修改 `config.yaml` 中的 `server.port` 为其他端口，同时更新前端 `vite.config.ts` 中的代理目标端口。

### Q: 前端页面空白或 API 请求 404

A: 确认后端已启动且端口正确。检查前端 `vite.config.ts` 中代理配置是否指向 `http://localhost:8888`。

### Q: Docker 相关功能不可用

A: 确认 Docker 已安装并启动，且当前用户有 Docker 访问权限（将用户加入 `docker` 组）。Windows 环境需安装 Docker Desktop。

### Q: 数据库文件在哪里？

A: 默认位于 `backend/cloudpivot.db`（SQLite）。生产环境建议定期备份此文件。

### Q: 如何重置管理员密码？

A: 删除数据库文件后重启服务，系统将重新初始化默认账号 `admin/admin123`。或使用 Web 终端直接操作 SQLite 数据库更新密码哈希。

### Q: JWT 密钥丢失会怎样？

A: JWT RSA 密钥对存储在 `runtime/keys/` 目录。丢失后所有已签发的 Token 将失效，用户需要重新登录。建议定期备份该目录。

### Q: 监控数据不更新？

A: 检查 `config.yaml` 中 `monitor.interval` 配置（默认 60 秒）。确认后端服务正常运行，监控采集器在服务启动时自动启动。
