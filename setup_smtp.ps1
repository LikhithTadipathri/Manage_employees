# Quick SMTP Setup Script for Windows
# This script helps you set up SMTP environment variables

Write-Host "=== SMTP Configuration Setup ===" -ForegroundColor Cyan
Write-Host ""

Write-Host "This script will help you configure SMTP settings for email notifications." -ForegroundColor Yellow
Write-Host ""

# Ask for SMTP provider
Write-Host "Which email provider are you using?" -ForegroundColor Cyan
Write-Host "1. Gmail (recommended for testing)" -ForegroundColor White
Write-Host "2. Outlook/Hotmail" -ForegroundColor White
Write-Host "3. Custom SMTP server" -ForegroundColor White
Write-Host ""

$provider = Read-Host "Enter choice (1-3)"

switch ($provider) {
    "1" {
        $smtpHost = "smtp.gmail.com"
        $smtpPort = "587"
        Write-Host "`n✅ Using Gmail SMTP settings" -ForegroundColor Green
        Write-Host ""
        Write-Host "⚠️  IMPORTANT: You MUST use a Gmail App Password, not your regular password!" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "How to get Gmail App Password:" -ForegroundColor Cyan
        Write-Host "1. Go to: https://myaccount.google.com/" -ForegroundColor White
        Write-Host "2. Click 'Security' in the left sidebar" -ForegroundColor White
        Write-Host "3. Enable '2-Step Verification' (if not already enabled)" -ForegroundColor White
        Write-Host "4. Go back to Security → 'App passwords'" -ForegroundColor White
        Write-Host "5. Select app: 'Mail', device: 'Windows Computer'" -ForegroundColor White
        Write-Host "6. Click 'Generate'" -ForegroundColor White
        Write-Host "7. Copy the 16-character password (no spaces)" -ForegroundColor White
        Write-Host ""
    }
    "2" {
        $smtpHost = "smtp-mail.outlook.com"
        $smtpPort = "587"
        Write-Host "`n✅ Using Outlook SMTP settings" -ForegroundColor Green
    }
    "3" {
        $smtpHost = Read-Host "Enter SMTP host (e.g., smtp.example.com)"
        $smtpPort = Read-Host "Enter SMTP port (usually 587 or 465)"
        Write-Host "`n✅ Using custom SMTP settings" -ForegroundColor Green
    }
    default {
        Write-Host "`n❌ Invalid choice. Exiting." -ForegroundColor Red
        exit 1
    }
}

# Get email credentials
Write-Host ""
$smtpUsername = Read-Host "Enter your email address (e.g., your-email@gmail.com)"
$smtpPassword = Read-Host "Enter your SMTP password (App Password for Gmail)" -AsSecureString
$smtpPasswordPlain = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($smtpPassword))

$smtpFromAddr = Read-Host "Enter 'From' email address (usually same as username) [$smtpUsername]"
if ([string]::IsNullOrWhiteSpace($smtpFromAddr)) {
    $smtpFromAddr = $smtpUsername
}

$smtpFromName = Read-Host "Enter 'From' name (e.g., HR Management System) [HR Management System]"
if ([string]::IsNullOrWhiteSpace($smtpFromName)) {
    $smtpFromName = "HR Management System"
}

$emailWorkers = Read-Host "Enter number of email queue workers [3]"
if ([string]::IsNullOrWhiteSpace($emailWorkers)) {
    $emailWorkers = "3"
}

# Summary
Write-Host ""
Write-Host "=== Configuration Summary ===" -ForegroundColor Cyan
Write-Host "SMTP Host:       $smtpHost" -ForegroundColor White
Write-Host "SMTP Port:       $smtpPort" -ForegroundColor White
Write-Host "SMTP Username:   $smtpUsername" -ForegroundColor White
Write-Host "SMTP Password:   ******** (hidden)" -ForegroundColor White
Write-Host "From Address:    $smtpFromAddr" -ForegroundColor White
Write-Host "From Name:       $smtpFromName" -ForegroundColor White
Write-Host "Queue Workers:   $emailWorkers" -ForegroundColor White
Write-Host ""

