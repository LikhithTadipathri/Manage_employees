# Implementation Summary - Secure Leave Management

## Changes Made

### 1. Database Schema Updates

#### New Migration Files Created:

- `migrations/005_add_user_id_to_employees.up.sql` - Adds user_id column with constraints
- `migrations/005_add_user_id_to_employees.down.sql` - Rollback migration

#### Modified Files:

- `utils/helpers/db.go` - Updated initial schema for both SQLite and PostgreSQL to include user_id column in employees table

### 2. Models Updated

#### `models/employee/employee.go`

- Added `UserID int` field to Employee struct
- Added `UserID int` field to CreateEmployeeRequest struct
- Added optional `HiredDate *time.Time` field to CreateEmployeeRequest

### 3. HTTP Handlers Updated

#### `http/handlers/employee_handler.go` - CreateEmployee

- **SECURITY FIX**: Auto-links employee to current authenticated user
- Captures `userCtx.UserID` and sets it in request
- Prevents users from creating employees for other users
- Updated error message for duplicate constraint

#### `http/handlers/leave_handler.go` - ApplyLeave

- **SECURITY FIX**: Added admin role check
- Rejects applications from users with `role=admin`
- Prevents admins from creating fake employee leave records
- Better error handling for missing employee records

### 4. Repositories Updated

#### `repositories/postgres/employee_repo.go` - CreateEmployee

- Updated INSERT query to include user_id column
- Updated SELECT queries (GetEmployeeByID, GetAllEmployees) to include user_id
- Both SQLite and PostgreSQL paths updated

### 5. Services Updated

#### `services/employee/employee_service.go` - CreateEmployee

- Updated to accept and use UserID from request
- Handles optional hired_date with fallback to current time
- Passes UserID to repository

### 6. Middleware & Security

#### `http/middlewares/` (No changes needed)

- Existing JWTMiddleware already extracts user_id and role from token
- Token is cryptographically signed and cannot be spoofed

## Security Improvements Summary

### Before:

❌ Employees could be created without user linkage
❌ Leave applications used user_id directly as employee_id (no validation)
❌ No role-based check for apply leave endpoint
❌ Users could potentially create employees for anyone with proper SQL knowledge

### After:

✅ Each employee is linked to exactly one user (UNIQUE constraint)
✅ Employees are auto-linked when created (backend enforces, not user input)
✅ Admin users cannot apply for leave (explicit role check)
✅ Leave applications use authenticated user's identity from JWT token
✅ Employee creation requires valid user_id in database
✅ Proper foreign key constraints prevent orphaned records

## API Behavior Changes

### Employee Creation

**Before**:

```bash
POST /api/v1/employees
{
  "first_name": "John",
  ...
}
```

- Created employee with no user linkage
- Could be created by multiple users for same person

**After**:

```bash
POST /api/v1/employees (with valid JWT token)
{
  "first_name": "John",
  ...
}
```

- Employee automatically linked to authenticated user
- Unique constraint prevents user from creating multiple employee records
- Only one user_id can create one employee record

### Leave Application

**Before**:

```bash
POST /leave/apply (admin or user)
{
  "leave_type": "annual",
  ...
}
```

- Admin could apply for leave (unintended)
- Left as user records in database

**After**:

```bash
POST /leave/apply (user only, not admin)
{
  "leave_type": "annual",
  ...
}
```

- 403 Forbidden if user role is "admin"
- Employee_id extracted from JWT token (user_id)
- Prevents admins from creating false leave records

## Database Constraints

### New Unique Constraint

```sql
ALTER TABLE employees ADD CONSTRAINT unique_user_id_employees UNIQUE(user_id);
```

**Effect**: Each user can have at most one employee record

### New Foreign Key

```sql
ALTER TABLE employees ADD CONSTRAINT fk_employees_user_id
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
```

**Effect**:

- Ensures employee_user_id exists in users table
- Cascading delete: if user deleted, employee record also deleted
- Prevents orphaned employee records

### New Index

```sql
CREATE INDEX idx_employees_user_id ON employees(user_id);
```

**Effect**: Fast lookups when finding employees by user_id

## Migration Path

### For Development:

1. Fresh database: Schema includes user_id by default
2. Existing database: Run migration 005 to add user_id column

### For Production:

1. Backup database
2. Run migration: `migrate up 005_add_user_id_to_employees`
3. Verify: `SELECT COUNT(*) FROM employees WHERE user_id IS NULL;` (should return 0)
4. Deploy code changes

## Testing Checklist

- [ ] Employee cannot apply for leave without login (401 error)
- [ ] Employee can apply for leave when authenticated
- [ ] Admin gets 403 when trying to apply for leave
- [ ] Admin cannot see/modify other admin's leave records
- [ ] Employee can only see/cancel own leave
- [ ] Admin can view all leave requests
- [ ] Admin can approve/reject leave with proper recorded approval_by
- [ ] User cannot create multiple employee records (unique constraint error)
- [ ] User cannot modify another user's employee record
- [ ] Leave request is linked to correct employee via user_id
- [ ] Database cascades correctly on user deletion

## Files Modified Summary

| File                                               | Changes                     | Type           |
| -------------------------------------------------- | --------------------------- | -------------- |
| `models/employee/employee.go`                      | Added UserID fields         | Model          |
| `http/handlers/employee_handler.go`                | Auto-link user to employee  | Security       |
| `http/handlers/leave_handler.go`                   | Added admin role check      | Security       |
| `repositories/postgres/employee_repo.go`           | Updated queries for user_id | Data Access    |
| `services/employee/employee_service.go`            | Handle UserID in creation   | Business Logic |
| `utils/helpers/db.go`                              | Schema update with user_id  | Database       |
| `migrations/005_add_user_id_to_employees.up.sql`   | New migration               | Migration      |
| `migrations/005_add_user_id_to_employees.down.sql` | Rollback migration          | Migration      |
| `docs/SECURITY_ARCHITECTURE.md`                    | Documentation               | Documentation  |

## Performance Impact

**Positive**:

- New index on user_id enables fast lookups
- Unique constraint prevents duplicate employee records

**Neutral**:

- One additional foreign key check (microseconds)
- One additional column in employees table

## Backward Compatibility

**Breaking Changes**:

- Employee creation now auto-links to user (frontend must ensure user is logged in)
- Admin users cannot apply for leave (expected behavior, not a bug)

**Non-Breaking**:

- All existing READ operations still work (added user_id column, no removal)
- Employee API endpoints unchanged (except behavior)
- Leave API endpoints unchanged (except admin rejection)

## Rollback Instructions

If needed to rollback:

```bash
migrate down 005_add_user_id_to_employees
# OR manually:
# ALTER TABLE employees DROP CONSTRAINT fk_employees_user_id;
# ALTER TABLE employees DROP CONSTRAINT unique_user_id_employees;
# DROP INDEX idx_employees_user_id;
# ALTER TABLE employees DROP COLUMN user_id;
```

## Recommended Next Steps

1. **Update API documentation** with new security model
2. **Add integration tests** for role-based access control
3. **Implement audit logging** for employee creation and leave operations
4. **Add employee verification workflow** (HR approval before employee can apply for leave)
5. **Implement leave balance tracking** (track used vs available days)
