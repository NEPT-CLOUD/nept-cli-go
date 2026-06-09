$ErrorActionPreference = 'Stop'

# Colors helper
function Write-Info ($msg) {
    Write-Host "[info] $msg" -ForegroundColor Cyan
}

function Write-Success ($msg) {
    Write-Host "[success] $msg" -ForegroundColor Green
}

function Write-Warn ($msg) {
    Write-Host "[warning] $msg" -ForegroundColor Yellow
}

function Write-ErrorMsg ($msg) {
    Write-Host "[error] $msg" -ForegroundColor Red
    exit 1
}

# Parse parameters
param (
    [switch]$Force
)

# 1. Detect Architecture
$arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64" -or $env:PROCESSOR_ARCHITEW6432 -eq "ARM64") {
    $arch = "arm64"
}

Write-Info "Detected platform: windows-$arch"

# 2. Resolve latest release tag
Write-Info "Fetching latest release version..."
$tag = ""
try {
    # Enable TLS 1.2/1.3 for API calls
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12 -bor [Net.SecurityProtocolType]::Tls13
    $response = Invoke-RestMethod -Uri "https://api.github.com/repos/NEPT-CLOUD/nept-cli-go/releases/latest" -UseBasicParsing
    $tag = $response.tag_name
} catch {
    Write-Warn "Failed to fetch latest tag from API. Trying redirect lookup..."
    try {
        $request = [System.Net.WebRequest]::Create("https://github.com/NEPT-CLOUD/nept-cli-go/releases/latest")
        $request.AllowAutoRedirect = $false
        $response = $request.GetResponse()
        $location = $response.Headers["Location"]
        if ($location -match 'tag/(.+)$') {
            $tag = $Matches[1].Trim()
        }
        $response.Close()
    } catch {
        Write-ErrorMsg "Failed to retrieve the latest version tag from GitHub."
    }
}

if (-not $tag) {
    Write-ErrorMsg "Failed to retrieve the latest version tag from GitHub."
}

Write-Info "Latest version is $tag"

# 3. Find current installed nept path
$neptPath = Get-Command nept -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source

# If not in path, check default locations
if (-not $neptPath) {
    $defaultPaths = @(
        (Join-Path $env:USERPROFILE ".nept\bin\nept.exe"),
        (Join-Path $env:ProgramFiles "nept\nept.exe"),
        ".\nept.exe"
    )
    foreach ($path in $defaultPaths) {
        if (Test-Path $path) {
            $neptPath = $path
            break
        }
    }
}

$currentVersion = ""
if ($neptPath) {
    Write-Info "Found existing Nept CLI at $neptPath"
    try {
        # Try JSON parsing
        $versionJson = & $neptPath version -f json | ConvertFrom-Json
        $currentVersion = $versionJson.version
    } catch {
        # Fallback to plain text
        try {
            $versionOutput = & $neptPath version
            if ($versionOutput -match 'nept version:\s+(\S+)') {
                $currentVersion = $Matches[1]
            }
        } catch {
            Write-Warn "Could not invoke version command on existing binary."
        }
    }
}

if ($currentVersion) {
    Write-Info "Currently installed version is: $currentVersion"
} else {
    Write-Warn "Could not determine currently installed version (or Nept CLI is not installed)."
    $currentVersion = "none"
}

# 4. Compare versions
$latestNorm = $tag -replace '^v', ''
$currentNorm = $currentVersion -replace '^v', ''

if ($currentNorm -eq $latestNorm -and -not $Force) {
    Write-Success "Nept CLI is already up-to-date (version $tag)."
    exit 0
}

if ($Force) {
    Write-Info "Force flag is set. Proceeding with upgrade to $tag..."
} else {
    Write-Info "Upgrading Nept CLI from $currentVersion to $tag..."
}

# 5. Determine install directory and file path
$binaryName = "nept-windows-$arch.exe"
$downloadUrl = "https://github.com/NEPT-CLOUD/nept-cli-go/releases/download/$tag/$binaryName"

if ($neptPath) {
    $destPath = $neptPath
    $installDir = Split-Path $destPath
} else {
    $installDir = Join-Path $env:USERPROFILE ".nept\bin"
    $destPath = Join-Path $installDir "nept.exe"
}

if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
}

Write-Info "Downloading binary from: $downloadUrl"
try {
    # If the process is running/locked, we might get an error when writing.
    # We download to a temp location first, then try to replace.
    $tempFile = [System.IO.Path]::GetTempFileName()
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing
    
    # Move/Rename file (and handle potential file lock)
    if (Test-Path $destPath) {
        Remove-Item -Force $destPath -ErrorAction SilentlyContinue
    }
    Move-Item -Path $tempFile -Destination $destPath -Force
} catch {
    Write-ErrorMsg "Failed to download/install binary. Details: $_"
}

# 6. Add to user PATH if it was a new installation and not already there
if (-not $neptPath) {
    Write-Info "Checking PATH environment variable..."
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $cleanUserPaths = $userPath -split ';' | ForEach-Object { $_.TrimEnd('\') }
    $cleanInstallDir = $installDir.TrimEnd('\')

    if ($cleanUserPaths -notcontains $cleanInstallDir) {
        Write-Info "Adding $installDir to the User PATH..."
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
        $env:Path += ";$installDir"
        Write-Warn "PATH has been updated. You may need to restart your terminal/IDE for the changes to take effect."
    }
}

# 7. Install the skill folder to the host
$skillDir = Join-Path $env:USERPROFILE ".nept\skill"
$destSkillPath = Join-Path $skillDir "SKILL.md"

if (-not (Test-Path $skillDir)) {
    New-Item -ItemType Directory -Force -Path $skillDir | Out-Null
}

$skillUrl = "https://raw.githubusercontent.com/NEPT-CLOUD/nept-cli-go/$tag/skill/SKILL.md"
Write-Info "Downloading skill file from: $skillUrl"
try {
    Invoke-WebRequest -Uri $skillUrl -OutFile $destSkillPath -UseBasicParsing
} catch {
    Write-Warn "Failed to download skill file from $skillUrl. Trying fallback to main..."
    try {
        $fallbackUrl = "https://raw.githubusercontent.com/NEPT-CLOUD/nept-cli-go/main/skill/SKILL.md"
        Invoke-WebRequest -Uri $fallbackUrl -OutFile $destSkillPath -UseBasicParsing
    } catch {
        Write-Warn "Failed to download skill file from fallback URL. Details: $_"
    }
}

Write-Success "Successfully upgraded nept to $destPath"
