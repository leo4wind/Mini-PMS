# Pack MiniPMS release artifacts into ../compose/release/ for Docker Compose.
# - Linux amd64 API binary
# - Frontend Vite dist -> release/web
#
# Usage (from anywhere):
#   mvp\deploy\pack-release.ps1
#   mvp\deploy\pack-release.ps1 -SkipFrontend
#   mvp\deploy\pack-release.ps1 -SkipBackend

param(
    [switch]$SkipFrontend,
    [switch]$SkipBackend
)

$ErrorActionPreference = "Stop"

$DeployRoot = $PSScriptRoot
$MvpRoot = Resolve-Path (Join-Path $DeployRoot "..")
$BackendRoot = Join-Path $MvpRoot "backend"
$FrontendRoot = Join-Path $MvpRoot "frontend"
$ReleaseDir = Join-Path $MvpRoot "compose\release"
$WebDir = Join-Path $ReleaseDir "web"
$LinuxBin = Join-Path $DeployRoot "linux-amd64\minipms-server"
$OutBin = Join-Path $ReleaseDir "minipms-server"

New-Item -ItemType Directory -Force -Path $ReleaseDir | Out-Null

# Sync DDL+seed into compose/initdb (Docker MySQL first-boot import)
$InitDbDir = Join-Path $MvpRoot "compose\initdb"
$SchemaSrc = Join-Path $MvpRoot "schema.sql"
New-Item -ItemType Directory -Force -Path $InitDbDir | Out-Null
Copy-Item $SchemaSrc (Join-Path $InitDbDir "01-schema.sql") -Force
Write-Host "Schema -> $(Join-Path $InitDbDir '01-schema.sql')"

# --- Backend ---
if (-not $SkipBackend) {
    Write-Host "==> Building Linux amd64 API..."
    & (Join-Path $DeployRoot "build-linux.ps1") -Arch amd64
    if ($LASTEXITCODE -ne 0 -and $null -ne $LASTEXITCODE) {
        throw "build-linux.ps1 failed"
    }
}

if (-not (Test-Path $LinuxBin)) {
    throw "Missing API binary: $LinuxBin (run without -SkipBackend, or build-linux.ps1 first)"
}
Copy-Item $LinuxBin $OutBin -Force
Write-Host "API -> $OutBin"

# --- Frontend ---
$DistDir = Join-Path $FrontendRoot "dist"
if (-not $SkipFrontend) {
    Write-Host "==> Building frontend (npm run build)..."
    Push-Location $FrontendRoot
    try {
        if (-not (Test-Path (Join-Path $FrontendRoot "node_modules"))) {
            Write-Host "npm install..."
            $npmInstall = @("exec", "node", "--", "npm", "install")
            & mise @npmInstall
            if ($LASTEXITCODE -ne 0) { throw "npm install failed" }
        }
        $npmBuild = @("exec", "node", "--", "npm", "run", "build")
        & mise @npmBuild
        if ($LASTEXITCODE -ne 0) { throw "npm run build failed" }
    } finally {
        Pop-Location
    }
}

if (-not (Test-Path (Join-Path $DistDir "index.html"))) {
    throw "Missing frontend dist: $DistDir (run without -SkipFrontend)"
}

if (Test-Path $WebDir) {
    Remove-Item $WebDir -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $WebDir | Out-Null
Copy-Item -Path (Join-Path $DistDir "*") -Destination $WebDir -Recurse -Force
if (-not (Test-Path (Join-Path $WebDir "index.html"))) {
    throw "compose/release/web/index.html missing after copy (expected Vite dist contents directly under web/)"
}
Write-Host "Web -> $WebDir"

$binSize = [math]::Round((Get-Item $OutBin).Length / 1MB, 2)
Write-Host ""
Write-Host "Release ready under mvp/compose/release/ (API ${binSize} MB + web/)"
Write-Host "Next:"
Write-Host "  cd mvp/compose"
Write-Host "  docker compose down -v"
Write-Host "  docker compose up -d --build"
Write-Host "Open: http://127.0.0.1/mini-pms/"
