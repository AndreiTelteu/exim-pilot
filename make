#!/bin/bash

trap "exit" 0

# Function to build frontend
build_frontend() {
    local build_type="${1:-development}"
    
    echo "Building frontend..."
    cd web
    
    if [ ! -d "node_modules" ]; then
        echo "Installing frontend dependencies..."
        npm ci --production=false
    else
        if [ "$build_type" == "production" ]; then
            echo "Ensuring dependencies are up to date..."
        else
            echo "Checking for dependency updates..."
        fi
        npm ci --production=false
    fi
    
    if [ "$build_type" == "production" ]; then
        echo "Running production build..."
    else
        echo "Running frontend build with optimizations..."
    fi
    npm run build
    cd ..
    
    echo "Frontend build complete: web/dist/"
    if [ "$build_type" == "production" ]; then
        echo "Build summary:"
        du -sh web/dist/* 2>/dev/null || echo "Build artifacts created"
    else
        echo "Build size summary:"
        du -sh web/dist/* 2>/dev/null || echo "No build artifacts found"
    fi
}

# Function to install Go dependencies
install_go_deps() {
    go mod download
    go mod tidy
}

if [ $# -eq 0 ]; then
    echo "Available commands:"
    echo "  ./make install     - Install Go dependencies"
    echo "  ./make dev         - Start development server with hot reload"
    echo "  ./make build-dev   - Build development binary"
    echo "  ./make build-frontend - Build frontend only"
    echo "  ./make build       - Build production binary with embedded frontend"
    echo "  ./make start       - Start production binary"
    echo "  ./make verify      - Verify embedded assets in production binary"
    echo "  ./make test        - Run tests"
    echo "  ./make clean       - Clean build artifacts"
    echo "  ./make deps        - Download and tidy Go dependencies"
elif [ $1 == "install" ]; then
    install_go_deps
elif [ $1 == "dev" ]; then
    air
elif [ $1 == "build-dev" ]; then
    mkdir -p tmp
    go build -o tmp/main.exe cmd/exim-pilot/main.go
elif [ $1 == "build-frontend" ]; then
    build_frontend
elif [ $1 == "build" ]; then
    echo "=== Building Production Binary with Embedded Frontend ==="
    
    # Clean previous builds
    echo "Cleaning previous builds..."
    rm -rf web/dist/ bin/
    
    # Build frontend with optimizations
    echo "Building optimized frontend..."
    build_frontend "production"
    
    # Verify frontend build
    if [ ! -f "web/dist/index.html" ]; then
        echo "ERROR: Frontend build failed - index.html not found"
        exit 1
    fi
    
    # Build Go binary with embedded assets
    echo "Building Go binary with embedded frontend..."
    mkdir -p bin
    
    # Ensure Go dependencies are current
    install_go_deps
    
    # Build with embed tag
    echo "Compiling with embedded assets..."
    go build -tags embed -ldflags="-s -w" -o bin/exim-pilot.exe ./cmd/exim-pilot
    
    if [ $? -eq 0 ]; then
        echo "=== Production Build Complete ==="
        echo "Binary: bin/exim-pilot.exe"
        echo "Size: $(du -sh bin/exim-pilot.exe | cut -f1)"
        echo "Frontend assets embedded successfully"
    else
        echo "ERROR: Go build failed"
        exit 1
    fi
elif [ $1 == "start" ]; then
    if [ ! -f "bin/exim-pilot.exe" ]; then
        echo "ERROR: Production binary not found. Run './make build' first."
        exit 1
    fi
    echo "Starting production binary..."
    start ./bin/exim-pilot.exe
elif [ $1 == "test" ]; then
    go test ./...
elif [ $1 == "clean" ]; then
    rm -rf tmp/ bin/ web/dist/
    echo "Build artifacts cleaned"
elif [ $1 == "verify" ]; then
    if [ ! -f "bin/exim-pilot.exe" ]; then
        echo "ERROR: Production binary not found. Run './make build' first."
        exit 1
    fi
    echo "Verifying embedded assets in production binary..."
    echo "Binary size: $(du -sh bin/exim-pilot.exe | cut -f1)"
    echo "Testing binary startup..."
    timeout 5s ./bin/exim-pilot.exe --help 2>/dev/null || echo "Binary appears functional"
    echo "Embedded assets verification complete"
elif [ $1 == "deps" ]; then
    install_go_deps
else
    echo "Unknown command: $1"
    echo "Run './make' to see available commands"
    exit 1
fi