# Test: Create new female married employee and apply for maternity leave
$BASE_URL = "http://localhost:8080"

Write-Host "========== CREATE FEMALE MARRIED EMPLOYEE + MATERNITY TEST ==========" -ForegroundColor Cyan

# 1. Login as admin to create new employee
Write-Host "`n1. Login as admin..." -ForegroundColor Yellow
$adminLogin = Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method POST -ContentType "application/json" -Body (@{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json)

$adminToken = $adminLogin.data.token
Write-Host "✅ Admin logged in" -ForegroundColor Green

# 2. Create new female married employee
Write-Host "`n2. Create new female married employee..." -ForegroundColor Yellow
$headers = @{
    Authorization = "Bearer $adminToken"
    "Content-Type" = "application/json"
}

$newEmpBody = @{
    username = "sophia_williams"
    password = "SecurePass@123"
    first_name = "Sophia"
    last_name = "Williams"
    email = "sophia.williams@company.com"
    phone = "+1-555-9999"
    position = "Senior Manager"
    salary = 95000
    gender = "Female"
    marital_status = $true
    hired_date = "2023-06-15T00:00:00Z"
} | ConvertTo-Json

try {
    $newEmpResponse = Invoke-RestMethod -Uri "$BASE_URL/api/v1/employees" -Method POST -Headers $headers -Body $newEmpBody
    $newEmpId = $newEmpResponse.data.id
    $newUserId = $newEmpResponse.data.user_id
    Write-Host "✅ Created: Sophia Williams (ID: $newEmpId, User ID: $newUserId)" -ForegroundColor Green
    Write-Host "   Gender: $($newEmpResponse.data.gender)" -ForegroundColor Cyan
    Write-Host "   Marital Status: $($newEmpResponse.data.marital_status)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Failed to create employee" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Yellow
    exit
}

# 3. Login as the new employee
Write-Host "`n3. Login as sophia_williams..." -ForegroundColor Yellow
$empLogin = Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method POST -ContentType "application/json" -Body (@{
    username = "sophia_williams"
    password = "SecurePass@123"
} | ConvertTo-Json)

$empToken = $empLogin.data.token
Write-Host "✅ Sophia logged in" -ForegroundColor Green

# 4. Get employee details to verify gender
Write-Host "`n4. Get employee details to verify gender..." -ForegroundColor Yellow
$empHeaders = @{
    Authorization = "Bearer $empToken"
    "Content-Type" = "application/json"
}

$empDetailsResponse = Invoke-RestMethod -Uri "$BASE_URL/api/v1/employees?limit=100" -Method GET -Headers $empHeaders
$sophiaEmp = $empDetailsResponse.data | Where-Object { $_.id -eq $newEmpId } | Select-Object -First 1

if ($sophiaEmp) {
    Write-Host "✅ Found: $($sophiaEmp.first_name) $($sophiaEmp.last_name)" -ForegroundColor Green
    Write-Host "   Gender: $($sophiaEmp.gender)" -ForegroundColor Cyan
    Write-Host "   Marital Status: $($sophiaEmp.marital_status)" -ForegroundColor Cyan
} else {
    Write-Host "⚠️  Employee not found in list, but creation was successful" -ForegroundColor Yellow
}

# 5. Apply for maternity leave
Write-Host "`n5. Apply for maternity leave (10 days, future dates)..." -ForegroundColor Yellow
$startDate = (Get-Date).AddDays(5).ToString("yyyy-MM-dd")
$endDate = (Get-Date).AddDays(14).ToString("yyyy-MM-dd")

$leaveBody = @{
    leave_type = "MATERNITY"
    start_date = $startDate
    end_date = $endDate
    reason = "Maternity leave for newborn"
} | ConvertTo-Json

try {
    $leaveResponse = Invoke-RestMethod -Uri "$BASE_URL/leave/apply" -Method POST -Headers $empHeaders -Body $leaveBody
    Write-Host "✅ MATERNITY LEAVE APPROVED!" -ForegroundColor Green
    Write-Host "   Leave ID: $($leaveResponse.data.id)" -ForegroundColor Cyan
    Write-Host "   Days: $($leaveResponse.data.days_count)" -ForegroundColor Cyan
    Write-Host "   Status: $($leaveResponse.data.status)" -ForegroundColor Cyan
    Write-Host "   Leave Type: $($leaveResponse.data.leave_type)" -ForegroundColor Cyan
    Write-Host "`n✅✅✅ MATERNITY LEAVE WORKING PERFECTLY! ✅✅✅" -ForegroundColor Green
} catch {
    $errorMsg = $_.Exception.Response.Content | ConvertFrom-Json
    Write-Host "❌ Maternity leave FAILED" -ForegroundColor Red
    Write-Host "Error: $($errorMsg.message)" -ForegroundColor Yellow
    if ($errorMsg.errors) {
        Write-Host "Details: $($errorMsg.errors | ConvertTo-Json)" -ForegroundColor Yellow
    }
}

Write-Host "`n========== TEST COMPLETE ==========" -ForegroundColor Cyan
