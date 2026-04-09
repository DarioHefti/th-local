param(
	[string]$ModelDir = "$env:LOCALAPPDATA\th\models",
	[switch]$Force
)

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

function Ensure-Directory {
	if (-not (Test-Path $ModelDir)) {
		try {
			New-Item -ItemType Directory -Path $ModelDir -Force | Out-Null
		} catch {
			Write-ErrorMsg "Failed to create model directory: $_"
			exit 1
		}
	}

	if (-not (Test-Path $ModelDir -PathType Container)) {
		Write-ErrorMsg "Model directory is not accessible: $ModelDir"
		exit 1
	}
}

function Download-Model {
	$finalPath = Join-Path $ModelDir $ModelName
	$partialPath = "$finalPath.part"
	$attempt = 1

	if ((Test-Path $finalPath) -and -not $Force) {
		Write-Info "Model already present at $finalPath"
		return $finalPath
	}

	if (Test-Path $partialPath) {
		Remove-Item $partialPath -Force -ErrorAction SilentlyContinue
	}

	while ($attempt -le $MaxRetries) {
		Write-Info "Downloading $ModelName (attempt $attempt of $MaxRetries)..."

		try {
			$ProgressPreference = 'SilentlyContinue'
			Invoke-WebRequest -Uri $ModelUrl -OutFile $partialPath -UseBasicParsing -ErrorAction Stop
			$ProgressPreference = 'Normal'

			if ((Test-Path $partialPath) -and ((Get-Item $partialPath).Length -gt 0)) {
				Move-Item -Path $partialPath -Destination $finalPath -Force
				Write-Info "Model ready at $finalPath"
				return $finalPath
			}

			Write-ErrorMsg "Downloaded model is empty"
		} catch {
			$ProgressPreference = 'Normal'
			Write-ErrorMsg "Download attempt $attempt failed: $_"
		}

		if (Test-Path $partialPath) {
			Remove-Item $partialPath -Force -ErrorAction SilentlyContinue
		}

		if ($attempt -lt $MaxRetries) {
			Write-Info "Retrying in ${RetryDelay}s..."
			Start-Sleep -Seconds $RetryDelay
		}

		$attempt++
	}

	Write-ErrorMsg "Failed to download model after $MaxRetries attempts"
	exit 1
}

Ensure-Directory
$installedModelPath = Download-Model

Write-Host ""
Write-Host "✓ Model available at $installedModelPath" -ForegroundColor Green