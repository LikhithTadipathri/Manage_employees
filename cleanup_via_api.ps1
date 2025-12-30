# Delete all employees via API to clear problematic records
# This will remove all employee records and leave data
# Seed data will be recreated on next server restart

$BASE_URL = "http://localhost:8080"

Write-Host "========== EMPLOYEE DATA CLEANUP ==========" -ForegroundColor Cyan
Write-Host "This will DELETE all employee records and associated leave data" -ForegroundColor Yellow
Write-Host "Seed data will be recreated automatically on next server restart`n" -ForegroundColor Yellow

# Login as admin
Write-Host "1. Logging in as admin..." -ForegroundColor Yellow
$adminLogin = Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method POST -ContentType "application/json" -Body (@{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json)

$adminToken = $adminLogin.data.token
$headers = @{
    Authorization = "Bearer $adminToken"
    "Content-Type" = "application/json"
}

Write-Host "✅ Admin logged in`n" -ForegroundColor Green

# Get all employees
Write-Host "2. Fetching all employees..." -ForegroundColor Yellow
$empResponse = Invoke-RestMethod -Uri "$BASE_URL/api/v1/employees?limit=1000" -Method GET -Headers $headers
$allEmps = $empResponse.data

if ($null -eq $allEmps -or $allEmps.Count -eq 0) {
    Write-Host "✅ No employees to delete - database is already clean!`n" -ForegroundColor Green
} else {
    Write-Host "Found $($allEmps.Count) employees`n" -ForegroundColor Cyan
    
    # Delete each employee (cascade will handle leave requests)
    Write-Host "3. Deleting all employees..." -ForegroundColor Yellow
    $deleted = 0
    $failed = 0
    
    foreach ($emp in $allEmps) {
        try {
            $deleteUrl = "$BASE_URL/api/v1/employees/$($emp.id)"
            Invoke-RestMethod -Uri $deleteUrl -Method DELETE -Headers $headers | Out-Null
            $deleted++
            Write-Host "   ✅ Deleted: $($emp.first_name) $($emp.last_name) (ID: $($emp.id))" -ForegroundColor Green
        } catch {
            $failed++
            Write-Host "   ⚠️  Failed to delete: $($emp.first_name) $($emp.last_name)" -ForegroundColor Yellow
        }
    }
    
    Write-Host "`n========== CLEANUP SUMMARY ==========" -ForegroundColor Cyan
    Write-Host "Deleted: $deleted employees" -ForegroundColor Green
    Write-Host "Failed: $failed employees" -ForegroundColor $(if ($failed -gt 0) { "Yellow" } else { "Green" })
}

Write-Host "`nNext steps:" -ForegroundColor Cyan
Write-Host "1. Restart the server (it will auto-seed with jane_smith as Female+Married)" -ForegroundColor Cyan
Write-Host "2. All existing invalid records are now gone" -ForegroundColor Cyan
Write-Host "3. New female employees can apply for maternity leave" -ForegroundColor Cyan
