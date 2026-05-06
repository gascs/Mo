# ============================================================
#  Mo Blog - Build & Deploy
#  Usage:
#    make bootstrap                一键安装依赖 + 构建 + 启动
#    make build-linux              构建 Linux 二进制
#    make docker-up                用 Docker 构建和启动
#    make deploy HOST=1.2.3.4      部署到服务器
#    make setup HOST=1.2.3.4       首次服务器初始化
#    make rollback HOST=1.2.3.4    回滚到上一个版本
#    make version                  显示当前版本信息
# ============================================================

APP       = myblog
WEB_DIR   = web
SRC       = .

# ---- Version (from git, override with VERSION=v1.0.0) ----
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TS  ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_REV   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# ---- Network (set GOPROXY=direct to bypass mirror) ----
GOPROXY   ?= https://goproxy.cn,direct

# ---- Linker flags ----
LDFLAGS   := -s -w \
             -X 'main.Version=$(VERSION)' \
             -X 'main.BuildTime=$(BUILD_TS)' \
             -X 'main.GitCommit=$(GIT_REV)' \
             -X 'mo/internal/handler.Version=$(VERSION)'

# ---- Deploy target (override defaults) ----
HOST      ?= your-server.com
DEPLOY    ?= /opt/myblog
SERVICE   ?= myblog

# ============================================================
#  Build
# ============================================================

all: build

.PHONY: frontend
frontend:
	@cd $(WEB_DIR) && (test -d node_modules || npm install) && npm run build

.PHONY: build build-linux build-darwin build-darwin-arm build-windows

build: build-linux

build-linux: frontend
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 GOPROXY=$(GOPROXY) \
		go build -ldflags "$(LDFLAGS)" -trimpath -o $(APP) $(SRC)

build-darwin: frontend
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 GOPROXY=$(GOPROXY) \
		go build -ldflags "$(LDFLAGS)" -trimpath -o $(APP) $(SRC)

build-darwin-arm: frontend
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 GOPROXY=$(GOPROXY) \
		go build -ldflags "$(LDFLAGS)" -trimpath -o $(APP) $(SRC)

build-windows: frontend
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GOPROXY=$(GOPROXY) \
		go build -ldflags "$(LDFLAGS)" -trimpath -o $(APP).exe $(SRC)

# ============================================================
#  Dev & Run
# ============================================================

.PHONY: dev run version

dev: frontend
	GOPROXY=$(GOPROXY) go run $(SRC)

run:
	./$(APP)

version:
	@echo "Version:   $(VERSION)"
	@echo "BuildTime: $(BUILD_TS)"
	@echo "GitCommit: $(GIT_REV)"

# ============================================================
#  One-click bootstrap (auto-install deps + build + run)
# ============================================================

.PHONY: bootstrap install

bootstrap:
	@chmod +x bootstrap.sh 2>/dev/null || true
	@./bootstrap.sh

install: frontend
	GOPROXY=$(GOPROXY) go build -ldflags "$(LDFLAGS)" -trimpath -o $(APP) $(SRC)
	@echo "Binary: $(APP)"
	@echo "Run: ./$(APP)"

# ============================================================
#  Docker
# ============================================================

.PHONY: docker-build docker-up docker-down docker-logs

DOCKER_IMAGE ?= myblog:latest

docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TS) \
		--build-arg GIT_COMMIT=$(GIT_REV) \
		-t $(DOCKER_IMAGE) .

docker-up: docker-build
	@mkdir -p data
	docker rm -f $(APP) 2>/dev/null || true
	docker run -d --name $(APP) \
		-p 8080:8080 \
		-v $$(pwd)/data:/data \
		$(DOCKER_IMAGE)
	@sleep 1
	@echo "Blog running at http://localhost:8080"
	@echo "Logs:  make docker-logs"
	@echo "Stop:  make docker-down"

docker-down:
	docker stop $(APP) 2>/dev/null || true
	docker rm $(APP) 2>/dev/null || true

docker-logs:
	docker logs -f $(APP)

# ============================================================
#  Deploy (update existing installation)
# ============================================================

.PHONY: deploy deploy-full

deploy: build-linux
	@echo "==> [1/5] Backup old binary on remote..."
	ssh $(HOST) "if [ -f $(DEPLOY)/$(APP) ]; then \
		cp $(DEPLOY)/$(APP) $(DEPLOY)/backups/$(APP).$$(date +%Y%m%d%H%M%S); \
		ls -t $(DEPLOY)/backups/$(APP).* 2>/dev/null | tail -n +8 | xargs -r rm -f; \
		echo '  Backup OK'; else echo '  No existing binary (first deploy)'; fi"
	@echo "==> [2/5] Upload new binary..."
	scp $(APP) $(HOST):$(DEPLOY)/$(APP).new
	@echo "==> [3/5] Atomically replace binary..."
	ssh $(HOST) "mv $(DEPLOY)/$(APP).new $(DEPLOY)/$(APP) && chmod +x $(DEPLOY)/$(APP)"
	@echo "==> [4/5] Restart service..."
	ssh $(HOST) "systemctl restart $(SERVICE)" || true
	@echo "==> [5/5] Health check..."
	@sleep 2
	@if ssh $(HOST) "curl -sf http://localhost:8080/healthz"; then \
		echo ""; \
		echo "  Deploy OK -- $(VERSION)"; \
	else \
		echo ""; \
		echo "  Deploy FAILED -- rolling back..."; \
		$(MAKE) rollback; \
		exit 1; \
	fi

