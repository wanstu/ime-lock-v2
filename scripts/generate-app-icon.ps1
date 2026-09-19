$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$windowSource = Join-Path $root "assets\icons\ime-lock-window.png"
$buildIcon = Join-Path $root "build\appicon.png"
$windowsIcon = Join-Path $root "build\windows\icon.ico"

if (-not (Test-Path $windowSource)) {
    throw "Missing icon source: $windowSource"
}

New-Item -ItemType Directory -Force -Path (Split-Path -Parent $buildIcon) | Out-Null

Push-Location $root
try {
    go run github.com/wanstu/wails-desktop-kit/cmd/desktopkit icon --input $windowSource --output $buildIcon --canvas 1024 --fill 0.98 --trim-alpha=false
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}
finally {
    Pop-Location
}

# Wails only regenerates the ICO reliably when the previous generated file is absent.
if (Test-Path $windowsIcon) {
    Remove-Item -Force $windowsIcon
}

Write-Host "Prepared IME Lock Windows app icon with Wails Desktop Kit:"
Write-Host "  source: $windowSource"
Write-Host "  output: $buildIcon"
