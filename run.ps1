# Quibit CLI Runner Script for Windows PowerShell
# This script builds and runs the Quibit CLI in a Docker container

# Colors for output
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$White = "White"

# Function to print colored output
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor $Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Red
}

# Check if Docker is available
try {
    $null = Get-Command docker -ErrorAction Stop
} catch {
    Write-Error "Docker is not installed or not in PATH"
    Write-Error "Please install Docker Desktop to run Quibit CLI"
    exit 1
}

# Check if Docker daemon is running
try {
    $null = docker info 2>$null
} catch {
    Write-Error "Docker daemon is not running"
    Write-Error "Please start Docker Desktop"
    exit 1
}

# Set variables
$ImageName = "quibit-cli"
$ContainerName = "quibit-runner"

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-Warning ".env file not found. CLI will run with default settings."
    Write-Warning "Copy .env.example to .env and configure if needed."
}

# Build Docker image
Write-Status "Building Quibit CLI Docker image..."
$BuildResult = docker build -t $ImageName .
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to build Docker image"
    exit 1
}

Write-Status "Docker image built successfully"

# Run the container
Write-Status "Starting Quibit CLI..."

# Prepare docker run command
$DockerRunCmd = @("run", "--rm", "-it")

# Add .env file if exists
if (Test-Path ".env") {
    $DockerRunCmd += "--env-file", ".env"
}

# Add volume for current directory (for any file operations)
$CurrentPath = (Get-Location).Path
$DockerRunCmd += "-v", "${CurrentPath}:/workspace"

# Set container name and image
$DockerRunCmd += "--name", $ContainerName, $ImageName

# Execute the command
& docker @DockerRunCmd

Write-Status "Quibit CLI session ended"
