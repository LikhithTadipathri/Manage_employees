# Deployment & Migration Guide

## Current Status

✅ **Security Implementation**: Complete  
✅ **Code Changes**: Complete (7 files modified)  
✅ **Database Migration**: Ready (migration 005 created)  
✅ **Testing**: Partial (auth, RBAC verified; full workflow pending migration)  
⏳ **Production Ready**: Pending migration application to database

---

## Step-by-Step Deployment Guide

### Step 1: Database Migration (REQUIRED)

The system has been updated to add a `user_id` column to the employees table. This is required for the full security model to work.

#### For PostgreSQL

```bash
# Option A: Using migrate CLI (recommended)
cd D:\Go\src\Task
migrate -path migrations -database "postgres://username:password@localhost:5432/employee_db?sslmode=disable" up

# Option B: Manual SQL execution
psql -U postgres -d employee_db -f migrations/005_add_user_id_to_employees.up.sql
```

**What This Does**:

- Adds `user_id` INTEGER column to employees table
- Creates FOREIGN KEY constraint to users table with CASCADE delete
- Creates UNIQUE constraint (prevents multiple employees per user)
- Creates INDEX for performance

#### For SQLite (Local Development)

Migration is automatically applied when you start the application with a fresh database. For existing SQLite databases, the migration will be skipped with a message:

```
Skipping user_id index - migration 005 not yet applied
```

To manually apply migration 005 to SQLite:

```sql
ALTER TABLE employees ADD COLUMN user_id INTEGER UNIQUE;
ALTER TABLE employees ADD FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX idx_employees_user_id ON employees(user_id);
```

### Step 2: Link Existing Employee Records (OPTIONAL)

If you have existing employee records in your database, you can link them to user accounts:

```sql
-- Link employees to users by matching email addresses
UPDATE employees e
SET user_id = u.id
FROM users u
WHERE e.email = u.email;

-- Verify all employees are linked
SELECT COUNT(*) as unlinked_employees FROM employees WHERE user_id IS NULL;
-- Should return: 0
```

### Step 3: Verify Migration Success

```sql
-- Check employees table structure
\d employees;
-- Should show: user_id column with UNIQUE constraint and FK

-- Check data integrity
SELECT COUNT(*) FROM employees WHERE user_id IS NULL;
-- Should return: 0 (all employees linked)

-- Verify uniqueness constraint
SELECT COUNT(*) as employee_count, COUNT(DISTINCT user_id) as unique_users
FROM employees;
-- Both numbers should be equal
```

### Step 4: Restart Application

```bash
cd D:\Go\src\Task

# Build the latest code
go build -o employee-service.exe .\cmd\employee-service\main.go

# Run the application
.\employee-service.exe

# Or in Docker
docker-compose -f docker/docker-compose.yml up -d
```

### Step 5: Run Comprehensive Tests

```bash
# PowerShell
cd D:\Go\src\Task
& .\test_api.ps1

# Or manually test
curl.exe -X GET http://localhost:8080/health
```

---

## What Changed

### Database Schema

**Before Migration 005**:

```
employees table:
- id
- first_name
- last_name
- email
- phone
- position
- salary
- hired_date
- created_at
- updated_at
```

**After Migration 005**:

```
employees table:
- id
- user_id (NEW - UNIQUE - FK to users)  ← This is the critical change
- first_name
- last_name
- email
- phone
- position
- salary
- hired_date
- created_at
- updated_at
```

### Code Changes Summary

| File                                     | Changes                         | Impact                                  |
| ---------------------------------------- | ------------------------------- | --------------------------------------- |
| `models/employee/employee.go`            | Added UserID field              | Employee model now tracks user linkage  |
| `http/handlers/employee_handler.go`      | Auto-link to authenticated user | Employees created are linked to creator |
| `http/handlers/leave_handler.go`         | Added admin role check          | Admins cannot apply for leave (403)     |
| `repositories/postgres/employee_repo.go` | Updated CREATE/SELECT queries   | User_id persisted and retrieved         |
| `services/employee/employee_service.go`  | Updated to handle UserID        | Service layer propagates user_id        |
| `utils/helpers/db.go`                    | Updated schema initialization   | New databases include user_id column    |
| `docker/docker-compose.yml`              | Fixed network configuration     | PostgreSQL and service on same network  |

### Migration Files

| File                                               | Purpose                                         |
| -------------------------------------------------- | ----------------------------------------------- |
| `migrations/005_add_user_id_to_employees.up.sql`   | Apply migration (add column, constraints)       |
| `migrations/005_add_user_id_to_employees.down.sql` | Rollback migration (remove column, constraints) |

---

## Security Features Enabled After Migration

### 1. Auto-Linking of Employees to Users

When a user creates an employee record:

```json
POST /api/v1/employees
Authorization: Bearer {user_jwt}
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  ...
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 3,  ← Auto-linked from JWT token
    "first_name": "John",
    ...
  }
}
```

### 2. One Employee Per User (Enforced)

```sql
-- This constraint prevents duplicates:
UNIQUE(user_id)

-- Result: Only one employee record per user
-- Trying to create a second employee will fail with:
-- unique constraint violation on column "user_id"
```

### 3. Admin Prevention from Applying Leave

