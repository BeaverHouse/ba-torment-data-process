# Run in PowerShell

$ErrorActionPreference = "Stop"

Write-Host "Building batorment for Windows..." -ForegroundColor Cyan
Push-Location "$PSScriptRoot\.."
$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o batorment.exe .
Pop-Location
Write-Host "Build complete. Run '.\batorment.exe' from this directory (with .env file)." -ForegroundColor Green
