# ============================================
# 云枢面板 - 前端 Dockerfile
# 构建上下文：frontend/ 目录
# 构建命令：docker build -t cloudpivot-frontend .
# ============================================

# ---------- 构建阶段 ----------
FROM node:20-alpine AS builder

WORKDIR /build

# 先复制依赖文件，利用 Docker 层缓存加速构建
COPY package.json package-lock.json* ./

# 使用 npm ci 精确安装依赖（要求 package-lock.json 存在）
RUN npm ci

# 复制前端全部源码
COPY . .

# 执行 TypeScript 类型检查 + Vite 生产构建
RUN npm run build

# ---------- 运行阶段 ----------
FROM nginx:alpine

# 复制自定义 nginx 配置（反向代理 + SPA 路由 + WebSocket）
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 从构建阶段复制静态资源到 nginx 默认目录
COPY --from=builder /build/dist /usr/share/nginx/html

# 暴露 HTTP 端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1
