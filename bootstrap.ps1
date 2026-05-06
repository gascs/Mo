# ============================================================
#  Mo Blog - One-Click Bootstrap for Windows
#  Usage:
#    .\bootstrap.ps1              构建 + 启动
#    .\bootstrap.ps1 -BuildOnly   仅构建
#    .\bootstrap.ps1 -Docker      Docker 构建 + 启动
# ============================================================

param(
    [switch]$BuildOnly,
    [switch]$Docker,
    [switch]$Help
)

if ($Help) {
    Write-Host @"
Mo Blog Bootstrap - Windows

Usage: .\bootstrap.ps1 [选项]

  (无参数)       构建并启动博客
  -BuildOnly    仅构建，不启动
  -Docker       使用 Docker 构建和运行
  -Help         显示帮助

"@
    exit 0
}

$ErrorActionPreference = "Stop"

$APP       = "myblog.exe"
$GOPROXY   = if ($env:GOPROXY) { $env:GOPROXY } else { "https://goproxy.cn,direct" }

# Version
$VERSION   = if ($env:VERSION) { $env:VERSION } else {
    try { git describe --tags --always --dirty 2>$null } catch { "dev" }
}
$BUILD_TS  = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$GIT_REV   = try { git rev-parse --short HEAD 2>$null } catch { "unknown" }

function Write-Step($msg) { Write-Host "[*] $msg" -ForegroundColor Green }
function Write-Info($msg) { Write-Host "    $msg" }
function Write-Warn($msg) { Write-Host "[!] $msg" -ForegroundColor Yellow }
function Write-OK($msg)   { Write-Host "[v] $msg" -ForegroundColor Cyan }

# ============================================================
#  Docker path
# ============================================================
if ($Docker) {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Warn "Docker not found. Install Docker Desktop first."
        Write-Info "  https://docs.docker.com/desktop/setup/install/windows-install/"
        exit 1
    }
    Write-Step "Building with Docker..."
    docker build `
        --build-arg "VERSION=$VERSION" `
        --build-arg "BUILD_TIME=$BUILD_TS" `
        --build-arg "GIT_COMMIT=$GIT_REV" `
        -t myblog:latest .
    Write-OK "Docker image built: myblog:latest"

    Write-Step "Starting container..."
    New-Item -ItemType Directory -Force -Path data | Out-Null
    docker rm -f myblog 2>$null
    docker run -d --name myblog -p 8080:8080 -v "$(Get-Location)/data:/data" myblog:latest
    Write-OK "Blog is running at http://localhost:8080"
    Write-Host ""
    Write-Host "  Stop:   docker stop myblog"
    Write-Host "  Logs:   docker logs -f myblog"
    exit 0
}

# ============================================================
#  Banner
# ============================================================
Write-Host ""
Write-Host "  Mo Blog Bootstrap v$VERSION" -ForegroundColor Cyan
Write-Host ""

# ============================================================
#  Check / Install Go
# ============================================================
Write-Step "Checking Go..."

$GO_INSTALLED = Get-Command go -ErrorAction SilentlyContinue
if ($GO_INSTALLED) {
    $goVer = (go version | Select-String -Pattern 'go(\d+\.\d+)').Matches.Groups[1].Value
    Write-OK "Go $goVer"
} else {
    Write-Warn "Go not found. Installing via winget..."
    $winget = Get-Command winget -ErrorAction SilentlyContinue
    if ($winget) {
        winget install --id GoLang.Go --silent --accept-package-agreements
    } else {
        Write-Warn "winget not available. Please install Go manually:"
        Write-Info "  https://go.dev/dl/"
        exit 1
    }
    # Refresh PATH
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
    Write-OK "Go: $(go version)"
}

# Set GOPROXY
& go env -w GOPROXY="$GOPROXY" 2>$null

# ============================================================
#  Check / Install Node.js
# ============================================================
Write-Step "Checking Node.js..."

$NODE_INSTALLED = Get-Command node -ErrorAction SilentlyContinue
if ($NODE_INSTALLED) {
    $nodeVer = node --version
    Write-OK "Node.js $nodeVer"
} else {
    Write-Warn "Node.js not found. Installing via winget..."
    $winget = Get-Command winget -ErrorAction SilentlyContinue
    if ($winget) {
        winget install --id OpenJS.NodeJS.LTS --silent --accept-package-agreements
    } else {
        Write-Warn "winget not available. Please install Node.js manually:"
        Write-Info "  https://nodejs.org/"
        exit 1
    }
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
    Write-OK "Node.js: $(node --version)"
}

# ============================================================
#  Build Frontend
# ============================================================
Write-Step "[1/3] Building frontend..."

Push-Location web
if (-not (Test-Path node_modules)) {
    npm install
} else {
    npm install 2>$null
}
npm run build
Pop-Location

Write-OK "Frontend built -> web\dist\"

# ============================================================
#  Build Go Binary
# ============================================================
Write-Step "[2/3] Building Go binary..."

$LDFLAGS = "-s -w " +
    "-X 'main.Version=$VERSION' " +
    "-X 'main.BuildTime=$BUILD_TS' " +
    "-X 'main.GitCommit=$GIT_REV' " +
    "-X 'mo/internal/handler.Version=$VERSION'"

$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
$env:GOPROXY = $GOPROXY

& go build -ldflags $LDFLAGS -trimpath -o $APP .

$binSize = (Get-Item $APP).Length / 1MB
Write-OK "Binary: $APP ($([math]::Round($binSize, 1)) MB)"
Write-Host ""

# Version check
& .\$APP -v 2>$null

# ============================================================
#  Start or Done
# ============================================================
if ($BuildOnly) {
    Write-Host ""
    Write-OK "Build complete: $APP"
    Write-Host "  Run: .\$APP"
} else {
    Write-Host ""
    Write-Host "  ==============================================" -ForegroundColor Green
    Write-Host "    Mo Blog is starting..." -ForegroundColor Green
    Write-Host "    http://localhost:8080" -ForegroundColor Cyan
    Write-Host "  ==============================================" -ForegroundColor Green
    Write-Host ""
    & .\$APP
}
