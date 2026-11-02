# ============================================
# 犀牛面板（RhinoPanel） - 单镜像（前端嵌入后端二进制）
# 构建上下文：项目根目录
# 构建命令：docker build -t xiniuplan .
# 产物：一个容器同时服务页面（SPA）与 API/WebSocket
# ============================================

# ---------- 阶段1：前端构建 ----------
FROM node:22-alpine AS frontend-builder

WORKDIR /fe

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ .
RUN npm run build

# ---------- 阶段2：后端构建（嵌入前端产物） ----------
FROM golang:1.25-alpine AS backend-builder

WORKDIR /build

RUN apk add --no-cache git

COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 前端产物嵌入位置：backend/web/dist
COPY --from=frontend-builder /fe/dist ./web/dist

# 后端源码（在 dist 之后复制，避免覆盖嵌入产物）
COPY backend/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o cloudpivot ./cmd/agent/

# ---------- 阶段3：运行 ----------
FROM alpine:3.20

# ca-certificates：HTTPS 出站请求根证书；tzdata：时区
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=backend-builder /build/cloudpivot ./cloudpivot

# 默认配置（可被挂载的 /app/config.yaml 覆盖）
COPY backend/config.yaml ./config.yaml

EXPOSE 8888

VOLUME ["/app/data"]

ENTRYPOINT ["./cloudpivot"]
