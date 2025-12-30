# Email Notification Diagnostic Script
# This script helps diagnose why emails are not being sent

Write-Host "=== Email Notification Diagnostic Tool ===" -ForegroundColor Cyan
Write-Host ""

# Step 1: Check if server is running
Write-Host "1. Checking if server is running..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET -ErrorAction Stop
    Write-Host "   ✅ Server is running" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Server is NOT running. Please start the server first." -ForegroundColor Red
    Write-Host "   Run: go run cmd/employee-service/main.go" -ForegroundColor Yellow
    exit 1
}

# Step 2: Check environment variables
Write-Host "`n2. Checking SMTP environment variables..." -ForegroundColor Yellow
$smtpVars = @("SMTP_HOST", "SMTP_PORT", "SMTP_USERNAME", "SMTP_PASSWORD", "SMTP_FROM_ADDR", "SMTP_FROM_NAME")
$missingVars = @()

foreach ($var in $smtpVars) {
    $value = [Environment]::GetEnvironmentVariable($var)
    if ([string]::IsNullOrEmpty($value)) {
        Write-Host "   ❌ $var is NOT set" -ForegroundColor Red
        $missingVars += $var
    } else {
        if ($var -eq "SMTP_PASSWORD") {
            Write-Host "   ✅ $var is set (hidden)" -ForegroundColor Green
        } else {
            Write-Host "   ✅ $var = $value" -ForegroundColor Green
        }
    }
}

if ($missingVars.Count -gt 0) {
    Write-Host "`n   ⚠️  WARNING: Missing SMTP configuration!" -ForegroundColor Yellow
    Write-Host "   The server will use default values which may not work." -ForegroundColor Yellow
    Write-Host "`n   To fix, create a .env file or set environment variables:" -ForegroundColor Cyan
    Write-Host "   SMTP_HOST=smtp.gmail.com" -ForegroundColor White
    Write-Host "   SMTP_PORT=587" -ForegroundColor White
    Write-Host "   SMTP_USERNAME=your-email@gmail.com" -ForegroundColor White
    Write-Host "   SMTP_PASSWORD=your-app-password" -ForegroundColor White
    Write-Host "   SMTP_FROM_ADDR=your-email@gmail.com" -ForegroundColor White
    Write-Host "   SMTP_FROM_NAME=HR Management System" -ForegroundColor White
}

