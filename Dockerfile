# ============================================================
# Stage 1: Build frontend
# ============================================================
FROM node:22-alpine AS frontend
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# ============================================================
# Stage 2: Build Go binary
# ============================================================
FROM golang:1.23-alpine AS builder
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
COPY --from=frontend /src/web/dist ./web/dist

RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w \
      -X 'main.Version=${VERSION}' \
      -X 'main.BuildTime=${BUILD_TIME}' \
      -X 'main.GitCommit=${GIT_COMMIT}' \
      -X 'mo/internal/handler.Version=${VERSION}'" \
    -trimpath -o myblog .

# ============================================================
# Stage 3: Minimal runtime
# ============================================================
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata curl
WORKDIR /data
COPY --from=builder /src/myblog /usr/local/bin/myblog

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

EXPOSE 8080
VOLUME ["/data"]
ENV BLOG_CONFIG=/data/config.yaml
ENTRYPOINT ["myblog", "-c", "/data/config.yaml"]
