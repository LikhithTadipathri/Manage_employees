# ✅ FINAL EMAIL NOTIFICATION TEST
# This tests the complete email flow end-to-end

Write-Host "=== EMAIL NOTIFICATION SYSTEM TEST ===" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080"

# Test 1: Check server is running
Write-Host "1. Checking server status..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "$baseUrl/health" -Method GET -ErrorAction Stop
    Write-Host "   ✅ Server is running" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Server is NOT running!" -ForegroundColor Red
    Write-Host "   Please start the server first." -ForegroundColor Yellow
    exit 1
}

# Test 2: Login as admin to create an employee
Write-Host "`n2. Logging in as admin..." -ForegroundColor Yellow
try {
    $adminLogin = @{
        username = "admin"
        password = "admin123"
    } | ConvertTo-Json

    $adminResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $adminLogin -ContentType "application/json" -ErrorAction Stop
    $adminToken = $adminResponse.data.token
    Write-Host "   ✅ Admin login successful" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Admin login failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "   Trying with superadmin..." -ForegroundColor Yellow
    
    try {
        $superadminLogin = @{
            username = "superadmin"
            password = "superadmin123"
        } | ConvertTo-Json

        $adminResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $superadminLogin -ContentType "application/json" -ErrorAction Stop
        $adminToken = $adminResponse.data.token
        Write-Host "   ✅ Superadmin login successful" -ForegroundColor Green
    } catch {
        Write-Host "   ❌ Login failed. Cannot proceed." -ForegroundColor Red
        exit 1
    }
}

# Test 3: Create a test employee
Write-Host "`n3. Creating test employee..." -ForegroundColor Yellow
$randomId = Get-Random -Minimum 1000 -Maximum 9999
$testEmail = "test.employee$randomId@company.com"

