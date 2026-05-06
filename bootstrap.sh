#!/bin/bash
# ============================================================
#  Mo Blog - One-Click Bootstrap
#  Usage:
#    ./bootstrap.sh             构建 + 启动
#    ./bootstrap.sh --build     仅构建
#    ./bootstrap.sh --docker    Docker 构建 + 启动
#    ./bootstrap.sh --deploy    构建 + 部署 (需设置 HOST)
# ============================================================

set -euo pipefail

# ---- Config ----
APP="myblog"
GO_MIN="1.22"
NODE_MIN="18"
GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"

# ---- Version (from git or env) ----
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
BUILD_TS="${BUILD_TS:-$(date -u '+%Y-%m-%dT%H:%M:%SZ')}"
GIT_REV="${GIT_REV:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}"

RED='\033[0;31m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'
log()    { echo -e "${GREEN}[*]${NC} $*"; }
info()   { echo -e "    $*"; }
warn()   { echo -e "${RED}[!]${NC} $*"; }
success(){ echo -e "${GREEN}[✓]${NC} $*"; }
banner() { echo -e "\n${CYAN}${BOLD}$*${NC}\n"; }

# ---- Parse args ----
DO_BUILD=true; DO_RUN=true; DO_DOCKER=false; DO_DEPLOY=false; HOST=""
for arg in "$@"; do
    case $arg in
        --build)            DO_RUN=false ;;
        --docker)           DO_DOCKER=true; DO_BUILD=false; DO_RUN=false ;;
        --deploy)           DO_DEPLOY=true; DO_RUN=false ;;
        --host=*)           HOST="${arg#*=}" ;;
        -h|--help)
            echo "Usage: $0 [--build | --docker | --deploy] [--host=IP]"
            echo "  (none)     Build and start the blog"
            echo "  --build    Build only (binary output)"
            echo "  --docker   Use Docker instead of local Go/Node"
            echo "  --deploy   Build and deploy to remote server"
            echo "  --host=IP  Server IP for --deploy"
            exit 0 ;;
    esac
done

# ============================================================
#  Step 0: Detect OS
# ============================================================
banner "Mo Blog - Bootstrap"

case "$(uname -s)" in
    Linux)  OS="linux" ;;
    Darwin) OS="darwin" ;;
    *)      die "Unsupported OS: $(uname -s)" ;;
esac

ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/' | sed 's/arm64/arm64/')
log "System: $OS / $ARCH"

if [ -f /etc/os-release ]; then
    . /etc/os-release
    DISTRO="${ID:-unknown}"
elif [ "$OS" = "darwin" ]; then
    DISTRO="macos"
else
    DISTRO="unknown"
fi
log "Distro: $DISTRO"

# ============================================================
#  Step 1: Docker path (if --docker)
# ============================================================
if $DO_DOCKER; then
    if ! command -v docker &>/dev/null; then
        warn "Docker not found. Install Docker first:"
        info "  https://docs.docker.com/get-docker/"
        exit 1
    fi
    log "Building with Docker..."
    docker build \
        --build-arg "VERSION=$VERSION" \
        --build-arg "BUILD_TIME=$BUILD_TS" \
        --build-arg "GIT_COMMIT=$GIT_REV" \
        -t myblog:latest .
    success "Docker image built: myblog:latest"

    log "Starting container on port 8080..."
    mkdir -p data
    docker rm -f myblog 2>/dev/null || true
    docker run -d --name myblog \
        -p 8080:8080 \
        -v "$(pwd)/data:/data" \
        myblog:latest
    success "Blog is running at http://localhost:8080"
    echo ""
    echo "  Stop:   docker stop myblog"
    echo "  Logs:   docker logs -f myblog"
    echo "  Remove: docker rm -f myblog"
    exit 0
fi