```
POST /leave/apply
Authorization: Bearer {admin_jwt}

Response: 403 Forbidden
{
  "success": false,
  "message": "admins cannot apply for leave"
}
```

### 4. Cascade Delete Protection

```sql
-- If a user is deleted, their employee record is automatically deleted:
DELETE FROM users WHERE id = 3;
-- Result: Employee record with user_id = 3 is also deleted
-- All their leave requests are deleted too (FK cascade)
```

---

## Testing After Deployment

### Test 1: Basic Health Check

```bash
curl http://localhost:8080/health
# Expected: 200 OK with {"status":"healthy"}
```

### Test 2: User Authentication

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","password":"john123"}'
# Expected: 200 OK with JWT token
```

### Test 3: Employee Creation with Auto-Linking

```bash
curl -X POST http://localhost:8080/api/v1/employees \
  -H "Authorization: Bearer {user_jwt}" \
  -H "Content-Type: application/json" \
  -d '{"first_name":"John",...}'
# Expected: 201 Created with user_id auto-populated
```

### Test 4: Admin Leave Prevention (CRITICAL)

```bash
curl -X POST http://localhost:8080/leave/apply \
  -H "Authorization: Bearer {admin_jwt}" \
  -H "Content-Type: application/json" \
  -d '{"leave_type":"annual",...}'
# Expected: 403 Forbidden with "admins cannot apply for leave"
```

### Test 5: User Leave Application

```bash
curl -X POST http://localhost:8080/leave/apply \
  -H "Authorization: Bearer {user_jwt}" \
  -H "Content-Type: application/json" \
  -d '{"leave_type":"annual",...}'
# Expected: 201 Created with leave request details
```

---

## Troubleshooting

### Error: "column 'user_id' does not exist"

**Cause**: Migration 005 hasn't been applied yet

**Fix**:

```bash
# Apply migration
migrate -path migrations -database "postgres://..." up

# Or restart with Docker to auto-initialize
docker-compose -f docker/docker-compose.yml down
docker-compose -f docker/docker-compose.yml up -d
```

### Error: "UNIQUE constraint violation on user_id"

**Cause**: User already has an employee record

**Fix**: Check existing records

```sql
SELECT id, user_id FROM employees WHERE user_id = {user_id};
-- Delete the old record if needed:
DELETE FROM employees WHERE user_id = {user_id} AND id < {new_id};
```

### Error: "Admin cannot apply for leave" when should succeed

**Cause**: User account has admin role instead of user role

**Fix**: Check user role

```sql
SELECT id, username, role FROM users WHERE username = 'john_doe';
-- Should show role = 'user', not 'admin'
```

---

## Rollback Procedure (If Needed)

If you need to rollback migration 005:

```bash
# Using migrate CLI
migrate -path migrations -database "postgres://..." down

# Or manually execute rollback SQL
psql -U postgres -d employee_db -f migrations/005_add_user_id_to_employees.down.sql
```

**Note**: This will remove the user_id column and all constraints. The application will continue to work but security features will be reduced.

---

## Performance Considerations

### Index on user_id

Migration 005 creates an index:

```sql
CREATE INDEX idx_employees_user_id ON employees(user_id);
```

This ensures fast lookups when:

- Creating employees (check uniqueness)
- Getting employee by user (leave handler)
- Linking users to employees

### Migration Downtime

Expected downtime: < 1 second for most databases

- Adding column: 0.01ms
- Adding constraint: 0.1ms
- Creating index: 0-100ms (depends on table size)

For large production databases (millions of rows), consider:

1. Running migration during off-hours
2. Using `CREATE INDEX CONCURRENTLY` for zero-downtime indexing

---

## Verification Checklist

After deployment, verify:

- [ ] Migration 005 applied successfully
- [ ] `SELECT user_id FROM employees LIMIT 1;` returns a value
- [ ] `SELECT COUNT(*) FROM employees WHERE user_id IS NULL;` returns 0
- [ ] Health endpoint returns 200 OK
- [ ] Admin can login and get JWT with role=admin
- [ ] User can login and get JWT with role=user
- [ ] Admin trying to apply for leave gets 403
- [ ] User can create employee and get auto-linked user_id
- [ ] User can apply for leave successfully
- [ ] Admin can view all leave requests
- [ ] All tests pass

---

## Documentation References

For more details, see:

- **Security Architecture**: `docs/SECURITY_ARCHITECTURE.md`
- **Implementation Changes**: `docs/IMPLEMENTATION_CHANGES.md`
- **API Documentation**: `docs/API_SECURITY_DOCUMENTATION.md`
- **Security Test Report**: `docs/SECURITY_TEST_REPORT.md`

---

## Support

If you encounter any issues during deployment:

1. Check the server logs:

   ```bash
   docker logs employee_service  # For Docker
   # or check stdout for local runs
   ```

2. Verify database connectivity:

   ```bash
   psql -U postgres -d employee_db -c "SELECT 1;"
   ```

3. Check migration status:

   ```bash
   migrate -path migrations -database "postgres://..." version
   ```

4. Review the migration files:
   - `migrations/005_add_user_id_to_employees.up.sql`
   - `migrations/005_add_user_id_to_employees.down.sql`
