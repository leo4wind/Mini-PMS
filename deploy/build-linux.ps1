# Cross-compile MiniPMS API for Linux, output to linux-<arch>/
# Usage: .\build-linux.ps1
#        .\build-linux.ps1 -Arch arm64

param(
    [ValidateSet("amd64", "arm64")]
    [string]$Arch = "amd64"
)

$ErrorActionPreference = "Stop"

$DeployRoot = $PSScriptRoot
$BackendRoot = Resolve-Path (Join-Path $DeployRoot "..\backend")
$OutDir = Join-Path $DeployRoot "linux-$Arch"
$OutBin = Join-Path $OutDir "minipms-server"
$ConfigSrc = Join-Path $BackendRoot "configs\config.yaml"
$ConfigDstDir = Join-Path $OutDir "configs"

New-Item -ItemType Directory -Force -Path $ConfigDstDir | Out-Null

$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = $Arch

Push-Location $BackendRoot
try {
    Write-Host "Building GOOS=linux GOARCH=$Arch CGO_ENABLED=0 ..."
    # Avoid -ldflags="..." in the command line: PowerShell splits it and mise sees -l.
    $ldflags = "-s -w"
    $miseArgs = @(
        "exec"
        "go"
        "--"
        "go"
        "build"
        "-ldflags=$ldflags"
        "-o"
        $OutBin
        "./cmd/server"
    )
    & mise @miseArgs
    if ($LASTEXITCODE -ne 0) {
        throw "go build failed with exit code $LASTEXITCODE"
    }
} finally {
    Pop-Location
}

Copy-Item $ConfigSrc (Join-Path $ConfigDstDir "config.yaml") -Force

$info = Get-Item $OutBin
$mb = [math]::Round($info.Length / 1MB, 2)
Write-Host "OK: $OutBin ($mb MB)"
Write-Host "Config: $(Join-Path $ConfigDstDir 'config.yaml')"
Write-Host "Done. See mvp/deploy/README.md for run steps."
