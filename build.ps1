# One-command build for Windows. Uses system Go, or downloads it into .\.toolchain.
$ErrorActionPreference = "Stop"
$GoVersion = "1.26.4"
Set-Location $PSScriptRoot

$go = (Get-Command go -ErrorAction SilentlyContinue).Source
if (-not $go) {
    $arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
    $go = Join-Path $PSScriptRoot ".toolchain\go\bin\go.exe"
    if (-not (Test-Path $go)) {
        $url = "https://go.dev/dl/go$GoVersion.windows-$arch.zip"
        Write-Host "Go not found; downloading $url"
        New-Item -ItemType Directory -Force .toolchain | Out-Null
        $zip = Join-Path $PSScriptRoot ".toolchain\go.zip"
        Invoke-WebRequest $url -OutFile $zip
        Expand-Archive $zip -DestinationPath .toolchain -Force
        Remove-Item $zip
    }
}

$version = (git describe --tags --always --dirty 2>$null)
if (-not $version) { $version = "dev" }
$env:CGO_ENABLED = "0"
& $go build -trimpath -ldflags "-s -w -X main.version=$version" -o md2html.exe .
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "Built $PSScriptRoot\md2html.exe ($version)"
