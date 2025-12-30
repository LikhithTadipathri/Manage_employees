# Quick test for maternity leave functionality
$BASE_URL = "http://localhost:8080"

Write-Host "========== MATERNITY LEAVE TEST ==========" -ForegroundColor Cyan

# 1. Login as jane_smith
Write-Host "`n1. Login as jane_smith..." -ForegroundColor Yellow
$loginResponse = Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method POST -ContentType "application/json" -Body (@{
    username = "jane_smith"
    password = "jane123"
} | ConvertTo-Json)

$token = $loginResponse.data.token
Write-Host "✅ Token: $($token.Substring(0, 20))..." -ForegroundColor Green

# 2. Get jane_smith employee details to verify gender
Write-Host "`n2. Get jane_smith employee details..." -ForegroundColor Yellow
$headers = @{
    Authorization = "Bearer $token"
    "Content-Type" = "application/json"
}

$empResponse = Invoke-RestMethod -Uri "$BASE_URL/api/v1/employees?limit=100" -Method GET -Headers $headers
$janeEmp = $empResponse.data | Where-Object { $_.last_name -eq "Smith" } | Select-Object -First 1

if ($janeEmp) {
    Write-Host "✅ Found: $($janeEmp.first_name) $($janeEmp.last_name)" -ForegroundColor Green
    Write-Host "   Gender: $($janeEmp.gender)" -ForegroundColor Cyan
    Write-Host "   Marital Status: $($janeEmp.marital_status)" -ForegroundColor Cyan
} else {
    Write-Host "❌ jane_smith not found" -ForegroundColor Red
    exit
}

# 3. Test maternity leave application
Write-Host "`n3. Apply for maternity leave (future dates)..." -ForegroundColor Yellow
$startDate = (Get-Date).AddDays(10).ToString("yyyy-MM-dd")
$endDate = (Get-Date).AddDays(20).ToString("yyyy-MM-dd")

$leaveBody = @{
    leave_type = "MATERNITY"
    start_date = $startDate
    end_date = $endDate
    reason = "Maternity leave"
} | ConvertTo-Json

try {
    $leaveResponse = Invoke-RestMethod -Uri "$BASE_URL/leave/apply" -Method POST -Headers $headers -Body $leaveBody
    Write-Host "✅ Maternity leave approved!" -ForegroundColor Green
    Write-Host "   Days: $($leaveResponse.data.days_count)" -ForegroundColor Cyan
    Write-Host "   Status: $($leaveResponse.data.status)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Maternity leave FAILED" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Yellow
}

Write-Host "`n========== TEST COMPLETE ==========" -ForegroundColor Cyan
