# Test Maternity Leave Fix
$baseUrl = "http://localhost:8080"

Write-Host "=== Testing Maternity Leave Initialization Fix ===" -ForegroundColor Cyan
Write-Host ""

# 1. Login as jane_smith (female, married)
Write-Host "Step 1: Login as jane_smith..." -ForegroundColor Yellow
$loginBody = @{
    username = "jane_smith"
    password = "password123"
} | ConvertTo-Json

$loginResponse = Invoke-WebRequest -Uri "$baseUrl/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body $loginBody -SkipHttpErrorCheck
$loginData = $loginResponse.Content | ConvertFrom-Json

if ($loginResponse.StatusCode -eq 200) {
    $token = $loginData.data.token
    Write-Host "✅ Login successful" -ForegroundColor Green
    Write-Host "Token: $($token.Substring(0, 50))..." -ForegroundColor Gray
} else {
    Write-Host "❌ Login failed" -ForegroundColor Red
    exit 1
}

Write-Host ""

# 2. Get current leave balance before applying
Write-Host "Step 2: Check current leave balance (MATERNITY)..." -ForegroundColor Yellow
$balanceResponse = Invoke-WebRequest -Uri "$baseUrl/leave/balance/MATERNITY" -Method GET -Headers @{"Authorization"="Bearer $token"} -SkipHttpErrorCheck
$balanceData = $balanceResponse.Content | ConvertFrom-Json

if ($balanceResponse.StatusCode -eq 200) {
    Write-Host "✅ Leave balance retrieved" -ForegroundColor Green
    Write-Host "Balance: $($balanceData.data.balance) days" -ForegroundColor Gray
} else {
    Write-Host "Balance response: $($balanceResponse.StatusCode)" -ForegroundColor Gray
    Write-Host "Response: $($balanceResponse.Content | ConvertFrom-Json | ConvertTo-Json)" -ForegroundColor Gray
}

Write-Host ""

# 3. Apply for maternity leave
Write-Host "Step 3: Apply for maternity leave..." -ForegroundColor Yellow
$leaveBody = @{
    leave_type = "MATERNITY"
    start_date = "2025-12-22"
    end_date = "2025-12-30"
    reason = "Maternity leave for childbirth"
    notes = "Doctor approved"
} | ConvertTo-Json

$leaveResponse = Invoke-WebRequest -Uri "$baseUrl/leave/apply" -Method POST -Headers @{"Authorization"="Bearer $token"; "Content-Type"="application/json"} -Body $leaveBody -SkipHttpErrorCheck
$leaveData = $leaveResponse.Content | ConvertFrom-Json

if ($leaveResponse.StatusCode -eq 201) {
    Write-Host "✅ Maternity leave applied successfully!" -ForegroundColor Green
    Write-Host "Leave Request ID: $($leaveData.data.id)" -ForegroundColor Gray
    Write-Host "Status: $($leaveData.data.status)" -ForegroundColor Gray
    Write-Host "Days: $($leaveData.data.days_count)" -ForegroundColor Gray
} else {
    Write-Host "❌ Failed to apply leave" -ForegroundColor Red
    Write-Host "Status: $($leaveResponse.StatusCode)" -ForegroundColor Gray
    $errorData = $leaveResponse.Content | ConvertFrom-Json
    Write-Host "Error: $($errorData.message)" -ForegroundColor Red
    if ($errorData.errors) {
        Write-Host "Details: $(ConvertTo-Json $errorData.errors)" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "=== Test Complete ===" -ForegroundColor Cyan
