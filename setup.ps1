# -------------------------------
# DNS Server Setup Script (Windows PowerShell)
# -------------------------------

Write-Host "`n=== DNS Server Setup Script (Windows) ===`n"

# Check Go installation
$goCheck = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCheck) {
    Write-Host "ERROR: Go is NOT installed."
    Write-Host "Download Go from: https://go.dev/dl/"
    exit
}
Write-Host "Go detected: $(go version)`n"

# Initialize go.mod if missing
if (-not (Test-Path "go.mod")) {
    Write-Host "Initializing Go module..."
    go mod init dns-server
    Write-Host "go.mod created.`n"
}
else {
    Write-Host "go.mod already exists.`n"
}

# Run tests
Write-Host "Running tests..."
go test -v -cover
Write-Host "`n"

# Build the server
Write-Host "Building DNS server..."
go build -o dns-server.exe .
Write-Host "Build complete -> dns-server.exe`n"

# Final instructions
Write-Host "NEXT STEPS:"
Write-Host "---------------------------------"
Write-Host "1) Run the server:"
Write-Host "   go run ."
Write-Host "   OR"
Write-Host "   .\dns-server.exe"
Write-Host ""

Write-Host "2) Open a NEW PowerShell window and test:"
Write-Host "   nslookup google.com 127.0.0.1:2053"
Write-Host "   OR if you have dig installed:"
Write-Host "   dig @127.0.0.1 -p 2053 google.com A"
Write-Host ""

Write-Host "3) Cache testing:"
Write-Host "   dig @127.0.0.1 -p 2053 google.com A"
Write-Host "   dig @127.0.0.1 -p 2053 google.com A"
Write-Host ""

Write-Host "Setup Complete."
