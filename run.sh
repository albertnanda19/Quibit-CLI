#!/bin/bash

# Quibit CLI Runner Script
# This script builds and runs the Quibit CLI in a Docker container

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed or not in PATH"
    print_error "Please install Docker to run Quibit CLI"
    exit 1
fi

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    print_error "Docker daemon is not running"
    print_error "Please start Docker daemon"
    exit 1
fi

# Set variables
IMAGE_NAME="quibit-cli"
CONTAINER_NAME="quibit-runner-$(date +%s)"

# Check if .env file exists
if [ ! -f ".env" ]; then
    print_warning ".env file not found. CLI will run with default settings."
    print_warning "Copy .env.example to .env and configure if needed."
fi

# Build Docker image
print_status "Building Quibit CLI Docker image..."
if ! docker build -t "$IMAGE_NAME" .; then
    print_error "Failed to build Docker image"
    exit 1
fi

print_status "Docker image built successfully"

# Run the container
print_status "Starting Quibit CLI..."

# Prepare docker run command
DOCKER_RUN_CMD="docker run --rm -it"

# Add .env file if exists
if [ -f ".env" ]; then
    DOCKER_RUN_CMD="$DOCKER_RUN_CMD --env-file .env"
fi

# Add volume for current directory (for any file operations)
DOCKER_RUN_CMD="$DOCKER_RUN_CMD -v $(pwd):/workspace"

# Set container name and image
DOCKER_RUN_CMD="$DOCKER_RUN_CMD --name $CONTAINER_NAME $IMAGE_NAME"

# Add command line arguments if provided
if [ $# -gt 0 ]; then
    DOCKER_RUN_CMD="$DOCKER_RUN_CMD $*"
fi

# Execute the command
eval $DOCKER_RUN_CMD

print_status "Quibit CLI session ended"
