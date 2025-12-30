# Simple test for leave endpoints
$baseURL = "http://localhost:8080"

# Login as admin
Write-Host "Logging in as admin..."
$loginResp = Invoke-WebRequest -Uri "$baseURL/auth/login" `
    -Method POST `
    -Headers @{"Content-Type" = "application/json"} `
    -Body '{"username":"admin","password":"admin123"}'

$token = ($loginResp.Content | ConvertFrom-Json).token
Write-Host "Token: $($token.Substring(0,30))..."

# Test 1: /leave/all without filter
Write-Host "`n=== Test 1: GET /leave/all (no filter) ===" 
try {
    $resp = Invoke-WebRequest -Uri "$baseURL/leave/all" `
        -Method GET `
        -Headers @{"Authorization" = "Bearer $token"}
    Write-Host "Status: $($resp.StatusCode)"
    $resp.Content | ConvertFrom-Json
} catch {
    Write-Host "ERROR: $($_.Exception.Message)"
    Write-Host "StatusCode: $($_.Exception.Response.StatusCode)"
}

# Test 2: /leave/all with PENDING filter
Write-Host "`n=== Test 2: GET /leave/all?status=PENDING ===" 
try {
    $resp = Invoke-WebRequest -Uri "$baseURL/leave/all?status=PENDING" `
        -Method GET `
        -Headers @{"Authorization" = "Bearer $token"}
    Write-Host "Status: $($resp.StatusCode)"
    $resp.Content | ConvertFrom-Json
} catch {
    Write-Host "ERROR: $($_.Exception.Message)"
    Write-Host "StatusCode: $($_.Exception.Response.StatusCode)"
}

# Test 3: Apply leave as employee then approve it
Write-Host "`n=== Test 3: Apply leave as employee ===" 
$empLogin = Invoke-WebRequest -Uri "$baseURL/auth/login" `
    -Method POST `
    -Headers @{"Content-Type" = "application/json"} `
    -Body '{"username":"john_doe","password":"john123"}'

$empToken = ($empLogin.Content | ConvertFrom-Json).token

$applyResp = Invoke-WebRequest -Uri "$baseURL/leave/apply" `
    -Method POST `
    -Headers @{
        "Content-Type" = "application/json"
        "Authorization" = "Bearer $empToken"
    } `
    -Body '{"leave_type":"CASUAL","start_date":"2025-12-27","end_date":"2025-12-29","reason":"Test"}'

$leaveId = ($applyResp.Content | ConvertFrom-Json).data.id
Write-Host "Applied leave with ID: $leaveId"

# Test 4: /leave/approve
Write-Host "`n=== Test 4: POST /leave/approve/$leaveId ===" 
try {
    $approveResp = Invoke-WebRequest -Uri "$baseURL/leave/approve/$leaveId" `
        -Method POST `
        -Headers @{
            "Content-Type" = "application/json"
            "Authorization" = "Bearer $token"
        } `
        -Body '{"notes":"Approved"}'
    Write-Host "Status: $($approveResp.StatusCode)"
    $approveResp.Content | ConvertFrom-Json
} catch {
    Write-Host "ERROR: $($_.Exception.Message)"
    Write-Host "StatusCode: $($_.Exception.Response.StatusCode)"
}

# Test 5: Apply another leave and reject it
Write-Host "`n=== Test 5: Apply another leave to reject ===" 
$applyResp2 = Invoke-WebRequest -Uri "$baseURL/leave/apply" `
    -Method POST `
    -Headers @{
        "Content-Type" = "application/json"
        "Authorization" = "Bearer $empToken"
    } `
    -Body '{"leave_type":"PERSONAL","start_date":"2026-01-15","end_date":"2026-01-17","reason":"Test reject"}'

$leaveId2 = ($applyResp2.Content | ConvertFrom-Json).data.id
Write-Host "Applied leave with ID: $leaveId2"

# Test 6: /leave/reject
Write-Host "`n=== Test 6: POST /leave/reject/$leaveId2 ===" 
try {
    $rejectResp = Invoke-WebRequest -Uri "$baseURL/leave/reject/$leaveId2" `
        -Method POST `
        -Headers @{
            "Content-Type" = "application/json"
            "Authorization" = "Bearer $token"
        } `
        -Body '{"reason":"Rejected by admin"}'
    Write-Host "Status: $($rejectResp.StatusCode)"
    $rejectResp.Content | ConvertFrom-Json
} catch {
    Write-Host "ERROR: $($_.Exception.Message)"
    Write-Host "StatusCode: $($_.Exception.Response.StatusCode)"
}
