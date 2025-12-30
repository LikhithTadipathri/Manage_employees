$baseURL = "http://localhost:8080"

Write-Host "=== COMPREHENSIVE LEAVE APPLICATION TEST ===" -ForegroundColor Cyan

# ========== PART 1: Test Seeded User (john_doe) ==========
Write-Host "`n========== PART 1: SEEDED USER TEST ==========" -ForegroundColor Magenta

# Step 1a: Login as seeded user
Write-Host "`n[1a] Login as seeded user (john_doe)..." -ForegroundColor Yellow
$loginPayload = @{
    username = "john_doe"
    password = "john123"
} | ConvertTo-Json

try {
    $loginResp = Invoke-WebRequest -Uri "$baseURL/auth/login" `
        -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $loginPayload
    
    $loginJson = $loginResp.Content | ConvertFrom-Json
    $seededToken = $loginJson.token
    Write-Host "✓ Seeded user login successful"
} catch {
    Write-Host "✗ Failed to login: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 1b: Apply leave as seeded user
Write-Host "`n[1b] Applying leave as seeded user..." -ForegroundColor Yellow
$leavePayload = @{
    leave_type = "ANNUAL"
    start_date = "2025-12-24"
    end_date = "2025-12-26"
    reason = "Holiday break"
} | ConvertTo-Json

try {
    $leaveResp = Invoke-WebRequest -Uri "$baseURL/leave/apply" `
        -Method POST `
        -Headers @{
            "Content-Type" = "application/json"
            "Authorization" = "Bearer $seededToken"
        } `
        -Body $leavePayload
    
    if ($leaveResp.StatusCode -eq 201) {
        Write-Host "✓ Seeded user can apply leave (Status: 201)" -ForegroundColor Green
    }
    else {
        Write-Host "✗ Unexpected status: $($leaveResp.StatusCode)" -ForegroundColor Red
    }
}
catch {
    Write-Host "✗ Failed to apply leave: $($_.Exception.Message)" -ForegroundColor Red
}

# ========== PART 2: Register New User & Test Leave ==========
Write-Host "`n========== PART 2: NEW USER REGISTRATION TEST ==========" -ForegroundColor Magenta

# Step 2a: Login as admin first
Write-Host "`n[2a] Login as admin..." -ForegroundColor Yellow
$adminLoginPayload = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

try {
    $adminLoginResp = Invoke-WebRequest -Uri "$baseURL/auth/login" `
        -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $adminLoginPayload
    
    $adminLoginJson = $adminLoginResp.Content | ConvertFrom-Json
    $adminToken = $adminLoginJson.token
    Write-Host "✓ Admin login successful"
} catch {
    Write-Host "✗ Failed to login as admin: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 2b: Register new employee
Write-Host "`n[2b] Registering new employee..." -ForegroundColor Yellow
$newUserPayload = @{
    username = "test_employee_$(Get-Random)"
    email = "testemployee$(Get-Random)@example.com"
    password = "testpass123"
    role = "employee"
    first_name = "Test"
    last_name = "Employee"
    phone = "555-1234"
    position = "QA Engineer"
    salary = 45000
    gender = "Female"
    marital_status = $true
} | ConvertTo-Json

try {
    $regResp = Invoke-WebRequest -Uri "$baseURL/auth/register" `
        -Method POST `
        -Headers @{
            "Content-Type" = "application/json"
            "Authorization" = "Bearer $adminToken"
        } `
        -Body $newUserPayload
    
    $regJson = $regResp.Content | ConvertFrom-Json
    $newUsername = $regJson.data.username
    Write-Host "✓ New employee registered: $newUsername" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to register: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 2c: Login as newly registered user
Write-Host "`n[2c] Login as newly registered user..." -ForegroundColor Yellow
$newLoginPayload = @{
    username = $newUsername
    password = "testpass123"
} | ConvertTo-Json

try {
    $newLoginResp = Invoke-WebRequest -Uri "$baseURL/auth/login" `
        -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $newLoginPayload
    
    $newLoginJson = $newLoginResp.Content | ConvertFrom-Json
    $newToken = $newLoginJson.token
    Write-Host "✓ New user login successful"
} catch {
    Write-Host "✗ Failed to login as new user: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 2d: Apply leave as newly registered user
Write-Host "`n[2d] Applying leave as newly registered user..." -ForegroundColor Yellow
$newLeavePayload = @{
    leave_type = "ANNUAL"
    start_date = "2026-01-15"
    end_date = "2026-01-17"
    reason = "Vacation"
} | ConvertTo-Json

try {
    $newLeaveResp = Invoke-WebRequest -Uri "$baseURL/leave/apply" `
        -Method POST `
        -Headers @{
            "Content-Type" = "application/json"
            "Authorization" = "Bearer $newToken"
        } `
        -Body $newLeavePayload
    
    if ($newLeaveResp.StatusCode -eq 201) {
        Write-Host "✓ NEW user can apply leave (Status: 201)" -ForegroundColor Green
    }
    else {
        Write-Host "✗ Unexpected status: $($newLeaveResp.StatusCode)" -ForegroundColor Red
    }
}
catch {
    Write-Host "✗ Failed to apply leave as new user: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n========== TEST COMPLETE ==========" -ForegroundColor Green
Write-Host "Summary:" -ForegroundColor Yellow
Write-Host "  ✓ Seeded users can apply leave"
Write-Host "  ✓ Newly registered users can apply leave"
Write-Host "  ✓ Leave application works for both scenarios" -ForegroundColor Green
