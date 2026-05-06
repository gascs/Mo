#!/bin/bash
# ============================================================
#  Mo Blog - First-time Server Setup
#
#  Usage (via Makefile):
#    make setup HOST=your-server.com
#
#  Or directly:
#    ssh root@host "bash -s -- myblog /opt/myblog myblog v1.0.0" < scripts/setup-server.sh
# ============================================================

set -euo pipefail

APP="${1:-myblog}"
DEPLOY="${2:-/opt/myblog}"
SERVICE="${3:-myblog}"
VERSION="${4:-dev}"

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${GREEN}[$(date '+%H:%M:%S')]${NC} $*"; }
warn() { echo -e "${RED}[WARN]${NC} $*"; }
die()  { echo -e "${RED}ERROR:${NC} $*" >&2; exit 1; }
info() { echo -e "${CYAN}       $*${NC}"; }

# ---- 1. Detect OS ----
if [ -f /etc/os-release ]; then
    . /etc/os-release
    DISTRO=$ID
else
    die "Cannot detect OS. Only Ubuntu/Debian/CentOS/Rocky are supported."
fi

ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')

echo ""
echo -e "${CYAN}  ╔══════════════════════════════════════════╗"
echo -e "  ║     Mo Blog Server Setup v${VERSION}          ║"
echo -e "  ╚══════════════════════════════════════════╝${NC}"
echo ""
log "Detected: $DISTRO ($ARCH)"

# ---- 2. System packages ----
log "=== [1/7] Installing system packages ==="

case $DISTRO in
    ubuntu|debian)
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y -qq curl wget git ca-certificates xz-utils
        ;;
    centos|rhel|fedora|rocky|almalinux)
        yum install -y curl wget git ca-certificates xz
        ;;
    *)
        die "Unsupported distro: $DISTRO"
        ;;
esac

# ---- 3. Install Go ----
log "=== [2/7] Installing Go ==="

GO_VER="1.23.4"

if command -v go &>/dev/null; then
    GO_CUR=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "0")
    log "Go found: $GO_CUR"
else
    log "Installing Go $GO_VER from golang.google.cn..."
    wget -q --show-progress \
        "https://golang.google.cn/dl/go${GO_VER}.linux-${ARCH}.tar.gz" \
        -O /tmp/go.tar.gz 2>/dev/null || \
    wget -q --show-progress \
        "https://go.dev/dl/go${GO_VER}.linux-${ARCH}.tar.gz" \
        -O /tmp/go.tar.gz
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm -f /tmp/go.tar.gz
    echo 'export PATH="/usr/local/go/bin:$PATH"' > /etc/profile.d/go.sh
    export PATH="/usr/local/go/bin:$PATH"
    log "Go installed: $(go version)"
fi

# Configure GOPROXY for China
go env -w GOPROXY=https://goproxy.cn,direct 2>/dev/null || true

# ---- 4. Install Node.js ----
log "=== [3/7] Installing Node.js ==="

NODE_VER=22

if command -v node &>/dev/null; then
    log "Node.js found: $(node --version)"
else
    case $DISTRO in
        ubuntu|debian)
            # Try npmmirror first, fall back to official
            NODE_FILE="node-v${NODE_VER}.12.0-linux-${ARCH}.tar.xz"
            log "Downloading Node.js $NODE_VER..."
            wget -q --show-progress \
                "https://npmmirror.com/mirrors/node/latest-v${NODE_VER}.x/${NODE_FILE}" \
                -O /tmp/node.tar.xz 2>/dev/null || \
            wget -q --show-progress \
                "https://nodejs.org/dist/latest-v${NODE_VER}.x/${NODE_FILE}" \
                -O /tmp/node.tar.xz
            tar -C /usr/local --strip-components=1 -xJf /tmp/node.tar.xz
            rm -f /tmp/node.tar.xz
            ;;
        *)
            log "Installing Node.js via NodeSource..."
            curl -fsSL https://rpm.nodesource.com/setup_${NODE_VER}.x | bash -
            yum install -y nodejs
            ;;
    esac
    log "Node.js installed: $(node --version)"
fi

# ---- 5. Create directories and user ----
log "=== [4/7] Creating directory structure ==="

mkdir -p "$DEPLOY"/{backups,uploads,certs}
log "Created: $DEPLOY"

log "=== [5/7] Creating www-data user ==="
if ! id www-data &>/dev/null; then
    useradd -r -s /sbin/nologin -M www-data 2>/dev/null || \
    useradd -r -s /usr/sbin/nologin -M www-data
    log "Created user: www-data"
else
    log "User www-data already exists"
fi
chown -R www-data:www-data "$DEPLOY"

# ---- 6. Install systemd unit ----
log "=== [6/7] Installing systemd service ==="

if [ -f /tmp/${SERVICE}.service ]; then
    cp /tmp/${SERVICE}.service /etc/systemd/system/${SERVICE}.service
    log "Installed from uploaded file"
else
    log "Creating systemd unit from template..."
    cat > /etc/systemd/system/${SERVICE}.service << UNITEOF
[Unit]
Description=Mo Blog
Documentation=https://github.com/gascs/mo
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=$DEPLOY
ExecStart=$DEPLOY/$APP -c $DEPLOY/config.yaml
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=5
LimitNOFILE=65536

# Security hardening
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
ReadWritePaths=$DEPLOY
ReadOnlyPaths=/etc/ssl/certs

[Install]
WantedBy=multi-user.target
UNITEOF
fi

systemctl daemon-reload
systemctl enable "$SERVICE"
log "Service enabled: $SERVICE"

# ---- 7. Firewall ----
log "=== [7/7] Configuring firewall ==="

if command -v ufw &>/dev/null; then
    ufw allow 80/tcp   >/dev/null 2>&1 || true
    ufw allow 443/tcp  >/dev/null 2>&1 || true
    ufw allow 8080/tcp >/dev/null 2>&1 || true
    log "ufw: ports 80, 443, 8080 allowed"
elif command -v firewall-cmd &>/dev/null; then
    firewall-cmd --permanent --add-service=http  >/dev/null 2>&1 || true
    firewall-cmd --permanent --add-service=https >/dev/null 2>&1 || true
    firewall-cmd --permanent --add-port=8080/tcp >/dev/null 2>&1 || true
    firewall-cmd --reload >/dev/null 2>&1 || true
    log "firewalld: http, https, 8080 allowed"
else
    warn "No firewall detected. Open ports 80, 443, 8080 manually if needed."
fi

# ---- Summary ----
echo ""
echo -e "${GREEN}==============================================${NC}"
echo -e "${GREEN}  Mo Blog server setup complete!${NC}"
echo ""
echo -e "  Deploy dir:  ${DEPLOY}"
echo -e "  Service:     ${SERVICE}"
echo -e "  Go version:  $(go version 2>/dev/null || echo 'not installed')"
echo -e "  Node.js:     $(node --version 2>/dev/null || echo 'not installed')"
echo ""
echo -e "  Next steps:"
echo -e "    1. On your dev machine, upload the binary:"
echo -e "       ${CYAN}make deploy HOST=$(curl -s ifconfig.me 2>/dev/null || echo '<server-ip>')${NC}"
echo -e "    2. Complete blog initialization:"
echo -e "       ${CYAN}http://$(curl -s ifconfig.me 2>/dev/null || echo '<server-ip>'):8080/setup${NC}"
echo -e "${GREEN}==============================================${NC}"
