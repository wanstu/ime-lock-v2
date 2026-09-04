$ErrorActionPreference = "Stop"

Add-Type -AssemblyName System.Drawing

$root = Split-Path -Parent $PSScriptRoot
$windowSource = Join-Path $root "assets\icons\ime-lock-window.png"
$buildIcon = Join-Path $root "build\appicon.png"
$windowsIcon = Join-Path $root "build\windows\icon.ico"

if (-not (Test-Path $windowSource)) {
    throw "Missing icon source: $windowSource"
}

New-Item -ItemType Directory -Force -Path (Split-Path -Parent $buildIcon) | Out-Null

function Export-FittedSquareIcon {
    param(
        [Parameter(Mandatory = $true)][string]$Source,
        [Parameter(Mandatory = $true)][string]$Destination,
        [double]$Fill = 0.98
    )

    $sourceBitmap = [System.Drawing.Bitmap]::FromFile($Source)
    try {
        $canvasSize = 1024
        $targetSize = $canvasSize * $Fill
        $scale = [Math]::Min($targetSize / $sourceBitmap.Width, $targetSize / $sourceBitmap.Height)
        $drawWidth = [int][Math]::Round($sourceBitmap.Width * $scale)
        $drawHeight = [int][Math]::Round($sourceBitmap.Height * $scale)
        $drawX = [int][Math]::Round(($canvasSize - $drawWidth) / 2)
        $drawY = [int][Math]::Round(($canvasSize - $drawHeight) / 2)

        $output = New-Object System.Drawing.Bitmap($canvasSize, $canvasSize, [System.Drawing.Imaging.PixelFormat]::Format32bppArgb)
        try {
            $graphics = [System.Drawing.Graphics]::FromImage($output)
            try {
                $graphics.Clear([System.Drawing.Color]::Transparent)
                $graphics.CompositingQuality = [System.Drawing.Drawing2D.CompositingQuality]::HighQuality
                $graphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
                $graphics.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
                $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
                $graphics.DrawImage($sourceBitmap, $drawX, $drawY, $drawWidth, $drawHeight)
            }
            finally {
                $graphics.Dispose()
            }
            $output.Save($Destination, [System.Drawing.Imaging.ImageFormat]::Png)
        }
        finally {
            $output.Dispose()
        }
    }
    finally {
        $sourceBitmap.Dispose()
    }
}

# Match CodexPro+: normalize the artwork onto a 1024x1024 canvas and let it
# occupy almost the entire Windows icon slot.
Export-FittedSquareIcon -Source $windowSource -Destination $buildIcon -Fill 0.98

# Wails only regenerates the ICO reliably when the previous generated file is absent.
if (Test-Path $windowsIcon) {
    Remove-Item -Force $windowsIcon
}

Write-Host "Prepared IME Lock Windows app icon:"
Write-Host "  source: $windowSource"
Write-Host "  output: $buildIcon"
