# Check and clear records with NULL/empty gender values
$BASE_URL = "http://localhost:8080"

Write-Host "========== DATABASE CLEANUP SCRIPT ==========" -ForegroundColor Cyan

# 1. Login as admin to check employees
Write-Host "`n1. Checking all employees..." -ForegroundColor Yellow
$adminLogin = Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method POST -ContentType "application/json" -Body (@{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json)

$adminToken = $adminLogin.data.token
$headers = @{
    Authorization = "Bearer $adminToken"
    "Content-Type" = "application/json"
}

# Get all employees (increased limit)
$empResponse = Invoke-RestMethod -Uri "$BASE_URL/api/v1/employees?limit=1000" -Method GET -Headers $headers
$allEmps = $empResponse.data

Write-Host "Total employees: $($allEmps.Count)" -ForegroundColor Cyan

# Check for problematic records
$badRecords = $allEmps | Where-Object { [string]::IsNullOrEmpty($_.gender) -or $_.gender -eq "" }
Write-Host "Employees with NULL/empty gender: $($badRecords.Count)" -ForegroundColor Yellow

if ($badRecords.Count -gt 0) {
    Write-Host "`nProblematic records:" -ForegroundColor Red
    $badRecords | ForEach-Object {
        Write-Host "  - ID: $($_.id), Name: $($_.first_name) $($_.last_name), Gender: '$($_.gender)'" -ForegroundColor Red
    }
    
    Write-Host "`n⚠️  These records cannot work with gender validation." -ForegroundColor Yellow
    Write-Host "Solution: Delete all employee records and start fresh (seed data will be recreated)" -ForegroundColor Cyan
} else {
    Write-Host "`n✅ All employees have valid gender values!" -ForegroundColor Green
}

Write-Host "`n========== SUMMARY ==========" -ForegroundColor Cyan
Write-Host "Total employees: $($allEmps.Count)" -ForegroundColor Cyan
Write-Host "Valid (with gender): $($allEmps.Count - $badRecords.Count)" -ForegroundColor Green
Write-Host "Invalid (no gender): $($badRecords.Count)" -ForegroundColor Red
