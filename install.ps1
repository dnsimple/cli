#Requires -Version 5.1
<#
.SYNOPSIS
    Installs the DNSimple CLI for Windows.

.DESCRIPTION
    Downloads the appropriate DNSimple CLI binary for your architecture (arm64 or amd64),
    extracts it, and adds the install directory to the current user's PATH.
    By default, fetches the latest release from GitHub. Use -Version to pin a specific version.

.PARAMETER InstallDir
    Directory to install the CLI into.
    Defaults to "%LOCALAPPDATA%\DNSimple\bin"

.PARAMETER Version
    Specific version to install, e.g. "0.5.2". Omit to install the latest release.

.EXAMPLE
    .\install.ps1

.EXAMPLE
    .\install.ps1 -Version "0.5.2"

.EXAMPLE
    .\install.ps1 -InstallDir "C:\Tools\dnsimple"

.EXAMPLE
    irm "https://your-host.com/install.ps1" | iex

.EXAMPLE
    $s = irm "https://your-host.com/install.ps1"
    Invoke-Expression "& { $s } -Version '0.5.2' -InstallDir 'C:\Tools\dnsimple'"
#>

[CmdletBinding()]
param (
    [string]$InstallDir = "$env:LOCALAPPDATA\DNSimple\bin",
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"

$Repo   = "dnsimple/cli"
$ExeName = "dnsimple.exe"

function Write-Step {
    param([string]$Message)
    Write-Host "  --> $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "  [OK] $Message" -ForegroundColor Green
}

function Write-Fail {
    param([string]$Message)
    Write-Host "  [ERROR] $Message" -ForegroundColor Red
}

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12 -bor [Net.SecurityProtocolType]::Tls13

# -- Resolve version ----------------------------------------------------------

if ($Version -eq "") {
    Write-Step "Fetching latest release version from GitHub..."
    try {
        $releaseUrl = "https://api.github.com/repos/$Repo/releases/latest"
        $release = Invoke-WebRequest -Uri $releaseUrl -UseBasicParsing -Headers @{ "User-Agent" = "DNSimple-Installer" }
        $releaseJson = $release.Content | ConvertFrom-Json
        $Version = $releaseJson.tag_name -replace "^v", ""
        Write-Success "Latest version: $Version"
    } catch {
        Write-Fail "Could not fetch latest release from GitHub: $_"
        exit 1
    }
} else {
    $Version = $Version -replace "^v", ""
    Write-Success "Using specified version: $Version"
}

Write-Host ""
Write-Host "DNSimple CLI Installer v$Version" -ForegroundColor White
Write-Host "---------------------------------" -ForegroundColor DarkGray

# -- Detect architecture ------------------------------------------------------

Write-Step "Detecting system architecture..."

$arch = $env:PROCESSOR_ARCHITECTURE

if ($arch -eq "ARM64") {
    $ZipName = "dnsimple_${Version}_windows_arm64.zip"
    Write-Success "Detected arm64"
} elseif ($arch -eq "AMD64" -or $arch -eq "x86_64") {
    $ZipName = "dnsimple_${Version}_windows_amd64.zip"
    Write-Success "Detected amd64"
} else {
    Write-Fail "Unsupported architecture: $arch"
    exit 1
}

$DownloadUrl = "https://github.com/$Repo/releases/download/v$Version/$ZipName"

# -- Download -----------------------------------------------------------------

$TempDir = Join-Path $env:TEMP ("dnsimple_install_" + [System.IO.Path]::GetRandomFileName())
$ZipPath = Join-Path $TempDir $ZipName

Write-Step "Creating temp directory: $TempDir"
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

Write-Step "Downloading $DownloadUrl ..."
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -UseBasicParsing
    Write-Success "Download complete"
} catch {
    Write-Fail "Download failed: $_"
    Write-Fail "Check that version v$Version exists at https://github.com/$Repo/releases"
    Remove-Item $TempDir -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
}

# -- Extract ------------------------------------------------------------------

Write-Step "Extracting archive..."
try {
    Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force
    Write-Success "Extraction complete"
} catch {
    Write-Fail "Extraction failed: $_"
    Remove-Item $TempDir -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
}

$ExePath = Get-ChildItem -Path $TempDir -Recurse -Filter $ExeName | Select-Object -First 1

if (-not $ExePath) {
    Write-Fail "$ExeName not found in the downloaded archive."
    Remove-Item $TempDir -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
}

# -- Install ------------------------------------------------------------------

Write-Step "Installing to $InstallDir ..."
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$Destination = Join-Path $InstallDir $ExeName
Copy-Item -Path $ExePath.FullName -Destination $Destination -Force
Write-Success "Binary copied to $Destination"

# -- PATH ---------------------------------------------------------------------

Write-Step "Updating PATH for current user..."

$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")

$pathParts = $UserPath -split ";"
$alreadyInPath = $false
foreach ($part in $pathParts) {
    if ($part -eq $InstallDir) {
        $alreadyInPath = $true
        break
    }
}

if ($alreadyInPath) {
    Write-Success "PATH already contains $InstallDir -- no changes needed"
} else {
    $NewPath = ($UserPath.TrimEnd(";") + ";" + $InstallDir).TrimStart(";")
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    $env:PATH = $env:PATH.TrimEnd(";") + ";" + $InstallDir
    Write-Success "Added $InstallDir to user PATH"
}

# -- Cleanup ------------------------------------------------------------------

Write-Step "Cleaning up temporary files..."
Remove-Item $TempDir -Recurse -Force -ErrorAction SilentlyContinue
Write-Success "Cleanup complete"

# -- Verify -------------------------------------------------------------------

Write-Host ""
Write-Host "Verifying installation..." -ForegroundColor White

try {
    $cliVersion = & $Destination --version 2>&1
    Write-Host "  $cliVersion" -ForegroundColor White
} catch {
    Write-Host "  Could not auto-verify -- run dnsimple --version in a new terminal" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Installation complete!" -ForegroundColor Green
Write-Host "Open a new terminal and run: dnsimple --help" -ForegroundColor White
Write-Host ""
