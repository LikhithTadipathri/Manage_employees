# Start Employee Service with SMTP Configuration
# This script sets SMTP environment variables and starts the server

Write-Host "=== Starting Employee Service with Email Notifications ===" -ForegroundColor Cyan
Write-Host ""

# Set SMTP environment variables
Write-Host "Setting SMTP configuration..." -ForegroundColor Yellow
$env:SMTP_HOST="smtp.gmail.com"
$env:SMTP_PORT="587"
$env:SMTP_USERNAME="no-reply@company.com"
$env:SMTP_PASSWORD="vjut usxv scsk pbuc"
$env:SMTP_FROM_ADDR="no-reply@company.com"
$env:SMTP_FROM_NAME="HR Management System"
$env:EMAIL_QUEUE_WORKERS="3"

Write-Host "✅ SMTP Configuration:" -ForegroundColor Green
Write-Host "   Host: $env:SMTP_HOST" -ForegroundColor White
Write-Host "   Port: $env:SMTP_PORT" -ForegroundColor White
Write-Host "   Username: $env:SMTP_USERNAME" -ForegroundColor White
Write-Host "   From: $env:SMTP_FROM_NAME <$env:SMTP_FROM_ADDR>" -ForegroundColor White
Write-Host "   Workers: $env:EMAIL_QUEUE_WORKERS" -ForegroundColor White
Write-Host ""

# Start the server
Write-Host "Starting server..." -ForegroundColor Yellow
Write-Host "Look for: '✅ Email queue started with 3 workers'" -ForegroundColor Cyan
Write-Host ""

go run cmd/employee-service/main.go