$confirm = Read-Host "Apply these settings? (Y/N)"
if ($confirm -ne "Y" -and $confirm -ne "y") {
    Write-Host "❌ Configuration cancelled." -ForegroundColor Red
    exit 0
}

# Set environment variables for current session
Write-Host ""
Write-Host "Setting environment variables for current session..." -ForegroundColor Yellow
$env:SMTP_HOST = $smtpHost
$env:SMTP_PORT = $smtpPort
$env:SMTP_USERNAME = $smtpUsername
$env:SMTP_PASSWORD = $smtpPasswordPlain
$env:SMTP_FROM_ADDR = $smtpFromAddr
$env:SMTP_FROM_NAME = $smtpFromName
$env:EMAIL_QUEUE_WORKERS = $emailWorkers

Write-Host "✅ Environment variables set for current session!" -ForegroundColor Green
Write-Host ""

# Offer to create a .env file
Write-Host "Would you like to save these settings to a .env file?" -ForegroundColor Cyan
Write-Host "This will allow the settings to persist across sessions." -ForegroundColor Gray
$saveEnv = Read-Host "(Y/N)"

if ($saveEnv -eq "Y" -or $saveEnv -eq "y") {
    $envContent = @"
# SMTP Email Configuration
SMTP_HOST=$smtpHost
SMTP_PORT=$smtpPort
SMTP_USERNAME=$smtpUsername
SMTP_PASSWORD=$smtpPasswordPlain
SMTP_FROM_ADDR=$smtpFromAddr
SMTP_FROM_NAME=$smtpFromName
EMAIL_QUEUE_WORKERS=$emailWorkers
"@

    $envContent | Out-File -FilePath ".env" -Encoding UTF8
    Write-Host "✅ Settings saved to .env file" -ForegroundColor Green
    Write-Host ""
    Write-Host "⚠️  NOTE: Make sure your application loads .env file on startup!" -ForegroundColor Yellow
    Write-Host "You may need to use a library like 'godotenv' in Go." -ForegroundColor Gray
}

# Offer to test SMTP
Write-Host ""
Write-Host "Would you like to test the SMTP configuration now?" -ForegroundColor Cyan
$testSmtp = Read-Host "(Y/N)"

if ($testSmtp -eq "Y" -or $testSmtp -eq "y") {
    Write-Host ""
    Write-Host "Running SMTP test..." -ForegroundColor Yellow
    Write-Host ""
    
    # Check if test_smtp_config.go exists
    if (Test-Path "test_smtp_config.go") {
        go run test_smtp_config.go
    } else {
        Write-Host "❌ test_smtp_config.go not found in current directory" -ForegroundColor Red
        Write-Host "Please run this script from the project root directory." -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "=== Next Steps ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. Start your server:" -ForegroundColor Yellow
Write-Host "   go run cmd/employee-service/main.go" -ForegroundColor White
Write-Host ""
Write-Host "2. Look for this message in the logs:" -ForegroundColor Yellow
Write-Host "   ✅ Email queue started with $emailWorkers workers" -ForegroundColor Green
Write-Host ""
Write-Host "3. Test by applying for leave via Postman/API" -ForegroundColor Yellow
Write-Host ""
Write-Host "4. Check your email inbox for notifications" -ForegroundColor Yellow
Write-Host ""
Write-Host "=== Troubleshooting ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "If emails don't work:" -ForegroundColor Yellow
Write-Host "- Check server logs for SMTP errors" -ForegroundColor White
Write-Host "- Verify Gmail App Password is correct (for Gmail users)" -ForegroundColor White
Write-Host "- Check spam/junk folder" -ForegroundColor White
Write-Host "- Run: go run test_smtp_config.go" -ForegroundColor White
Write-Host ""
Write-Host "For more help, see:" -ForegroundColor Yellow
Write-Host "- EMAIL_INVESTIGATION_SUMMARY.md" -ForegroundColor White
Write-Host "- EMAIL_NOTIFICATION_FIX.md" -ForegroundColor White
Write-Host ""
Write-Host "✅ Setup complete!" -ForegroundColor Green
