<#
.SYNOPSIS
    Install th (Terminal Help) CLI
    
.DESCRIPTION
    Installs the th CLI tool on Windows
    
.PARAMETER InstallDir
    Installation directory (default: $env:LOCALAPPDATA\th\bin)

.PARAMETER ModelDir
    Model directory (default: $env:LOCALAPPDATA\th\models)
    
.PARAMETER Force
    Force reinstall
    
.EXAMPLE
    .\install.ps1
    
.EXAMPLE
    .\install.ps1 -InstallDir "C:\Program Files\th" -Force
#>

param(
    [string]$InstallDir = "$env:LOCALAPPDATA\th\bin",
    [string]$ModelDir = "$env:LOCALAPPDATA\th\models",
    [switch]$Force
)

$Repo = "DarioHefti/th-local"
$BinaryName = "th"
$ModelName = "google_gemma-4-E2B-it-IQ2_M.gguf"
$ModelUrl = "https://huggingface.co/bartowski/google_gemma-4-E2B-it-GGUF/resolve/main/google_gemma-4-E2B-it-IQ2_M.gguf?download=true"
$MaxRetries = 3
$RetryDelay = 2

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Cyan
}

function Write-ErrorMsg {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing -TimeoutSec 30
        $version = $response.tag_name -replace '^v', ''
        
        if ($version -notmatch '^\d+\.\d+\.\d+$') {
            Write-ErrorMsg "Invalid version format: $version"
            return $null
        }
        
        return $version
    } catch {
        Write-ErrorMsg "Failed to get latest version: $_"
        return $null
    }
}

function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        "x86"   { return "386" }
        default { 
            Write-ErrorMsg "Unsupported architecture: $arch"
            return $null
        }
    }
}

function Add-ToPath {
    param([string]$Path)
    
    try {
        $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($currentPath -notlike "*$Path*") {
            [Environment]::SetEnvironmentVariable("Path", "$currentPath;$Path", "User")
            Write-Info "Added $Path to user PATH"
            Write-Info "Please restart your terminal for changes to take effect"
        }
    } catch {
        Write-ErrorMsg "Failed to update PATH: $_"
    }
}

function Test-Binary {
    param([string]$Path)
    
    if (-not (Test-Path $Path)) {
        Write-ErrorMsg "Binary not found at $Path"
        return $false
    }
    
    $fileInfo = Get-Item $Path
    if ($fileInfo.Length -eq 0) {
        Write-ErrorMsg "Binary is empty"
        return $false
    }
    
    return $true
}

function Ensure-Directory {
    param(
        [string]$Path,
        [string]$Label
    )

    if (-not (Test-Path $Path)) {
        try {
            New-Item -ItemType Directory -Path $Path -Force | Out-Null
        } catch {
            Write-ErrorMsg "Failed to create ${Label} directory: $_"
            exit 1
        }
    }

    if (-not (Test-Path $Path -PathType Container)) {
        Write-ErrorMsg "$Label directory is not accessible: $Path"
        exit 1
    }
}

function Download-File {
    param(
        [string]$Url,
        [string]$OutputPath,
        [string]$Description
    )

    $attempt = 1
    $success = $false
    
    while ($attempt -le $MaxRetries) {
        Write-Info "Downloading $Description (attempt $attempt of $MaxRetries)..."
        
        try {
            $ProgressPreference = 'SilentlyContinue'
            Invoke-WebRequest -Uri $Url -OutFile $OutputPath -UseBasicParsing -ErrorAction Stop
            $ProgressPreference = 'Normal'
            
            if ((Test-Path $OutputPath) -and ((Get-Item $OutputPath).Length -gt 0)) {
                $success = $true
                break
            } else {
                Write-ErrorMsg "Downloaded file is empty or missing"
            }
        } catch {
            Write-ErrorMsg "Download attempt $attempt failed: $_"
        }
        
        if ($attempt -lt $MaxRetries) {
            Write-Info "Retrying in ${RetryDelay}s..."
            Start-Sleep -Seconds $RetryDelay
        }
        
        $attempt++
    }
    
    $ProgressPreference = 'Normal'
    
    if ($success) {
        Write-Info "Download complete"
        return $true
    }
    
    if (Test-Path $OutputPath) {
        Remove-Item $OutputPath -Force -ErrorAction SilentlyContinue
    }

    return $false
}

function Ensure-Model {
    Ensure-Directory -Path $ModelDir -Label "Model"

    $finalPath = Join-Path $ModelDir $ModelName
    $partialPath = "$finalPath.part"

    if ((Test-Path $finalPath) -and -not $Force) {
        Write-Info "Model already present at $finalPath"
        return $finalPath
    }

    if (Test-Path $partialPath) {
        Remove-Item $partialPath -Force -ErrorAction SilentlyContinue
    }

    if (-not (Download-File -Url $ModelUrl -OutputPath $partialPath -Description $ModelName)) {
        Write-ErrorMsg "Failed to download model after $MaxRetries attempts"
        exit 1
    }

    try {
        Move-Item -Path $partialPath -Destination $finalPath -Force
    } catch {
        Write-ErrorMsg "Failed to install model: $_"
        if (Test-Path $partialPath) {
            Remove-Item $partialPath -Force -ErrorAction SilentlyContinue
        }
        exit 1
    }

    Write-Info "Model ready at $finalPath"
    return $finalPath
}

function Main {
    $finalPath = Join-Path $InstallDir "$BinaryName.exe"
    
    if ((Test-Path $finalPath) -and -not $Force) {
        Write-Info "$BinaryName is already installed at $finalPath"
        if (-not (Test-Binary -Path $finalPath)) {
            Write-ErrorMsg "Existing binary verification failed; rerun with -Force to replace it"
            exit 1
        }
    } else {
        Ensure-Directory -Path $InstallDir -Label "Install"

        $os = "windows"
        $arch = Get-Architecture
        
        if (-not $arch) {
            exit 1
        }
        
        Write-Info "Fetching latest version..."
        $version = Get-LatestVersion
        
        if (-not $version) {
            Write-ErrorMsg "Failed to get latest version"
            Write-ErrorMsg "Please check your network connection and try again"
            exit 1
        }
        
        $filename = "$BinaryName-$os-$arch.exe"
        $url = "https://github.com/$Repo/releases/download/v$version/$filename"
        $outputPath = Join-Path $InstallDir $filename
    
        Write-Info "Installing th v$version for $os/$arch..."
        
        if (-not (Download-File -Url $url -OutputPath $outputPath -Description $filename)) {
            Write-ErrorMsg "Failed to download binary after $MaxRetries attempts"
            if (Test-Path $outputPath) {
                Remove-Item $outputPath -Force
            }
            exit 1
        }
        
        if (Test-Path $finalPath) {
            Remove-Item $finalPath -Force
        }
        
        try {
            Move-Item -Path $outputPath -Destination $finalPath -Force
        } catch {
            Write-ErrorMsg "Failed to install binary: $_"
            exit 1
        }
        
        if (-not (Test-Binary -Path $finalPath)) {
            Write-ErrorMsg "Binary verification failed"
            if (Test-Path $finalPath) {
                Remove-Item $finalPath -Force
            }
            exit 1
        }
    }

    $installedModelPath = Ensure-Model
    
    Write-Host ""
    Write-Host "✓ Binary available at $finalPath" -ForegroundColor Green
    Write-Host "✓ Model available at $installedModelPath" -ForegroundColor Green
    
    Add-ToPath -Path $InstallDir
    
    Write-Host ""
    Write-Host "Run 'th --config' anytime to change your default model" -ForegroundColor Cyan
}

Main
