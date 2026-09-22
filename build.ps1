# Builds SymChit for Windows: the application, then the setup program that
# carries it.
#
#   ./build.ps1                 the gate, then the application and the setup program
#   ./build.ps1 -SkipInstaller  the application alone
#   ./build.ps1 -Fast           no gate, for a working loop
#
# Outputs:
#   build/bin/SymChit.exe            the application
#   dist-installer/SymChitSetup.exe  the setup program, carrying the application
#                                    as an embedded payload
#
# The gate runs first and cannot be skipped by any switch that ships: -Fast is
# for a working loop and says so in its output, so a release is never cut from
# a tree nobody verified.
#
# The version comes from VERSION and reaches the binary through -ldflags -X,
# which only writes to a var: against a const it silently does nothing, which is
# why main.appVersion is declared var.
param(
    [switch]$Fast,
    [switch]$SkipInstaller
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

if ($SkipInstaller) { return }

# The setup program is a second Wails application in the same module. It embeds
# the built application as a zip, so one file is the whole distribution.
# Section 4 of the GNU GPL asks that a copy of the licence be given to every
# recipient along with the program; section 6 carries that into conveying
# the built binary. So the licence travels in the payload and lands in the
# install folder beside the executable. The setup program shows it on screen as
# well; a file the user keeps is the part that answers the licence.
Copy-Item (Join-Path $root 'LICENSE') (Join-Path $root 'build\bin\LICENSE') -Force

Write-Host 'Packaging the application as the setup payload...'
$payload = Join-Path $root 'installer\payload.zip'
if (Test-Path $payload) { Remove-Item $payload -Force }
Compress-Archive -Path (Join-Path $root 'build\bin\*') -DestinationPath $payload

Write-Host 'Building the setup program...'
Push-Location (Join-Path $root 'installer')
try {
    wails build -ldflags "-X main.appVersion=$version"
    if ($LASTEXITCODE -ne 0) { throw "wails build failed with exit code $LASTEXITCODE" }
} finally {
    Pop-Location
}

Write-Host 'Collecting the setup program...'
$distDir = Join-Path $root 'dist-installer'
New-Item -ItemType Directory -Force -Path $distDir | Out-Null
$built = Join-Path $root 'installer\build\bin\SymChitSetup.exe'
if (-not (Test-Path $built)) { throw "the setup build reported success but $built is not there" }
$setup = Join-Path $distDir 'SymChitSetup.exe'
Copy-Item $built $setup -Force

# Put the empty-zip placeholder back, so `go build ./...` and the tests keep
# working without a full build and so a payload of megabytes never reaches a
# commit.
$empty = [byte[]](0x50, 0x4B, 0x05, 0x06) + (New-Object byte[] 18)
[System.IO.File]::WriteAllBytes($payload, $empty)

Write-Host ("Built {0} ({1:N0} bytes)" -f $setup, (Get-Item $setup).Length)
