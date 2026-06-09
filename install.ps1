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
        # Fallback
        $tag = "v1.0.0"
    }
}

if (-not $tag) {
    Write-ErrorMsg "Failed to retrieve the latest version tag from GitHub."
}

Write-Info "Latest version is $tag"

# 3. Download the binary
$binaryName = "nept-windows-$arch.exe"
$downloadUrl = "https://github.com/NEPT-CLOUD/nept-cli-go/releases/download/$tag/$binaryName"
$installDir = Join-Path $env:USERPROFILE ".nept\bin"
$destPath = Join-Path $installDir "nept.exe"

if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
}

Write-Info "Downloading binary from: $downloadUrl"
try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $destPath -UseBasicParsing
} catch {
    Write-ErrorMsg "Failed to download $downloadUrl. Details: $_"
}

# 4. Update PATH if necessary
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

# 5. Install the skill folder to the host
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

Write-Success "Successfully installed nept to $destPath"