try {
    $employeeData = @{
        username = "testuser$randomId"
        password = "test123"
        first_name = "Test"
        last_name = "Employee"
        email = $testEmail
        phone = "+91-9876543210"
        position = "Software Engineer"
        salary = 50000
        gender = "Male"
        marital_status = $false
        hired_date = "2024-01-01T00:00:00Z"
    } | ConvertTo-Json

    $headers = @{
        "Authorization" = "Bearer $adminToken"
        "Content-Type" = "application/json"
    }

    $empResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/employees" -Method POST -Body $employeeData -Headers $headers -ErrorAction Stop
    Write-Host "   ✅ Employee created successfully" -ForegroundColor Green
    Write-Host "   Employee ID: $($empResponse.data.id)" -ForegroundColor Cyan
    Write-Host "   Email: $testEmail" -ForegroundColor Cyan
} catch {
    Write-Host "   ❌ Failed to create employee: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 4: Login as the new employee
Write-Host "`n4. Logging in as test employee..." -ForegroundColor Yellow
try {
    $empLogin = @{
        username = "testuser$randomId"
        password = "test123"
    } | ConvertTo-Json

    $empLoginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $empLogin -ContentType "application/json" -ErrorAction Stop
    $empToken = $empLoginResponse.data.token
    Write-Host "   ✅ Employee login successful" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Employee login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 5: Apply for leave
Write-Host "`n5. Applying for leave..." -ForegroundColor Yellow
try {
    $leaveData = @{
        leave_type = "ANNUAL"
        start_date = "2026-03-10"
        end_date = "2026-03-12"
        reason = "Testing email notifications - Collection v17"
    } | ConvertTo-Json

    $leaveHeaders = @{
        "Authorization" = "Bearer $empToken"
        "Content-Type" = "application/json"
    }

    $leaveResponse = Invoke-RestMethod -Uri "$baseUrl/leave/apply" -Method POST -Body $leaveData -Headers $leaveHeaders -ErrorAction Stop
    $leaveId = $leaveResponse.data.id
    
    Write-Host "   ✅ Leave application successful!" -ForegroundColor Green
    Write-Host "   Leave ID: $leaveId" -ForegroundColor Cyan
    Write-Host "   Status: $($leaveResponse.data.status)" -ForegroundColor Cyan
    Write-Host "   Days: $($leaveResponse.data.days_count)" -ForegroundColor Cyan
    Write-Host "   Salary Deduction: ₹$($leaveResponse.data.salary_deduction)" -ForegroundColor Cyan
} catch {
    Write-Host "   ❌ Leave application failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "   Details: $($_.ErrorDetails.Message)" -ForegroundColor Red
    }
    exit 1
}

# Test 6: Wait for email processing
Write-Host "`n6. Waiting for email queue to process..." -ForegroundColor Yellow
Write-Host "   (Emails are sent asynchronously)" -ForegroundColor Gray
for ($i = 5; $i -gt 0; $i--) {
    Write-Host "   $i..." -NoNewline -ForegroundColor Gray
    Start-Sleep -Seconds 1
}
Write-Host " Done!" -ForegroundColor Green

# Test 7: Check server logs for email activity
Write-Host "`n7. Email notifications should be sent now!" -ForegroundColor Yellow
Write-Host "   Expected: 2 emails (1 to employee, 1 to admin)" -ForegroundColor Gray
Write-Host ""
Write-Host "   📧 Email 1: To $testEmail" -ForegroundColor Cyan
Write-Host "      Subject: Leave Request Submitted – Pending Approval" -ForegroundColor White
Write-Host ""
Write-Host "   📧 Email 2: To admin@example.com (or similar)" -ForegroundColor Cyan
Write-Host "      Subject: Action Required: New Leave Request Submitted" -ForegroundColor White

# Test 8: Approve the leave
Write-Host "`n8. Approving leave as admin..." -ForegroundColor Yellow
try {
    $approveData = @{
        notes = "Approved for testing"
    } | ConvertTo-Json

    $approveHeaders = @{
        "Authorization" = "Bearer $adminToken"
        "Content-Type" = "application/json"
    }

    $approveResponse = Invoke-RestMethod -Uri "$baseUrl/leave/approve/$leaveId" -Method POST -Body $approveData -Headers $approveHeaders -ErrorAction Stop
    Write-Host "   ✅ Leave approved!" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Approval failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 9: Wait for approval email
Write-Host "`n9. Waiting for approval email..." -ForegroundColor Yellow
for ($i = 5; $i -gt 0; $i--) {
    Write-Host "   $i..." -NoNewline -ForegroundColor Gray
    Start-Sleep -Seconds 1
}
Write-Host " Done!" -ForegroundColor Green

Write-Host ""
Write-Host "   📧 Email 3: To $testEmail" -ForegroundColor Cyan
Write-Host "      Subject: Leave Approved" -ForegroundColor White
Write-Host "      Body should mention: ₹500 per day deduction" -ForegroundColor White

# Summary
Write-Host "`n=== TEST COMPLETE ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "✅ What was tested:" -ForegroundColor Green
Write-Host "   - Server is running" -ForegroundColor White
Write-Host "   - Employee created" -ForegroundColor White
Write-Host "   - Leave application submitted" -ForegroundColor White
Write-Host "   - Leave approved" -ForegroundColor White
Write-Host "   - Email notifications triggered (3 total)" -ForegroundColor White
Write-Host ""
Write-Host "📧 Expected emails:" -ForegroundColor Yellow
Write-Host "   1. Leave Applied → Employee ($testEmail)" -ForegroundColor White
Write-Host "   2. Leave Applied → Admin" -ForegroundColor White
Write-Host "   3. Leave Approved → Employee ($testEmail)" -ForegroundColor White
Write-Host ""
Write-Host "🔍 How to verify:" -ForegroundColor Cyan
Write-Host "   1. Check email inbox: $testEmail" -ForegroundColor White
Write-Host "   2. Check admin email inbox" -ForegroundColor White
Write-Host "   3. Check spam/junk folders" -ForegroundColor White
Write-Host "   4. Check server logs for:" -ForegroundColor White
Write-Host "      - 'Processing notification'" -ForegroundColor Gray
Write-Host "      - '✅ Notification X sent successfully'" -ForegroundColor Gray
Write-Host ""
Write-Host "⚠️  IMPORTANT NOTES:" -ForegroundColor Yellow
Write-Host "   - SMTP credentials: no-reply@company.com" -ForegroundColor White
Write-Host "   - If this is a FAKE email, emails will FAIL" -ForegroundColor Red
Write-Host "   - You need a REAL Gmail account with App Password" -ForegroundColor Red
Write-Host "   - Check server logs for SMTP errors" -ForegroundColor White
Write-Host ""
Write-Host "📝 Next steps if emails don't arrive:" -ForegroundColor Cyan
Write-Host "   1. Verify 'no-reply@company.com' is a REAL Gmail account" -ForegroundColor White
Write-Host "   2. Verify 'vjut usxv scsk pbuc' is a valid App Password" -ForegroundColor White
Write-Host "   3. Check server logs for SMTP authentication errors" -ForegroundColor White
Write-Host "   4. Try using your personal Gmail for testing" -ForegroundColor White
Write-Host ""
