@echo off
REM Cleanup PostgreSQL database
REM Remove all records and reset sequences

setlocal enabledelayedexpansion

echo.
echo ================================================
echo  CLEANING UP DATABASE - DELETING ALL RECORDS
echo ================================================
echo.

REM PostgreSQL connection parameters
set PGHOST=localhost
set PGPORT=5432
set PGUSER=postgres
set PGPASSWORD=tiger
set PGDATABASE=employee_db

REM Run cleanup commands
echo Deleting leave_requests...
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -c "DELETE FROM leave_requests;" || goto error

echo Deleting employees...
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -c "DELETE FROM employees;" || goto error

echo Deleting users...
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -c "DELETE FROM users;" || goto error

echo.
echo Resetting sequences...
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -c "ALTER SEQUENCE IF EXISTS leave_requests_id_seq RESTART WITH 1;" || echo Sequence might not exist
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -c "ALTER SEQUENCE IF EXISTS employees_id_seq RESTART WITH 1;" || echo Sequence might not exist  
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -c "ALTER SEQUENCE IF EXISTS users_id_seq RESTART WITH 1;" || echo Sequence might not exist

echo.
echo Verifying cleanup...
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -c "SELECT COUNT(*) as users_count FROM users;"
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -c "SELECT COUNT(*) as employees_count FROM employees;"
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -c "SELECT COUNT(*) as leave_requests_count FROM leave_requests;"

echo.
echo ================================================
echo  CLEANUP COMPLETE - DATABASE IS EMPTY
echo ================================================
echo.
echo Next steps:
echo 1. Restart server: cd d:\Go\src\Task\cmd\employee-service
echo 2. Run: go run main.go
echo 3. Server will auto-seed superadmin account
echo 4. Start fresh in Postman!
echo.

goto end

:error
echo ERROR: Cleanup failed!
exit /b 1

:end