# Step 3: Check database for notifications table
Write-Host "`n3. Checking if notifications table exists..." -ForegroundColor Yellow
$dbFile = "employee.db"
if (Test-Path $dbFile) {
    try {
        # Check if sqlite3 is available
        $sqliteCheck = Get-Command sqlite3 -ErrorAction SilentlyContinue
        if ($sqliteCheck) {
            $tableCheck = sqlite3 $dbFile "SELECT name FROM sqlite_master WHERE type='table' AND name='notifications';"
            if ($tableCheck -eq "notifications") {
                Write-Host "   ✅ Notifications table exists" -ForegroundColor Green
            } else {
                Write-Host "   ❌ Notifications table NOT found" -ForegroundColor Red
                Write-Host "   The table should be created automatically on server start." -ForegroundColor Yellow
            }
        } else {
            Write-Host "   ⚠️  sqlite3 command not found, skipping table check" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "   ⚠️  Could not check database: $_" -ForegroundColor Yellow
    }
} else {
    Write-Host "   ⚠️  Database file not found at: $dbFile" -ForegroundColor Yellow
}

# Step 4: Test leave application and check notifications
Write-Host "`n4. Testing leave application flow..." -ForegroundColor Yellow
Write-Host "   This will apply a leave and check if notifications are created." -ForegroundColor Gray

# First, login as employee
Write-Host "`n   4a. Logging in as employee (john)..." -ForegroundColor Gray
try {
    $loginBody = @{
        username = "john"
        password = "password123"
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/auth/login" -Method POST -Body $loginBody -ContentType "application/json" -ErrorAction Stop
    $token = $loginResponse.data.token
    Write-Host "   ✅ Login successful" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Login failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "   Cannot proceed with leave application test." -ForegroundColor Yellow
    exit 1
}

# Apply leave
Write-Host "`n   4b. Applying leave request..." -ForegroundColor Gray
try {
    $leaveBody = @{
        leave_type = "ANNUAL"
        start_date = "2026-01-15"
        end_date = "2026-01-17"
        reason = "Testing email notifications"
    } | ConvertTo-Json

    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

    $leaveResponse = Invoke-RestMethod -Uri "http://localhost:8080/leave/apply" -Method POST -Body $leaveBody -Headers $headers -ErrorAction Stop
    Write-Host "   ✅ Leave application successful" -ForegroundColor Green
    Write-Host "   Leave Request ID: $($leaveResponse.data.id)" -ForegroundColor Cyan
    
    # Wait for async email processing
    Write-Host "`n   4c. Waiting 3 seconds for email queue to process..." -ForegroundColor Gray
    Start-Sleep -Seconds 3

} catch {
    Write-Host "   ❌ Leave application failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Step 5: Check notifications in database
Write-Host "`n5. Checking notifications in database..." -ForegroundColor Yellow
if (Test-Path $dbFile) {
    try {
        $sqliteCheck = Get-Command sqlite3 -ErrorAction SilentlyContinue
        if ($sqliteCheck) {
            Write-Host "`n   Recent notifications:" -ForegroundColor Cyan
            $notifications = sqlite3 $dbFile "SELECT id, event_type, recipient_email, status, created_at FROM notifications ORDER BY created_at DESC LIMIT 5;" -separator " | "
            if ($notifications) {
                Write-Host "   ID | Event Type | Recipient | Status | Created At" -ForegroundColor White
                Write-Host "   $notifications" -ForegroundColor White
                
                # Check for failed notifications
                $failedCount = sqlite3 $dbFile "SELECT COUNT(*) FROM notifications WHERE status='FAILED';"
                $pendingCount = sqlite3 $dbFile "SELECT COUNT(*) FROM notifications WHERE status='PENDING';"
                $sentCount = sqlite3 $dbFile "SELECT COUNT(*) FROM notifications WHERE status='SENT';"
                
                Write-Host "`n   Statistics:" -ForegroundColor Cyan
                Write-Host "   - Sent: $sentCount" -ForegroundColor Green
                Write-Host "   - Pending: $pendingCount" -ForegroundColor Yellow
                Write-Host "   - Failed: $failedCount" -ForegroundColor Red
                
                if ([int]$failedCount -gt 0) {
                    Write-Host "`n   ⚠️  There are FAILED notifications. Checking error messages..." -ForegroundColor Yellow
                    $errors = sqlite3 $dbFile "SELECT id, error_message FROM notifications WHERE status='FAILED' LIMIT 3;" -separator " | "
                    Write-Host "   $errors" -ForegroundColor Red
                }
                
                if ([int]$pendingCount -gt 0) {
                    Write-Host "`n   ⚠️  There are PENDING notifications. They may still be processing..." -ForegroundColor Yellow
                    Write-Host "   Check server logs for email queue activity." -ForegroundColor Gray
                }
            } else {
                Write-Host "   ❌ No notifications found in database!" -ForegroundColor Red
                Write-Host "   This means the email notification code is NOT being triggered." -ForegroundColor Yellow
            }
        } else {
            Write-Host "   ⚠️  sqlite3 not available, cannot check notifications" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "   ❌ Error checking notifications: $_" -ForegroundColor Red
    }
} else {
    Write-Host "   ❌ Database file not found" -ForegroundColor Red
}

# Step 6: Recommendations
Write-Host "`n=== RECOMMENDATIONS ===" -ForegroundColor Cyan
Write-Host ""

if ($missingVars.Count -gt 0) {
    Write-Host "1. ⚠️  Configure SMTP settings:" -ForegroundColor Yellow
    Write-Host "   - Set environment variables for SMTP" -ForegroundColor White
    Write-Host "   - Or create a .env file (if using godotenv)" -ForegroundColor White
    Write-Host "   - Restart the server after configuration" -ForegroundColor White
    Write-Host ""
}

Write-Host "2. 📋 Check server logs:" -ForegroundColor Yellow
Write-Host "   - Look for 'Email queue started' message" -ForegroundColor White
Write-Host "   - Look for 'Processing notification' messages" -ForegroundColor White
Write-Host "   - Look for any SMTP errors" -ForegroundColor White
Write-Host ""

Write-Host "3. 🔍 For Gmail users:" -ForegroundColor Yellow
Write-Host "   - Use App Password, not regular password" -ForegroundColor White
Write-Host "   - Enable 2FA on Gmail account" -ForegroundColor White
Write-Host "   - Generate App Password: Google Account → Security → App passwords" -ForegroundColor White
Write-Host ""

Write-Host "4. 🧪 Test SMTP connection:" -ForegroundColor Yellow
Write-Host "   - Use telnet or online SMTP tester" -ForegroundColor White
Write-Host "   - Verify port 587 is not blocked by firewall" -ForegroundColor White
Write-Host ""

Write-Host "=== END OF DIAGNOSTIC ===" -ForegroundColor Cyan
