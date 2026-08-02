Write-Host "=== Deploying Daemon Core Background Service ===" -ForegroundColor Magenta

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
if (-not $scriptDir) { $scriptDir = $PSScriptRoot }
if (-not $scriptDir) { $scriptDir = "." }
$binaryPath = Join-Path $scriptDir "daemon\daemon.exe"

if (-not (Test-Path $binaryPath)) {
    Write-Error "daemon.exe was not found in daemon/ directory! Please run compilation first."
    exit 1
}

Write-Host "Found Daemon executable at: $binaryPath"

# Stop existing running instances to prevent file locking
Write-Host "Checking for running daemon.exe background processes..."
$running = Get-Process -Name "daemon" -ErrorAction SilentlyContinue
if ($running) {
    Write-Host "Stopping existing Daemon Core background instance..." -ForegroundColor Yellow
    Stop-Process -Name "daemon" -Force
    Start-Sleep -Seconds 1
}

# Start daemon as background process running mission-control Web Server (port 5000)
Write-Host "Launching Daemon Core (mission-control Web Server) in background..." -ForegroundColor Green
$token = (& $binaryPath token | Select-String "Token:" | ForEach-Object { $_.Line -replace "Token:\s+", "" }).Trim()
$env:DAEMON_PASSWORD = $token
$process = Start-Process -FilePath $binaryPath -ArgumentList "dashboard" -NoNewWindow -PassThru

# Write process info
$procId = $process.Id
Write-Host "Daemon Core background service successfully launched with PID: $procId" -ForegroundColor Green
Write-Host "Mission Control dashboard is active at: http://localhost:5000" -ForegroundColor Green

# Add a permanent User command alias for "daemon" pointing to daemon.exe
Write-Host "Setting temporary session command alias for 'daemon'..."
Set-Alias -Name "daemon" -Value $binaryPath

Write-Host "✅ Deployment Completed successfully!" -ForegroundColor Green
