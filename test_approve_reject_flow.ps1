# Test Leave Application and Approval Flow
$baseURL = "http://localhost:8080"

Write-Host "=== Testing Leave Application and Approval ===" -ForegroundColor Cyan

# Step 1: Login as Employee
Write-Host "`nStep 1: Login as employee (john_doe)..."
$loginResp = Invoke-WebRequest -Uri "$baseURL/auth/login" `
    -Method POST `
    -Headers @{"Content-Type" = "application/json"} `
    -Body '{"username":"john_doe","password":"password123"}' `
    -ErrorAction SilentlyContinue

if ($loginResp.StatusCode -eq 200) {
    $empToken = ($loginResp.Content | ConvertFrom-Json).token
    Write-Host "[OK] Employee login successful" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Employee login failed" -ForegroundColor Red
    exit
}

# Step 2: Login as Admin
Write-Host "`nStep 2: Login as admin..."
$adminLoginResp = Invoke-WebRequest -Uri "$baseURL/auth/login" `
    -Method POST `
    -Headers @{"Content-Type" = "application/json"} `
    -Body '{"username":"admin","password":"admin123"}' `
    -ErrorAction SilentlyContinue

if ($adminLoginResp.StatusCode -eq 200) {
    $adminToken = ($adminLoginResp.Content | ConvertFrom-Json).token
    Write-Host "[OK] Admin login successful" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Admin login failed" -ForegroundColor Red
    exit
}

# Step 3: Apply for leave
Write-Host "`nStep 3: Apply for ANNUAL leave..."
$leaveBody = @{
    leave_type = "ANNUAL"
    start_date = "2026-01-10"
    end_date = "2026-01-15"
    reason = "Winter vacation"
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
    $leaveId = $leaveData.id
    Write-Host "[OK] Leave applied successfully" -ForegroundColor Green
    Write-Host "     Leave ID: $leaveId | Status: $($leaveData.status)" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Leave application failed: $($applyResp.StatusCode)" -ForegroundColor Red
    Write-Host $applyResp.Content
    exit
}

# Step 4: Approve the leave
Write-Host "`nStep 4: Admin approving leave (ID: $leaveId)..."
$approveBody = @{
    notes = "Approved - enjoy your vacation!"
} | ConvertTo-Json

$approveResp = Invoke-WebRequest -Uri "$baseURL/leave/approve/$leaveId" `
    -Method POST `
    -Headers @{
        "Content-Type" = "application/json"
        "Authorization" = "Bearer $adminToken"
    } `
    -Body $approveBody `
    -ErrorAction SilentlyContinue

if ($approveResp.StatusCode -eq 200) {
    Write-Host "[OK] Leave approved successfully" -ForegroundColor Green
    Write-Host $approveResp.Content
} else {
    Write-Host "[FAIL] Approve failed: $($approveResp.StatusCode)" -ForegroundColor Red
    Write-Host $approveResp.Content
}

# Step 5: Reject another leave
Write-Host "`nStep 5: Apply another leave to test rejection..."
$leaveBody2 = @{
    leave_type = "ANNUAL"
    start_date = "2026-02-01"
    end_date = "2026-02-05"
    reason = "Summer plans"
} | ConvertTo-Json

$applyResp2 = Invoke-WebRequest -Uri "$baseURL/leave/apply" `
    -Method POST `
    -Headers @{
        "Content-Type" = "application/json"
        "Authorization" = "Bearer $empToken"
    } `
    -Body $leaveBody2 `
    -ErrorAction SilentlyContinue

if ($applyResp2.StatusCode -eq 201) {
    $leaveData2 = ($applyResp2.Content | ConvertFrom-Json).data
    $leaveId2 = $leaveData2.id
    Write-Host "[OK] Second leave applied (ID: $leaveId2)" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Second leave application failed" -ForegroundColor Red
    exit
}

# Step 6: Reject the leave
Write-Host "`nStep 6: Admin rejecting leave (ID: $leaveId2)..."
$rejectBody = @{
    reason = "Budget constraints for this period"
} | ConvertTo-Json

$rejectResp = Invoke-WebRequest -Uri "$baseURL/leave/reject/$leaveId2" `
    -Method POST `
    -Headers @{
        "Content-Type" = "application/json"
        "Authorization" = "Bearer $adminToken"
    } `
    -Body $rejectBody `
    -ErrorAction SilentlyContinue

if ($rejectResp.StatusCode -eq 200) {
    Write-Host "[OK] Leave rejected successfully" -ForegroundColor Green
    Write-Host $rejectResp.Content
} else {
    Write-Host "[FAIL] Rejection failed: $($rejectResp.StatusCode)" -ForegroundColor Red
    Write-Host $rejectResp.Content
}

Write-Host "`n=== All Tests Complete ===" -ForegroundColor Cyan
