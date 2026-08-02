Write-Host "=== Daemon Global Installer ===" -ForegroundColor Magenta
Write-Host "Installing Daemon Engineering OS v1.0 globally..." -ForegroundColor Cyan

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
if (-not $scriptDir) { $scriptDir = "c:\Users\MAHESH\OneDrive\Desktop\Daemon CLI" }
$binaryPath = Join-Path $scriptDir "daemon\daemon.exe"

if (-not (Test-Path $binaryPath)) {
    Write-Error "daemon.exe not found. Please compile first: cd daemon && go build -o daemon.exe"
    exit 1
}

# 1. Install daemon.exe globally to WindowsApps so it's on PATH everywhere
$globalTarget = "$env:LOCALAPPDATA\Microsoft\WindowsApps\daemon.exe"
Write-Host "Installing binary to: $globalTarget"
Copy-Item -Path $binaryPath -Destination $globalTarget -Force
Write-Host "  [OK] daemon.exe installed globally" -ForegroundColor Green

# 2. Install VS Code Extension globally
$extensionSrc = Join-Path $scriptDir "vscode-extension"
$extensionDest = "$env:USERPROFILE\.vscode\extensions\daemon-vscode"

if (Test-Path $extensionSrc) {
    Write-Host "Installing VS Code Extension to: $extensionDest"
    Copy-Item -Path $extensionSrc -Destination $extensionDest -Recurse -Force
    Write-Host "  [OK] VS Code Extension installed globally" -ForegroundColor Green
}

# 3. Set DAEMON_PATH environment variable
[System.Environment]::SetEnvironmentVariable("DAEMON_PATH", $globalTarget, "User")
Write-Host "  [OK] DAEMON_PATH set to: $globalTarget" -ForegroundColor Green

# 4. Verify installation
Write-Host ""
Write-Host "Verifying installation..." -ForegroundColor Cyan
$secretToken = (& $globalTarget token | Select-String "Token:" | ForEach-Object { $_.Line -replace "Token:\s+", "" }).Trim()
if ($secretToken) {
    $env:DAEMON_PASSWORD = $secretToken
    & $globalTarget version
} else {
    & $globalTarget version
}
Write-Host ""
Write-Host "=== Installation Complete ===" -ForegroundColor Magenta
Write-Host "Run 'daemon --help' from any terminal directory." -ForegroundColor Cyan
Write-Host "Restart VS Code to activate the Daemon extension." -ForegroundColor Cyan
