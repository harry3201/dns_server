#!/bin/bash

# DNS Server Setup Script
# This script helps you quickly set up and test your DNS server

set -e  # Exit on error

echo "🚀 DNS Server Setup Script"
echo "=========================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed!"
    echo "Please install Go from: https://go.dev/dl/"
    exit 1
fi

echo "✅ Go $(go version) detected"
echo ""

# Check if dig is installed
if ! command -v dig &> /dev/null; then
    echo "⚠️  Warning: 'dig' command not found"
    echo "   On Ubuntu/Debian: sudo apt-get install dnsutils"
    echo "   On macOS: dig comes pre-installed"
    echo ""
fi

# Initialize Go module if go.mod doesn't exist
if [ ! -f "go.mod" ]; then
    echo "📦 Initializing Go module..."
    go mod init dns-server
    echo "✅ Go module initialized"
else
    echo "✅ Go module already exists"
fi
echo ""

# Run tests
echo "🧪 Running tests..."
go test -v -cover
echo ""

# Build the server
echo "🔨 Building DNS server..."
go build -o dns-server .
echo "✅ Build successful! Binary: ./dns-server"
echo ""

# Instructions
echo "📋 Setup Complete! Next steps:"
echo ""
echo "1️⃣  Start the server:"
echo "   go run ."
echo "   OR"
echo "   ./dns-server"
echo ""
echo "2️⃣  In another terminal, test it:"
echo "   dig @127.0.0.1 -p 2053 google.com A"
echo ""
echo "3️⃣  Test caching (run same query twice):"
echo "   dig @127.0.0.1 -p 2053 google.com A  # Cache miss (slower)"
echo "   dig @127.0.0.1 -p 2053 google.com A  # Cache hit (faster!)"
echo ""
echo "4️⃣  Test concurrency:"
echo "   for i in {1..50}; do dig @127.0.0.1 -p 2053 google.com A & done"
echo "   wait"
echo ""
echo "5️⃣  Try different record types:"
echo "   dig @127.0.0.1 -p 2053 google.com AAAA    # IPv6"
echo "   dig @127.0.0.1 -p 2053 www.github.com CNAME # CNAME"
echo "   dig @127.0.0.1 -p 2053 gmail.com MX        # Mail servers"
echo ""
echo "🎯 Pro tip: Watch the server logs to see cache hits/misses!"
echo ""
echo "✨ Happy DNS hacking!"
