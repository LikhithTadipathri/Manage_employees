$baseURL = "http://localhost:8080"

Write-Host "=== LEAVE APPLICATION COMPLETE TEST ===" -ForegroundColor Cyan

# Step 1: Login
Write-Host "`n[1] Testing Login..." -ForegroundColor Yellow
$loginPayload = @{
    username = "john_doe"
    password = "john123"
} | ConvertTo-Json

try {
    $loginResp = Invoke-WebRequest -Uri "$baseURL/auth/login" `
        -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $loginPayload
} catch {
    Write-Host "Error: $($_.Exception.Message)"
    exit 1
}

Write-Host "Status: $($loginResp.StatusCode)"
$loginJson = $loginResp.Content | ConvertFrom-Json
$token = $loginJson.token

if (-not $token) {
    Write-Host "Failed to get token. Response: $($loginResp.Content)"
    exit 1
}

Write-Host "Login successful - Token received"

# Step 2: Check Leave Balance
Write-Host "`n[2] Checking Leave Balance..." -ForegroundColor Yellow
try {
    $balResp = Invoke-WebRequest -Uri "$baseURL/leave/balance" `
        -Method GET `
        -Headers @{
            "Content-Type" = "application/json"
            "Authorization" = "Bearer $token"
        }
} catch {
    Write-Host "Error: $($_.Exception.Message)"
    exit 1
}

Write-Host "Status: $($balResp.StatusCode)"
$balJson = $balResp.Content | ConvertFrom-Json
Write-Host "Leave Balances:"
$balJson.data.leave_balances | ForEach-Object { 
    Write-Host "  - $($_.leave_type): $($_.balance) days"
}

# Step 3: Apply Leave
Write-Host "`n[3] Applying for Leave..." -ForegroundColor Yellow
$leavePayload = @{
    leave_type = "ANNUAL"
    start_date = "2025-12-24"
    end_date = "2025-12-26"
    reason = "Holiday break"
} | ConvertTo-Json

Write-Host "Request: $leavePayload"
try {
    $leaveResp = Invoke-WebRequest -Uri "$baseURL/leave/apply" `
        -Method POST `
        -Headers @{
            "Content-Type" = "application/json"
            "Authorization" = "Bearer $token"
        } `
        -Body $leavePayload
} catch {
    Write-Host "Error: $($_.Exception.Message)"
    exit 1
}

Write-Host "Status: $($leaveResp.StatusCode)"
$leaveJson = $leaveResp.Content | ConvertFrom-Json

if ($leaveResp.StatusCode -eq 201) {
    Write-Host "Leave applied successfully!"
    Write-Host "Leave Request ID: $($leaveJson.data.id)"
    Write-Host "Status: $($leaveJson.data.status)"
    Write-Host "Days Count: $($leaveJson.data.days_count)"
} else {
    Write-Host "Failed to apply leave - Status: $($leaveResp.StatusCode)"
    Write-Host "Response: $($leaveResp.Content)"
}

# Step 4: Get My Leave Requests
Write-Host "`n[4] Retrieving Leave Requests..." -ForegroundColor Yellow
try {
    $reqResp = Invoke-WebRequest -Uri "$baseURL/leave/my-requests" `
        -Method GET `
        -Headers @{
            "Content-Type" = "application/json"
            "Authorization" = "Bearer $token"
        }
} catch {
    Write-Host "Error: $($_.Exception.Message)"
    exit 1
}

Write-Host "Status: $($reqResp.StatusCode)"
$reqJson = $reqResp.Content | ConvertFrom-Json
Write-Host "Total Requests: $($reqJson.data.count)"
$reqJson.data.leave_requests | ForEach-Object {
    Write-Host "  - ID: $($_.id), Type: $($_.leave_type), Status: $($_.status), Days: $($_.days_count)"
}

Write-Host "`nAll tests completed!" -ForegroundColor Green
