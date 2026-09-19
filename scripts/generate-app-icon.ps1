$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$windowSource = Join-Path $root "assets\icons\ime-lock-window.png"

if (-not (Test-Path $windowSource)) {
    throw "Missing icon source: $windowSource"
}

Push-Location $root
try {
    $args = @(
        "run", "github.com/wanstu/wails-desktop-kit/cmd/desktopkit",
        "icon", "prepare-wails",
        "--input", $windowSource,
        "--desktop-dir", $root,
        "--normalize",
        "--canvas", "1024",
        "--fill", "0.98",
        "--trim-alpha=false"
    )
    & go @args
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
finally {
    Pop-Location
}

Write-Host "Prepared IME Lock Windows app icon with Wails Desktop Kit:"
Write-Host "  source: $windowSource"
Write-Host ("  output: " + (Join-Path $root "build\appicon.png"))
