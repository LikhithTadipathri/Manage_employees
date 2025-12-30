# Test Email Notification Flow
# This script tests if emails are sent when leave is applied

Write-Host "=== Testing Email Notification Flow ===" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080"

# Step 1: Login as employee (john)
Write-Host "1. Logging in as employee (john)..." -ForegroundColor Yellow
try {
    $loginBody = @{
        username = "john"
        password = "password123"
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json" -ErrorAction Stop
    $token = $loginResponse.data.token
    Write-Host "   ✅ Login successful" -ForegroundColor Green
    Write-Host "   Token: $($token.Substring(0,20))..." -ForegroundColor Gray
} catch {
    Write-Host "   ❌ Login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 2: Apply for leave
Write-Host "`n2. Applying for leave..." -ForegroundColor Yellow
try {
    $leaveBody = @{
        leave_type = "ANNUAL"
        start_date = "2026-02-10"
        end_date = "2026-02-12"
        reason = "Testing email notifications - Collection v17"
    } | ConvertTo-Json

    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

    $leaveResponse = Invoke-RestMethod -Uri "$baseUrl/leave/apply" -Method POST -Body $leaveBody -Headers $headers -ErrorAction Stop
    Write-Host "   ✅ Leave application successful!" -ForegroundColor Green
    Write-Host "   Leave Request ID: $($leaveResponse.data.id)" -ForegroundColor Cyan
    Write-Host "   Status: $($leaveResponse.data.status)" -ForegroundColor Cyan
    Write-Host "   Days: $($leaveResponse.data.days_count)" -ForegroundColor Cyan
    Write-Host "   Salary Deduction: ₹$($leaveResponse.data.salary_deduction)" -ForegroundColor Cyan
    
    $leaveId = $leaveResponse.data.id
} catch {
    Write-Host "   ❌ Leave application failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "   Details: $($_.ErrorDetails.Message)" -ForegroundColor Red
    }
    exit 1
}

# Step 3: Wait for email queue to process
Write-Host "`n3. Waiting for email queue to process..." -ForegroundColor Yellow
Write-Host "   (Emails are sent asynchronously)" -ForegroundColor Gray
Start-Sleep -Seconds 5

# Step 4: Check database for notifications
Write-Host "`n4. Checking notifications in database..." -ForegroundColor Yellow
$dbFile = "employee.db"
if (Test-Path $dbFile) {
    try {
        $sqliteCheck = Get-Command sqlite3 -ErrorAction SilentlyContinue
        if ($sqliteCheck) {
            Write-Host "`n   Recent notifications:" -ForegroundColor Cyan
            $query = "SELECT id, event_type, recipient_email, status, subject FROM notifications WHERE leave_request_id = $leaveId ORDER BY created_at DESC;"
            $notifications = sqlite3 $dbFile $query
            
            if ($notifications) {
                Write-Host "   $notifications" -ForegroundColor White
                
                # Count by status
                $sentCount = (sqlite3 $dbFile "SELECT COUNT(*) FROM notifications WHERE leave_request_id = $leaveId AND status='SENT';")
                $pendingCount = (sqlite3 $dbFile "SELECT COUNT(*) FROM notifications WHERE leave_request_id = $leaveId AND status='PENDING';")
                $failedCount = (sqlite3 $dbFile "SELECT COUNT(*) FROM notifications WHERE leave_request_id = $leaveId AND status='FAILED';")
                
                Write-Host "`n   Notification Status:" -ForegroundColor Cyan
                Write-Host "   - Sent: $sentCount" -ForegroundColor Green
                Write-Host "   - Pending: $pendingCount" -ForegroundColor Yellow
                Write-Host "   - Failed: $failedCount" -ForegroundColor Red
                
                if ([int]$failedCount -gt 0) {
                    Write-Host "`n   ❌ Some notifications FAILED. Checking errors..." -ForegroundColor Red
                    $errors = sqlite3 $dbFile "SELECT recipient_email, error_message FROM notifications WHERE leave_request_id = $leaveId AND status='FAILED';"
                    Write-Host "   $errors" -ForegroundColor Red
                }
                
                if ([int]$sentCount -eq 2) {
                    Write-Host "`n   ✅ SUCCESS! Both emails sent (employee + admin)" -ForegroundColor Green
                } elseif ([int]$pendingCount -gt 0) {
                    Write-Host "`n   ⏳ Emails are still pending. Wait a bit longer..." -ForegroundColor Yellow
                } else {
                    Write-Host "`n   ⚠️  Unexpected status. Check server logs." -ForegroundColor Yellow
                }
            } else {
                Write-Host "   ❌ No notifications found for this leave request!" -ForegroundColor Red
                Write-Host "   This means email notification code was NOT triggered." -ForegroundColor Yellow
            }
        } else {
            Write-Host "   ⚠️  sqlite3 not available. Cannot check database." -ForegroundColor Yellow
            Write-Host "   Install SQLite tools to view notifications." -ForegroundColor Gray
        }
    } catch {
        Write-Host "   ❌ Error checking database: $_" -ForegroundColor Red
    }
} else {
    Write-Host "   ❌ Database file not found: $dbFile" -ForegroundColor Red
}

# Step 5: Test approval flow
Write-Host "`n5. Would you like to test approval email? (Y/N)" -ForegroundColor Cyan
$testApproval = Read-Host

if ($testApproval -eq "Y" -or $testApproval -eq "y") {
    # Login as admin
    Write-Host "`n   5a. Logging in as admin..." -ForegroundColor Yellow
    try {
        $adminLoginBody = @{
            username = "admin"
            password = "admin123"
        } | ConvertTo-Json

        $adminLoginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $adminLoginBody -ContentType "application/json" -ErrorAction Stop
        $adminToken = $adminLoginResponse.data.token
        Write-Host "   ✅ Admin login successful" -ForegroundColor Green
    } catch {
        Write-Host "   ❌ Admin login failed: $($_.Exception.Message)" -ForegroundColor Red
        exit 1
    }

    # Approve the leave
    Write-Host "`n   5b. Approving leave request #$leaveId..." -ForegroundColor Yellow
    try {
        $approveBody = @{
            notes = "Approved for testing email notifications"
        } | ConvertTo-Json

        $adminHeaders = @{
            "Authorization" = "Bearer $adminToken"
            "Content-Type" = "application/json"
        }

        $approveResponse = Invoke-RestMethod -Uri "$baseUrl/leave/approve/$leaveId" -Method POST -Body $approveBody -Headers $adminHeaders -ErrorAction Stop
        Write-Host "   ✅ Leave approved!" -ForegroundColor Green
    } catch {
        Write-Host "   ❌ Approval failed: $($_.Exception.Message)" -ForegroundColor Red
    }

    # Wait and check
    Write-Host "`n   5c. Waiting for approval email..." -ForegroundColor Yellow
    Start-Sleep -Seconds 5

    if (Test-Path $dbFile) {
        $sqliteCheck = Get-Command sqlite3 -ErrorAction SilentlyContinue
        if ($sqliteCheck) {
            $approvalNotif = sqlite3 $dbFile "SELECT event_type, recipient_email, status FROM notifications WHERE leave_request_id = $leaveId AND event_type = 'LEAVE_APPROVED';"
            if ($approvalNotif) {
                Write-Host "   ✅ Approval notification: $approvalNotif" -ForegroundColor Green
            } else {
                Write-Host "   ❌ No approval notification found" -ForegroundColor Red
            }
        }
    }
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Check your email inbox (no-reply@company.com)" -ForegroundColor White
Write-Host "2. Check spam/junk folder if not in inbox" -ForegroundColor White
Write-Host "3. Check server logs for detailed SMTP activity" -ForegroundColor White
Write-Host ""
