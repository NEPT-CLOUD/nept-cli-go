[CmdletBinding()]
param(
    [switch]$Force
)

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

$installDir = Join-Path $env:USERPROFILE ".nept"
$binDir = Join-Path $installDir "bin"

$removedAny = $false

# 1. Remove binary and its folder
if (Test-Path $installDir) {
    Write-Info "Removing installation folder at: $installDir"
    try {
        Remove-Item -Path $installDir -Recurse -Force
        $removedAny = $true
    } catch {
        Write-Warn "Could not remove installation directory: $_"
    }
}

# 2. Clean up PATH
Write-Info "Checking PATH environment variable..."
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath) {
    $cleanInstallDir = $binDir.TrimEnd('\')
    $paths = $userPath -split ';'
    $newPaths = @()
    $pathRemoved = $false

    foreach ($p in $paths) {
        if ($p.Trim().TrimEnd('\') -ne $cleanInstallDir) {
            $newPaths += $p
        } else {
            $pathRemoved = $true
        }
    }

    if ($pathRemoved) {
        Write-Info "Removing $binDir from User PATH..."
        $newUserPath = $newPaths -join ';'
        [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
        $removedAny = $true
        
        # Also clean up current session PATH
        $currentSessionPaths = $env:Path -split ';' | Where-Object { $_.Trim().TrimEnd('\') -ne $cleanInstallDir }
        $env:Path = $currentSessionPaths -join ';'
    }
}

# 3. Clean up config file
$configFile = Join-Path $env:USERPROFILE ".nept.yaml"
if (Test-Path $configFile) {
    $deleteConfig = $false
    if ($Force) {
        $deleteConfig = $true
    } else {
        $confirmation = Read-Host "Do you want to delete the global configuration file ($configFile)? [y/N]"
        if ($confirmation -match '^[yY]([eE][sS])?$') {
            $deleteConfig = $true
        }
    }

    if ($deleteConfig) {
        try {
            Remove-Item -Path $configFile -Force
            Write-Success "Removed configuration file: $configFile"
        } catch {
            Write-Warn "Could not remove configuration file: $_"
        }
    } else {
        Write-Info "Kept configuration file: $configFile"
    }
}

if ($removedAny) {
    Write-Success "Successfully uninstalled nept."
} else {
    Write-Warn "No active installations or environment updates were needed."
}
