# Cleanup PostgreSQL database using PowerShell and .NET
# Delete all records and reset sequences

Write-Host "`n" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "  CLEANING UP DATABASE - DELETING ALL RECORDS" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "`n"

# PostgreSQL connection
$host = "localhost"
$port = 5432
$user = "postgres"
$password = "tiger"
$database = "employee_db"

# Create connection string
$connectionString = "Host=$host;Port=$port;Username=$user;Password=$password;Database=$database"

try {
    # Load Npgsql provider
    Add-Type -Path "C:\Program Files\PostgreSQL\16\lib\Npgsql.dll" -ErrorAction SilentlyContinue
    
    $connection = New-Object Npgsql.NpgsqlConnection($connectionString)
    $connection.Open()
    
    Write-Host "✓ Connected to PostgreSQL" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Could not connect to PostgreSQL" -ForegroundColor Red
    Write-Host "Make sure PostgreSQL is running at localhost:5432" -ForegroundColor Yellow
    Write-Host "Connection string: $connectionString" -ForegroundColor Yellow
    exit 1
}

# Delete records in order (respecting foreign keys)
$deleteTables = @(
    @{ name = "leave_requests"; query = "DELETE FROM leave_requests" },
    @{ name = "employees"; query = "DELETE FROM employees" },
    @{ name = "users"; query = "DELETE FROM users" }
)

try {
    foreach ($table in $deleteTables) {
        Write-Host "Deleting $($table.name)..." -NoNewline
        
        $command = $connection.CreateCommand()
        $command.CommandText = $table.query
        $rows = $command.ExecuteNonQuery()
        
        Write-Host " ✓ Deleted $rows records" -ForegroundColor Green
    }
    
    # Reset sequences
    Write-Host "`nResetting PostgreSQL sequences..." -ForegroundColor Yellow
    
    $sequences = @(
        "ALTER SEQUENCE IF EXISTS leave_requests_id_seq RESTART WITH 1",
        "ALTER SEQUENCE IF EXISTS employees_id_seq RESTART WITH 1",
        "ALTER SEQUENCE IF EXISTS users_id_seq RESTART WITH 1"
    )
    
    foreach ($seq in $sequences) {
        $command = $connection.CreateCommand()
        $command.CommandText = $seq
        try {
            $command.ExecuteNonQuery() | Out-Null
            Write-Host "✓ $($seq.Split('_')[2])" -ForegroundColor Green
        } catch {
            # Sequence might not exist, that's OK
        }
    }
    
    # Verify deletion
    Write-Host "`n📊 VERIFICATION:" -ForegroundColor Cyan
    
    $verifyQueries = @(
        @{ table = "users"; query = "SELECT COUNT(*) FROM users" },
        @{ table = "employees"; query = "SELECT COUNT(*) FROM employees" },
        @{ table = "leave_requests"; query = "SELECT COUNT(*) FROM leave_requests" }
    )
    
    foreach ($verify in $verifyQueries) {
        $command = $connection.CreateCommand()
        $command.CommandText = $verify.query
        $count = $command.ExecuteScalar()
        Write-Host "✓ $($verify.table): $count records" -ForegroundColor Green
    }
    
    $connection.Close()
    
    Write-Host "`n" + ("=" * 50) -ForegroundColor Cyan
    Write-Host "✅ DATABASE CLEANUP COMPLETE - READY FOR FRESH START" -ForegroundColor Green
    Write-Host ("=" * 50) -ForegroundColor Cyan
    Write-Host "`nNext steps:" -ForegroundColor Yellow
    Write-Host "1. Stop current server (if running)" -ForegroundColor Yellow
    Write-Host "2. Restart: cd d:\Go\src\Task\cmd\employee-service" -ForegroundColor Yellow
    Write-Host "3. Run: go run main.go" -ForegroundColor Yellow
    Write-Host "4. Server will auto-seed superadmin account" -ForegroundColor Yellow
    Write-Host "5. Start fresh in Postman!" -ForegroundColor Yellow
    Write-Host "`n"
    
} catch {
    Write-Host "ERROR: $_" -ForegroundColor Red
    $connection.Close()
    exit 1
}
