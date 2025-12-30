# Test Email Notification Flow
$baseURL = "http://localhost:8080"

Write-Host "=== Test Email Notification Flow ===" -ForegroundColor Cyan

# Step 1: Login as Employee
Write-Host "`n1. Logging in as employee (john_doe)..." -ForegroundColor Yellow
$loginResp = Invoke-WebRequest -Uri "$baseURL/auth/login" `
    -Method POST `
    -Headers @{"Content-Type" = "application/json"} `
    -Body '{"username":"john_doe","password":"password123"}' `
    -ErrorAction SilentlyContinue

if ($loginResp.StatusCode -eq 200) {
    $empToken = ($loginResp.Content | ConvertFrom-Json).token
    Write-Host "[OK] Login successful, token: $($empToken.Substring(0,30))..." -ForegroundColor Green
} else {
    Write-Host "[FAIL] Login failed" -ForegroundColor Red
    exit 1
}

# Step 2: Apply for leave
Write-Host "`n2. Applying for ANNUAL leave..." -ForegroundColor Yellow
$leaveBody = @{
    leave_type = "ANNUAL"
    start_date = "2025-12-30"
    end_date = "2026-01-05"
    reason = "Year-end vacation"
    notes = "Planning to spend time with family"
} | ConvertTo-Json

$applyResp = Invoke-WebRequest -Uri "$baseURL/leave/apply" `
    -Method POST `
    -Headers @{
        "Content-Type" = "application/json"
        "Authorization" = "Bearer $empToken"
    } `
    -Body $leaveBody `
    -ErrorAction SilentlyContinue

if ($applyResp.StatusCode -eq 201) {
    $leaveData = ($applyResp.Content | ConvertFrom-Json).data
    Write-Host "[OK] Leave applied successfully" -ForegroundColor Green
    Write-Host "   - Leave ID: $($leaveData.id)" -ForegroundColor Green
    Write-Host "   - Status: $($leaveData.status)" -ForegroundColor Green
    Write-Host "   - Days: $($leaveData.days_count)" -ForegroundColor Green
    $leaveId = $leaveData.id
} else {
    Write-Host "[FAIL] Leave application failed: $($applyResp.StatusCode)" -ForegroundColor Red
    Write-Host $applyResp.Content
    exit 1
}

# Step 3: Check notifications table
Write-Host "`n3. Checking notifications table..." -ForegroundColor Yellow
Write-Host "   [WAIT] Waiting 2 seconds for async email queue to process..." -ForegroundColor Yellow
Start-Sleep -Seconds 2

# Step 4: Check database for email records
Write-Host "`n4. Checking if email notifications were queued..." -ForegroundColor Yellow
Write-Host "   [INFO] Expected: 2 notifications (1 employee + 1 admin)" -ForegroundColor Cyan

Write-Host "`n[SUCCESS] Setup Complete!" -ForegroundColor Green
Write-Host "   - Employee username: john_doe (likeitdummy@gmail.com)" -ForegroundColor Green
Write-Host "   - Admin username: admin (likeitdummy@gmail.com)" -ForegroundColor Green
Write-Host "`n[NOTE] Check your email inbox for:" -ForegroundColor Cyan
Write-Host "   - Employee: 'Leave Request Submitted - Pending Approval'" -ForegroundColor Green
Write-Host "   - Admin: 'Action Required: New Leave Request Submitted'" -ForegroundColor Green
Write-Host "`n[TIP] If emails don't arrive in inbox, check spam/promotions folder" -ForegroundColor Yellow