# ============================================================
#  Step 2: Install Go
# ============================================================
install_go() {
    local ver="${1:-1.23.4}"
    log "Installing Go $ver ..."

    case $DISTRO in
        ubuntu|debian)
            wget -q --show-progress \
                "https://golang.google.cn/dl/go${ver}.linux-${ARCH}.tar.gz" \
                -O /tmp/go.tar.gz 2>/dev/null || \
            wget -q --show-progress \
                "https://go.dev/dl/go${ver}.linux-${ARCH}.tar.gz" \
                -O /tmp/go.tar.gz
            tar -C /usr/local -xzf /tmp/go.tar.gz
            rm -f /tmp/go.tar.gz
            export PATH="/usr/local/go/bin:$PATH"
            echo 'export PATH="/usr/local/go/bin:$PATH"' | tee /etc/profile.d/go.sh >/dev/null 2>&1 || \
            echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.bashrc
            ;;
        macos)
            if command -v brew &>/dev/null; then
                brew install go
            else
                warn "Install Homebrew first: https://brew.sh"
                info "Or download Go from: https://go.dev/dl/"
                exit 1
            fi
            ;;
        centos|rhel|fedora|rocky|almalinux)
            wget -q --show-progress \
                "https://golang.google.cn/dl/go${ver}.linux-${ARCH}.tar.gz" \
                -O /tmp/go.tar.gz 2>/dev/null || \
            wget -q --show-progress \
                "https://go.dev/dl/go${ver}.linux-${ARCH}.tar.gz" \
                -O /tmp/go.tar.gz
            tar -C /usr/local -xzf /tmp/go.tar.gz
            rm -f /tmp/go.tar.gz
            export PATH="/usr/local/go/bin:$PATH"
            echo 'export PATH="/usr/local/go/bin:$PATH"' > /etc/profile.d/go.sh
            ;;
        *)
            warn "Unknown distro. Install Go manually: https://go.dev/dl/"
            info "Go $GO_MIN+ required"
            exit 1
            ;;
    esac
    success "Go: $(go version)"
}

if command -v go &>/dev/null; then
    GO_VER=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+' | head -1 || echo "0")
    if [ "$(printf '%s\n' "$GO_MIN" "$GO_VER" | sort -V | head -1)" = "$GO_MIN" ] && [ "$GO_VER" != "$GO_MIN" ]; then
        success "Go: $(go version)"
    else
        warn "Go $GO_VER < $GO_MIN required, upgrading..."
        install_go "1.23.4"
    fi
else
    install_go "1.23.4"
fi

# Always set GOPROXY
go env -w GOPROXY="$GOPROXY" 2>/dev/null || true

# ============================================================
#  Step 3: Install Node.js
# ============================================================
install_node() {
    local ver="${1:-22}"
    log "Installing Node.js $ver ..."

    case $DISTRO in
        ubuntu|debian)
            local node_file="node-v${ver}.12.0-linux-${ARCH}.tar.xz"
            wget -q --show-progress \
                "https://npmmirror.com/mirrors/node/latest-v${ver}.x/${node_file}" \
                -O /tmp/node.tar.xz 2>/dev/null || \
            wget -q --show-progress \
                "https://nodejs.org/dist/latest-v${ver}.x/${node_file}" \
                -O /tmp/node.tar.xz
            tar -C /usr/local --strip-components=1 -xJf /tmp/node.tar.xz
            rm -f /tmp/node.tar.xz
            ;;
        macos)
            if command -v brew &>/dev/null; then
                brew install node
            else
                warn "Install Homebrew first: https://brew.sh"
                exit 1
            fi
            ;;
        centos|rhel|fedora|rocky|almalinux)
            local node_file="node-v${ver}.12.0-linux-${ARCH}.tar.xz"
            wget -q --show-progress \
                "https://npmmirror.com/mirrors/node/latest-v${ver}.x/${node_file}" \
                -O /tmp/node.tar.xz 2>/dev/null || \
            wget -q --show-progress \
                "https://nodejs.org/dist/latest-v${ver}.x/${node_file}" \
                -O /tmp/node.tar.xz
            tar -C /usr/local --strip-components=1 -xJf /tmp/node.tar.xz
            rm -f /tmp/node.tar.xz
            ;;
        *)
            warn "Unknown distro. Install Node.js manually: https://nodejs.org/"
            exit 1
            ;;
    esac
    success "Node.js: $(node --version)"
}

