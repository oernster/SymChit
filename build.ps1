# Builds SymChit for Windows.
#
#   ./build.ps1          the gate, then the application
#   ./build.ps1 -Fast    the application alone, while working on it
#
# The gate runs first and cannot be skipped by any switch that ships: -Fast is
# for a working loop and says so in its output, so a release is never cut from
# a tree nobody verified.
#
# The version comes from VERSION and reaches the binary through -ldflags -X,
# which only writes to a var: against a const it silently does nothing, which is
# why main.appVersion is declared var.
param(
    [switch]$Fast
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

$version = (Get-Content (Join-Path $root 'VERSION') -Raw).Trim()
if (-not $version) { throw 'VERSION is empty' }

if ($Fast) {
    Write-Host "Building $version without the gate (working build, not a release)..."
} else {
    & (Join-Path $root 'test.ps1')
    if ($LASTEXITCODE -ne 0) { throw "the gate failed with exit code $LASTEXITCODE" }
    Write-Host "Building $version..."
}

wails build -ldflags "-X main.appVersion=$version"
if ($LASTEXITCODE -ne 0) { throw "wails build failed with exit code $LASTEXITCODE" }

$binary = Join-Path $root 'build\bin\SymChit.exe'
if (-not (Test-Path $binary)) { throw "the build reported success but $binary is not there" }
Write-Host ("Built {0} ({1:N0} bytes)" -f $binary, (Get-Item $binary).Length)
