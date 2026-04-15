param(
    [string]$Version = "",
    [string]$InstallDir = "",
    [string]$BaseUrl = "",
    [switch]$Yes
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

$DefaultBaseUrl = "https://github.com/dnsimple/homebrew-tap/releases"
$BinaryName = "dnsimple.exe"

function Write-Info {
    param([string]$Message)
    Write-Host "> $Message"
}

function Write-Warn {
    param([string]$Message)
    Write-Host "! $Message" -ForegroundColor Yellow
}

function Fail {
    param([string]$Message)
    throw $Message
}

function Get-InstallDir {
    if ($InstallDir) {
        return $InstallDir
    }

    if ($env:DNSIMPLE_INSTALL) {
        return $env:DNSIMPLE_INSTALL
    }

    return Join-Path $HOME ".dnsimple"
}

function Get-BaseUrl {
    if ($BaseUrl) {
        return $BaseUrl.TrimEnd("/")
    }

    if ($env:DNSIMPLE_BASE_URL) {
        return $env:DNSIMPLE_BASE_URL.TrimEnd("/")
    }

    return $DefaultBaseUrl
}

function Get-Arch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()

    switch ($arch) {
        "x64" { return "amd64" }
        "arm64" { return "arm64" }
        default { Fail "Unsupported architecture: $arch" }
    }
}

function Download-File {
    param(
        [string]$Url,
        [string]$Destination
    )

    try {
        Invoke-WebRequest -Uri $Url -OutFile $Destination
    } catch {
        Fail "Download failed: $Url`n$($_.Exception.Message)"
    }
}

function Get-LatestVersion {
    param([string]$ResolvedBaseUrl)

    Write-Info "Fetching latest version..."

    $handler = [System.Net.Http.HttpClientHandler]::new()
    $handler.AllowAutoRedirect = $false
    $client = [System.Net.Http.HttpClient]::new($handler)

    try {
        $request = [System.Net.Http.HttpRequestMessage]::new([System.Net.Http.HttpMethod]::Head, "$ResolvedBaseUrl/latest")
        $response = $client.SendAsync($request).GetAwaiter().GetResult()

        if ($response.Headers.Location) {
            $location = $response.Headers.Location.ToString()
        } else {
            $location = $response.RequestMessage.RequestUri.ToString()
        }
    } finally {
        if ($null -ne $response) {
            $response.Dispose()
        }
        $request.Dispose()
        $client.Dispose()
        $handler.Dispose()
    }

    if ($location -match "/v?([0-9][^/]*)$") {
        return $Matches[1]
    }

    Fail "Failed to determine latest version. Please specify -Version 0.1.0."
}

function Resolve-Version {
    param([string]$ResolvedBaseUrl)

    if ($Version) {
        return $Version.TrimStart("v")
    }

    return Get-LatestVersion -ResolvedBaseUrl $ResolvedBaseUrl
}

function Verify-Checksum {
    param(
        [string]$FilePath,
        [string]$ChecksumsPath,
        [string]$ArchiveName
    )

    $actual = (Get-FileHash -Path $FilePath -Algorithm SHA256).Hash.ToLowerInvariant()
    $expected = $null

    foreach ($line in Get-Content -Path $ChecksumsPath) {
        if ($line -match "^(?<hash>[A-Fa-f0-9]+)\s+\*?(?<name>.+)$" -and $Matches["name"] -eq $ArchiveName) {
            $expected = $Matches["hash"].ToLowerInvariant()
            break
        }
    }

    if (-not $expected) {
        Write-Warn "No checksum found for $ArchiveName. Skipping verification."
        return
    }

    if ($actual -ne $expected) {
        Fail "Checksum verification failed for $ArchiveName."
    }
}

function Confirm-Install {
    param([string]$ResolvedVersion)

    if ($Yes.IsPresent) {
        return
    }

    if ([Console]::IsInputRedirected) {
        return
    }

    $answer = Read-Host "Install DNSimple CLI v$ResolvedVersion? [y/N]"
    if ($answer -notmatch "^(?i:y(?:es)?)$") {
        Fail "Installation cancelled."
    }
}

function Ensure-UserPath {
    param([string]$BinDir)

    $normalizedBinDir = $BinDir.TrimEnd("\")
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $entries = @()

    if ($userPath) {
        $entries = $userPath -split ";" | Where-Object { $_ }
        foreach ($entry in $entries) {
            if ($entry.TrimEnd("\") -ieq $normalizedBinDir) {
                if ($env:Path -notlike "*$BinDir*") {
                    $env:Path = "$BinDir;$env:Path"
                }
                return "already"
            }
        }
    }

    $newUserPath = if ($entries.Count -gt 0) {
        ($entries + $BinDir) -join ";"
    } else {
        $BinDir
    }

    try {
        [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
        $env:Path = "$BinDir;$env:Path"
        return "added"
    } catch {
        return "manual"
    }
}

function Show-Banner {
    param([string]$ResolvedVersion)

    Write-Host ""
    Write-Host "DNSimple CLI v$ResolvedVersion installed." -ForegroundColor Green
    Write-Host "Run dnsimple --help to get started."
    Write-Host ""
}

$resolvedBaseUrl = Get-BaseUrl
$resolvedVersion = Resolve-Version -ResolvedBaseUrl $resolvedBaseUrl
$arch = Get-Arch
$resolvedInstallDir = Get-InstallDir
$binDir = Join-Path $resolvedInstallDir "bin"
$archiveName = "dnsimple_${resolvedVersion}_windows_${arch}.zip"
$downloadUrl = "$resolvedBaseUrl/download/v$resolvedVersion/$archiveName"
$checksumsUrl = "$resolvedBaseUrl/download/v$resolvedVersion/checksums.txt"

Write-Host ""
Write-Info "Version:  $resolvedVersion"
Write-Info "Platform: windows/$arch"
Write-Info "Location: $binDir"
Write-Host ""

Confirm-Install -ResolvedVersion $resolvedVersion

$tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ("dnsimple-install-" + [Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmpDir | Out-Null

try {
    $archivePath = Join-Path $tmpDir $archiveName
    $checksumsPath = Join-Path $tmpDir "checksums.txt"
    $binaryPath = Join-Path $tmpDir $BinaryName
    $installedBinaryPath = Join-Path $binDir $BinaryName

    Write-Info "Downloading $archiveName..."
    Download-File -Url $downloadUrl -Destination $archivePath

    Write-Info "Verifying checksum..."
    Download-File -Url $checksumsUrl -Destination $checksumsPath
    Verify-Checksum -FilePath $archivePath -ChecksumsPath $checksumsPath -ArchiveName $archiveName

    Write-Info "Extracting..."
    New-Item -ItemType Directory -Force -Path $binDir | Out-Null
    Expand-Archive -Path $archivePath -DestinationPath $tmpDir -Force

    if (-not (Test-Path -Path $binaryPath)) {
        Fail "Extracted archive did not contain $BinaryName."
    }

    Move-Item -Path $binaryPath -Destination $installedBinaryPath -Force

    $pathResult = Ensure-UserPath -BinDir $binDir
    switch ($pathResult) {
        "already" { Write-Info "PATH already includes $binDir" }
        "added" { Write-Info "PATH updated for the current user" }
        default { Write-Warn "Add $binDir to your user PATH manually." }
    }

    Show-Banner -ResolvedVersion $resolvedVersion
} finally {
    Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