if command -v node &>/dev/null; then
    NODE_VER=$(node --version | grep -oP 'v\K[0-9]+' | head -1 || echo "0")
    if [ "$NODE_VER" -ge "$NODE_MIN" ] 2>/dev/null; then
        success "Node.js: $(node --version)"
    else
        warn "Node.js v$NODE_VER < v$NODE_MIN required, upgrading..."
        install_node
    fi
else
    install_node
fi

# Configure npm mirror for China
npm config set registry https://registry.npmmirror.com 2>/dev/null || true

# ============================================================
#  Step 4: Build Frontend
# ============================================================
banner "Building Mo Blog $VERSION"

log "[1/3] Building frontend..."
cd web

if [ ! -d node_modules ]; then
    npm install --prefer-offline
else
    npm install --prefer-offline 2>/dev/null || true
fi

npm run build
cd ..
success "Frontend built -> web/dist/"

# ============================================================
#  Step 5: Build Go Binary
# ============================================================
log "[2/3] Building Go binary..."

LDFLAGS="-s -w \
  -X 'main.Version=$VERSION' \
  -X 'main.BuildTime=$BUILD_TS' \
  -X 'main.GitCommit=$GIT_REV' \
  -X 'mo/internal/handler.Version=$VERSION'"

GOOS=$OS GOARCH=$AMD64 CGO_ENABLED=0 \
    go build -ldflags "$LDFLAGS" -trimpath -o "$APP" .

BIN_SIZE=$(ls -lh "$APP" | awk '{print $5}')
success "Binary: $APP ($BIN_SIZE)"

# Show version
./"$APP" -v 2>/dev/null || true

# ============================================================
#  Step 6: Deploy or Run
# ============================================================
if $DO_DEPLOY; then
    HOST="${HOST:-}"
    if [ -z "$HOST" ]; then
        echo ""
        read -rp "  Server IP/domain: " HOST
    fi

    DEPLOY="${DEPLOY:-/opt/myblog}"
    SERVICE="${SERVICE:-myblog}"

    log "[3/3] Deploying to $HOST ..."
    ssh "$HOST" "mkdir -p $DEPLOY/backups"

    # Setup server if first time
    if ! ssh "$HOST" "test -f /etc/systemd/system/$SERVICE.service"; then
        log "First-time setup on server..."
        ssh "$HOST" "mkdir -p $DEPLOY/backups"
        scp deploy/myblog.service "$HOST:/tmp/$SERVICE.service"
        ssh "$HOST" "bash -s -- '$APP' '$DEPLOY' '$SERVICE' '$VERSION'" < scripts/setup-server.sh
    fi

    # Backup + upload + replace
    ssh "$HOST" "test -f $DEPLOY/$APP && cp $DEPLOY/$APP $DEPLOY/backups/$APP.\$(date +%Y%m%d%H%M%S) || true"
    scp "$APP" "$HOST:$DEPLOY/$APP.new"
    ssh "$HOST" "mv $DEPLOY/$APP.new $DEPLOY/$APP && chmod +x $DEPLOY/$APP"

    # Optional config upload
    if [ -f config.yaml ]; then
        scp config.yaml "$HOST:$DEPLOY/" 2>/dev/null || true
    fi

    # Restart + health check
    ssh "$HOST" "chown -R www-data:www-data $DEPLOY 2>/dev/null || true; systemctl restart $SERVICE" || true
    sleep 2

    if ssh "$HOST" "curl -sf http://localhost:8080/healthz" 2>/dev/null; then
        echo ""
        success "Deploy OK - $VERSION running at http://$HOST"
    else
        warn "Health check failed. Check: ssh $HOST systemctl status $SERVICE"
    fi

elif $DO_RUN; then
    log "[3/3] Starting Mo Blog..."
    echo ""
    echo -e "${GREEN}  ╔══════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}  ║     Mo Blog is starting...              ║${NC}"
    echo -e "${GREEN}  ║     打开 http://localhost:8080           ║${NC}"
    echo -e "${GREEN}  ╚══════════════════════════════════════════╝${NC}"
    echo ""
    ./"$APP"
else
    banner "Build complete: $APP"
    info "Run: ./$APP"
    info "Or:  ./$APP -v"
fi
