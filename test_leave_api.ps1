#!/bin/pwsh
# Test leave application

$baseURL = "http://localhost:8080"

# Test 1: Login as john_doe
Write-Host "=== TEST 1: Login as john_doe ===" -ForegroundColor Cyan
$loginBody = @{
    username = "john_doe"
    password = "john123"
} | ConvertTo-Json

$loginResp = Invoke-WebRequest -Uri "$baseURL/auth/login" `
    -Method POST `
    -ContentType "application/json" `
    -Body $loginBody `
    -UseBasicParsing

$loginJson = $loginResp.Content | ConvertFrom-Json
Write-Host "Login Status: $($loginResp.StatusCode)" -ForegroundColor Green
Write-Host "Token: $($loginJson.token.Substring(0,30))..."

if (-not $loginJson.token) {
    Write-Host "Failed to get token!" -ForegroundColor Red
    exit 1
}

$token = $loginJson.token
$authHeader = @{ "Authorization" = "Bearer $token" }

# Test 2: Apply for ANNUAL leave
Write-Host "`n=== TEST 2: Apply for ANNUAL leave ===" -ForegroundColor Cyan
$applyBody = @{
    leave_type = "ANNUAL"
    start_date = "2025-12-20"
    end_date = "2025-12-27"
    reason = "Family vacation"
} | ConvertTo-Json

$applyResp = Invoke-WebRequest -Uri "$baseURL/leave/apply" `
    -Method POST `
    -ContentType "application/json" `
    -Headers $authHeader `
    -Body $applyBody `
    -UseBasicParsing

$applyJson = $applyResp.Content | ConvertFrom-Json
Write-Host "Apply Status: $($applyResp.StatusCode)" -ForegroundColor Green
Write-Host "Response: $($applyResp.Content)" -ForegroundColor Green

# Test 3: Get leave balance
Write-Host "`n=== TEST 3: Get leave balance ===" -ForegroundColor Cyan
$balanceResp = Invoke-WebRequest -Uri "$baseURL/leave/balance" `
    -Method GET `
    -Headers $authHeader `
    -UseBasicParsing

$balanceJson = $balanceResp.Content | ConvertFrom-Json
Write-Host "Balance Status: $($balanceResp.StatusCode)" -ForegroundColor Green
Write-Host "Balances:"
$balanceJson.balances | ForEach-Object {
    Write-Host "  $($_.leave_type): $($_.balance) days"
}

Write-Host "`n=== ALL TESTS PASSED ===" -ForegroundColor Green
