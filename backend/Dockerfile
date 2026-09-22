# ---- 阶段1: 构建前端 ----
FROM node:20-alpine AS frontend-builder

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json* ./
RUN npm config set registry https://registry.npmmirror.com && npm install

COPY frontend .
RUN npm run build

# ---- 阶段2: 构建后端 ----
FROM golang:1.22-alpine AS backend-builder

WORKDIR /app

RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.cernet.edu.cn/alpine#g' /etc/apk/repositories && \
    apk add --no-cache gcc musl-dev

COPY backend/go.mod backend/go.sum ./
RUN go env -w  GOPROXY=https://goproxy.cn,direct && go mod download

COPY backend .
RUN CGO_ENABLED=0 GOOS=linux go build -o mcloud-server .

# ---- 阶段3: 运行 ----
FROM alpine:3.19

WORKDIR /app

RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.cernet.edu.cn/alpine#g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata nginx

# 复制前端构建产物
COPY --from=frontend-builder /app/dist /usr/share/nginx/html

# 复制后端二进制
COPY --from=backend-builder /app/mcloud-server .
COPY --from=backend-builder /app/.env.example .env.example

# 复制 nginx 配置与启动脚本
RUN rm -f /etc/nginx/http.d/default.conf /etc/nginx/conf.d/default.conf
COPY backend/nginx.conf /etc/nginx/http.d/default.conf
COPY backend/run.sh /app/run.sh
RUN chmod +x /app/run.sh

EXPOSE 5678

CMD ["/app/run.sh"]