deploy-full: build-linux
	@echo "==> [1/6] Backup on remote..."
	ssh $(HOST) "if [ -f $(DEPLOY)/$(APP) ]; then \
		cp $(DEPLOY)/$(APP) $(DEPLOY)/backups/$(APP).$$(date +%Y%m%d%H%M%S); \
		ls -t $(DEPLOY)/backups/$(APP).* 2>/dev/null | tail -n +8 | xargs -r rm -f; fi"
	@echo "==> [2/6] Backup config..."
	ssh $(HOST) "test -f $(DEPLOY)/config.yaml && cp $(DEPLOY)/config.yaml $(DEPLOY)/backups/config.yaml.$$(date +%Y%m%d%H%M%S) || true"
	@echo "==> [3/6] Upload binary + config..."
	scp $(APP) config.yaml $(HOST):$(DEPLOY)/
	@echo "==> [4/6] Set permissions..."
	ssh $(HOST) "chmod +x $(DEPLOY)/$(APP) && chown www-data:www-data $(DEPLOY)/*"
	@echo "==> [5/6] Restart service..."
	ssh $(HOST) "systemctl restart $(SERVICE)" || true
	@echo "==> [6/6] Health check..."
	@sleep 2
	@if ssh $(HOST) "curl -sf http://localhost:8080/healthz"; then \
		echo ""; \
		echo "  Deploy OK -- $(VERSION)"; \
	else \
		echo ""; \
		echo "  Deploy FAILED -- rolling back..."; \
		$(MAKE) rollback; \
		exit 1; \
	fi

# ============================================================
#  Rollback
# ============================================================

.PHONY: rollback

rollback:
	@echo "==> Rolling back to previous binary..."
	@ssh $(HOST) 'LATEST_BACKUP=$$(ls -t $(DEPLOY)/backups/$(APP).* 2>/dev/null | head -1); \
		if [ -n "$$LATEST_BACKUP" ]; then \
			cp $$LATEST_BACKUP $(DEPLOY)/$(APP) && \
			chmod +x $(DEPLOY)/$(APP) && \
			systemctl restart $(SERVICE) && \
			echo "  Rollback OK: $$LATEST_BACKUP"; \
		else \
			echo "  No backup found"; \
			exit 1; \
		fi'

# ============================================================
#  First-time server setup
# ============================================================

.PHONY: setup

setup: build-linux
	@test -n "$(HOST)" || (echo "ERROR: set HOST=your-server.com"; exit 1)
	@test -n "$(DEPLOY)" || (echo "ERROR: set DEPLOY=/opt/myblog"; exit 1)
	@echo "==> Copying systemd unit and setup script to server..."
	ssh $(HOST) "mkdir -p $(DEPLOY)/backups"
	scp deploy/$(SERVICE).service $(HOST):/tmp/$(SERVICE).service
	ssh $(HOST) "bash -s -- '$(APP)' '$(DEPLOY)' '$(SERVICE)' '$(VERSION)'" < scripts/setup-server.sh
	@echo ""
	@echo "==> Uploading initial binary and config..."
	scp $(APP) $(HOST):$(DEPLOY)/
	scp config.yaml $(HOST):$(DEPLOY)/ 2>/dev/null || echo "  (no config.yaml, will use setup wizard)"
	ssh $(HOST) "chmod +x $(DEPLOY)/$(APP) && chown -R www-data:www-data $(DEPLOY)"
	@echo "==> Starting service..."
	ssh $(HOST) "systemctl start $(SERVICE)" || true
	@sleep 2
	@if ssh $(HOST) "curl -sf http://localhost:8080/healthz"; then \
		echo ""; \
		echo "  Setup complete! Blog is running."; \
	else \
		echo ""; \
		echo "  Service may need initialization. Open http://$(HOST):8080/setup"; \
	fi

# ============================================================
#  Status
# ============================================================

.PHONY: status logs

status:
	ssh $(HOST) "systemctl status $(SERVICE) --no-pager" || true
	@echo ""
	@ssh $(HOST) "$(DEPLOY)/$(APP) -v 2>/dev/null || echo '(binary version unavailable)'"

logs:
	ssh $(HOST) "journalctl -u $(SERVICE) -n 50 --no-pager" || true

# ============================================================
#  Clean
# ============================================================

.PHONY: clean

clean:
	rm -f $(APP) $(APP).exe
	rm -rf $(WEB_DIR)/dist
