Write-Host "=== LEAVE MANAGEMENT ENDPOINTS TEST ===" -ForegroundColor Cyan

$baseURL = "http://localhost:8080"

# First, login as admin
Write-Host "`nStep 1: Login as admin..." -ForegroundColor Yellow
$adminLogin = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

try {
    $adminResp = Invoke-WebRequest -Uri "$baseURL/auth/login" `
        -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $adminLogin
    $adminJson = $adminResp.Content | ConvertFrom-Json
    $adminToken = $adminJson.token
    Write-Host "Admin login successful"
} catch {
    Write-Host "Admin login failed: $($_.Exception.Message)"
    exit 1
}

# Step 2: Test /leave/all endpoint
Write-Host "`nStep 2: Testing GET /leave/all (no filter)..." -ForegroundColor Yellow
try {
    $allResp = Invoke-WebRequest -Uri "$baseURL/leave/all" `
        -Method GET `
        -Headers @{
            "Content-Type" = "application/json"
            "Authorization" = "Bearer $adminToken"
        }
    Write-Host "Status: $($allResp.StatusCode) - SUCCESS"
    $allJson = $allResp.Content | ConvertFrom-Json
    Write-Host "Total leave requests: $($allJson.data.count)"
} catch {
    Write-Host "Status: $($_.Exception.Response.StatusCode) - FAILED"
    Write-Host "Error: $($_.Exception.Message)"
}

# Step 3: Test /leave/all with status filter
Write-Host "`nStep 3: Testing GET /leave/all?status=PENDING..." -ForegroundColor Yellow
try {
    $pendingResp = Invoke-WebRequest -Uri "$baseURL/leave/all?status=PENDING" `
        -Method GET `
        -Headers @{
            "Content-Type" = "application/json"
            "Authorization" = "Bearer $adminToken"
        }
    Write-Host "Status: $($pendingResp.StatusCode) - SUCCESS"
    $pendingJson = $pendingResp.Content | ConvertFrom-Json
    Write-Host "Pending requests: $($pendingJson.data.count)"
    if ($pendingJson.data.count -gt 0) {
        $firstReq = $pendingJson.data.leave_requests[0]
        Write-Host "First request: ID=$($firstReq.id), Type=$($firstReq.leave_type), Days=$($firstReq.days_count)"
        $leaveId = $firstReq.id
    }
} catch {
    Write-Host "Status: $($_.Exception.Response.StatusCode) - FAILED"
    Write-Host "Error: $($_.Exception.Message)"
}

# Step 4: Test /leave/approve/:id
if ($leaveId) {
    Write-Host "`nStep 4: Testing POST /leave/approve/$leaveId..." -ForegroundColor Yellow
    $approvePayload = @{
        notes = "Approved by admin"
    } | ConvertTo-Json
    
    try {
        $approveResp = Invoke-WebRequest -Uri "$baseURL/leave/approve/$leaveId" `
            -Method POST `
            -Headers @{
                "Content-Type" = "application/json"
                "Authorization" = "Bearer $adminToken"
            } `
            -Body $approvePayload
        Write-Host "Status: $($approveResp.StatusCode) - APPROVED"
        $approveJson = $approveResp.Content | ConvertFrom-Json
        Write-Host "Message: $($approveJson.message)"
    } catch {
        Write-Host "Status: $($_.Exception.Response.StatusCode) - FAILED"
        $errBody = $_.Exception.Response.Content.ReadAsStream() | ForEach-Object { [System.IO.StreamReader]::new($_).ReadToEnd() }
        Write-Host "Error: $errBody"
    }
} else {
    Write-Host "`nStep 4: Skipping approve test (no pending requests)"
}

# Step 5: Apply a new leave to test reject
Write-Host "`nStep 5: Applying new leave to test reject..." -ForegroundColor Yellow
$empLogin = @{
    username = "john_doe"
    password = "john123"
} | ConvertTo-Json

try {
    $empResp = Invoke-WebRequest -Uri "$baseURL/auth/login" `
        -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $empLogin
    $empJson = $empResp.Content | ConvertFrom-Json
    $empToken = $empJson.token
    
    $leavePayload = @{
        leave_type = "CASUAL"
        start_date = "2025-12-30"
        end_date = "2025-12-31"
        reason = "Testing reject endpoint"
    } | ConvertTo-Json
    
    $applyResp = Invoke-WebRequest -Uri "$baseURL/leave/apply" `
        -Method POST `
        -Headers @{
            "Content-Type" = "application/json"
            "Authorization" = "Bearer $empToken"
        } `
        -Body $leavePayload
    $applyJson = $applyResp.Content | ConvertFrom-Json
    $rejectLeaveId = $applyJson.data.id
    Write-Host "Leave applied with ID: $rejectLeaveId"
} catch {
    Write-Host "Failed to apply leave: $($_.Exception.Message)"
}

# Step 6: Test /leave/reject/:id
if ($rejectLeaveId) {
    Write-Host "`nStep 6: Testing POST /leave/reject/$rejectLeaveId..." -ForegroundColor Yellow
    $rejectPayload = @{
        reason = "Not meeting requirements"
    } | ConvertTo-Json
    
    try {
        $rejectResp = Invoke-WebRequest -Uri "$baseURL/leave/reject/$rejectLeaveId" `
            -Method POST `
            -Headers @{
                "Content-Type" = "application/json"
                "Authorization" = "Bearer $adminToken"
            } `
            -Body $rejectPayload
        Write-Host "Status: $($rejectResp.StatusCode) - REJECTED"
        $rejectJson = $rejectResp.Content | ConvertFrom-Json
        Write-Host "Message: $($rejectJson.message)"
    } catch {
        Write-Host "Status: $($_.Exception.Response.StatusCode) - FAILED"
        $errBody = $_.Exception.Response.Content.ReadAsStream() | ForEach-Object { [System.IO.StreamReader]::new($_).ReadToEnd() }
        Write-Host "Error: $errBody"
    }
}

Write-Host "`n=== TEST COMPLETE ===" -ForegroundColor Green
