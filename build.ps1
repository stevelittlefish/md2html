# One-command build for Windows. Uses system Go, or downloads it into .\.toolchain.
$ErrorActionPreference = "Stop"
$GoVersion = (Select-String -Path (Join-Path $PSScriptRoot "go.mod") -Pattern "^go (\S+)").Matches[0].Groups[1].Value
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

$version = if (Test-Path VERSION) { (Get-Content VERSION -Raw).Trim() } else { (git describe --tags --always --dirty 2>$null) }
if (-not $version) { $version = "dev" }
$env:CGO_ENABLED = "0"
$env:GOTOOLCHAIN = "local"
& $go build -trimpath -ldflags "-s -w -X main.version=$version" -o md2html.exe .
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "Built $PSScriptRoot\md2html.exe ($version)"
