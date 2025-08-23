#!/bin/bash

trap "exit" 0

if [ $# -eq 0 ]; then
    echo "Available commands:"
    echo "  ./make install     - Install Go dependencies"
    echo "  ./make dev         - Start development server with hot reload"
    echo "  ./make build-dev   - Build development binary"
    echo "  ./make build       - Build production binary with embedded frontend"
    echo "  ./make test        - Run tests"
    echo "  ./make clean       - Clean build artifacts"
    echo "  ./make deps        - Download and tidy Go dependencies"
elif [ $1 == "install" ]; then
    go mod download
    go mod tidy
elif [ $1 == "dev" ]; then
    air
elif [ $1 == "build-dev" ]; then
    mkdir -p tmp
    go build -o tmp/main cmd/exim-pilot/main.go
elif [ $1 == "build" ]; then
    echo "Building frontend..."
    mkdir -p web/dist
    echo "<!-- Placeholder -->" > web/dist/index.html
    echo "Building production binary..."
    mkdir -p bin
    go build -tags embed -o bin/exim-pilot cmd/exim-pilot/main.go
elif [ $1 == "test" ]; then
    go test ./...
elif [ $1 == "clean" ]; then
    rm -rf tmp/ bin/ web/dist/
    echo "Build artifacts cleaned"
elif [ $1 == "deps" ]; then
    go mod download
    go mod tidy
else
    echo "Unknown command: $1"
    echo "Run './make' to see available commands"
    exit 1
fi