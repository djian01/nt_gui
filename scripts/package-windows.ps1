param(
    [string]$Version = '',
    [switch]$Help
)
$ErrorActionPreference = 'Stop'
if ($Help) {
    Write-Host 'Usage: .\scripts\package-windows.ps1 [-Version major.minor.patch]'
    exit 0
}
if ($env:OS -ne 'Windows_NT') { throw 'Build Windows packages on Windows.' }
$Root = Split-Path -Parent $PSScriptRoot
if (!$Version) {
    $Plist = Get-Content (Join-Path $Root 'desktop/build/darwin/Info.plist') -Raw
    $Version = [regex]::Match($Plist, '<key>CFBundleShortVersionString</key><string>([^<]+)</string>').Groups[1].Value
}
if ($Version -notmatch '^\d+\.\d+\.\d+$') { throw 'Version must be major.minor.patch.' }
foreach ($Tool in @('go', 'npm.cmd', 'makensis')) {
    if (!(Get-Command $Tool -ErrorAction SilentlyContinue)) { throw "Missing required tool: $Tool (see BUILD_AND_INSTALL.md)." }
}
function Invoke-Checked {
    param([string]$Command, [string[]]$Arguments)
    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Command failed with exit code $LASTEXITCODE." }
}
$Arch = (& go env GOARCH).Trim()
$HostArch = (& go env GOHOSTARCH).Trim()
if ((& go env GOOS).Trim() -ne 'windows' -or $Arch -ne $HostArch -or $Arch -notin @('amd64', 'arm64')) {
    throw 'Use native Windows amd64 or arm64 Go (unset GOOS/GOARCH overrides).'
}
$OutputDir = Join-Path $Root 'installation_package'
New-Item -ItemType Directory -Force $OutputDir | Out-Null
$Work = Join-Path ([IO.Path]::GetTempPath()) ([guid]::NewGuid().ToString())
New-Item -ItemType Directory $Work | Out-Null
Push-Location (Join-Path $Root 'desktop')
try {
    Invoke-Checked 'npm.cmd' @('--prefix', 'frontend', 'ci')
    Invoke-Checked 'go' @('mod', 'download')
    Invoke-Checked 'npm.cmd' @('--prefix', 'frontend', 'run', 'build')
    Invoke-Checked 'go' @('run', 'github.com/akavel/rsrc@v0.10.2', '-arch', $Arch, '-ico', 'build/icons/net-test.ico', '-o', "rsrc_windows_$Arch.syso")
    Invoke-Checked 'go' @('build', '-tags', 'production', '-ldflags', '-H windowsgui', '-o', (Join-Path $Work 'net-test.exe'), '.')
    $Package = Join-Path $OutputDir "NET-Test-$Version-windows-$Arch-Setup.exe"
    Invoke-Checked 'makensis' @("/DVERSION=$Version", "/DARCH=$Arch", "/DSOURCE_DIR=$Work", "/DREPO_DIR=$Root", "/DOUTPUT_FILE=$Work\Setup.exe", (Join-Path $PSScriptRoot 'windows-installer.nsi'))
    Move-Item -Force (Join-Path $Work 'Setup.exe') $Package
    Write-Host "Created: $Package"
    Write-Host 'This installer is unsigned. WebView2 Runtime must be installed on the destination PC.'
} finally {
    Pop-Location
    Remove-Item -Recurse -Force $Work
}
