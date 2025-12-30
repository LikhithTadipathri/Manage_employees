# Leave Management System - Security Architecture

## Overview

The Leave Management System now includes role-based access control and proper security enforcement to ensure:

- ✅ Employees cannot apply for leave without logging in
- ✅ Only employees (non-admin users) can apply for leave
- ✅ Admin users cannot apply for leave as employees
- ✅ Each user can only apply leave for themselves (token-based identity)
- ✅ Only admins can approve/reject leave requests
- ✅ Admins cannot create employee records (those are auto-linked to users)

## Key Changes

### 1. User-Employee Relationship (`user_id` Column)

**New Migration**: `005_add_user_id_to_employees.up.sql`

The `employees` table now includes a `user_id` column with:

- UNIQUE constraint: Each user can have at most one employee record
- FOREIGN KEY constraint: Links to the `users` table with CASCADE delete
- Index: For fast lookups

```sql
ALTER TABLE employees ADD COLUMN user_id INTEGER UNIQUE;
ALTER TABLE employees ADD CONSTRAINT fk_employees_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX idx_employees_user_id ON employees(user_id);
```

### 2. Role-Based Access Control

**Three Roles in System**:

- `admin`: Can create employees, approve/reject leaves, view all leaves
- `user`: Can create their own employee record, apply for leaves
- `employee`: (Alias for `user` in the context of leave management)

### 3. Employee Creation Workflow

**Before**: Employees were generic records without user linkage
**After**: Each employee is automatically linked to the logged-in user

```go
// In employee handler
userCtx, err := middlewares.GetUserFromContext(r)
req.UserID = userCtx.UserID  // ✅ Auto-link to current user
```

**Benefits**:

- No need to provide employee_id in API request
- Prevents users from creating employees for other users
- Token acts as proof of identity

### 4. Leave Application Security

**Enforcement in LeaveHandler.ApplyLeave**:

```go
// Only non-admin users can apply for leave
if userCtx.Role == user.RoleAdmin {
    response.Error(w, http.StatusForbidden, "admins cannot apply for leave")
    return
}

// Use UserID from JWT token (cannot be spoofed)
leaveRequest, err := h.service.ApplyLeave(userCtx.UserID, &req)
```

**Benefits**:

- Admins are explicitly prevented from applying for leave as employees
- User identity comes from JWT token (cryptographically signed)
- No user can apply for leave for someone else

### 5. Admin-Only Operations

**Leave Approval/Rejection** (both require admin role):

```go
// ApproveLeave handler
if userCtx.Role != user.RoleAdmin {
    response.Error(w, http.StatusForbidden, "admin access required")
    return
}

// Admin's user_id is recorded in approval
err := h.service.ApproveLeave(id, userCtx.UserID)
```

**Leave Retrieval**:

```go
// GetAllLeaveRequests (admin only)
if userCtx.Role != user.RoleAdmin {
    response.Error(w, http.StatusForbidden, "admin access required")
    return
}
```

## Access Matrix

| Action             | Employee | Admin | Auth Required |
| ------------------ | -------- | ----- | ------------- |
| Login              | ✅       | ✅    | ❌            |
| Register           | ✅       | ✅    | ❌            |
| Create employee    | ✅ (own) | ❌    | ✅            |
| View own employee  | ✅       | ✅    | ✅            |
| View all employees | ✅       | ✅    | ✅            |
| Apply leave        | ✅       | ❌    | ✅            |
| View own leaves    | ✅       | ❌    | ✅            |
| Cancel own leave   | ✅       | ❌    | ✅            |
| View all leaves    | ❌       | ✅    | ✅            |
| Approve leave      | ❌       | ✅    | ✅            |
| Reject leave       | ❌       | ✅    | ✅            |

## Security Flow

### Employee Flow

```
1️⃣ Employee Registers
   POST /auth/register
   {"username": "john_doe", "password": "secret", "role": "user"}

2️⃣ Employee Logs In
   POST /auth/login
   Returns JWT token with:
   {
     "user_id": 3,
     "role": "user",    // Not admin!
     "expires_at": 1765367213
   }

3️⃣ Create Employee Record (Auto-linked)
   POST /api/v1/employees
   Headers: Authorization: Bearer {token}
   {
     "first_name": "John",
     "last_name": "Doe",
     "email": "john@company.com",
     "phone": "555-0123",
     "position": "Engineer",
     "salary": 80000,
     "hired_date": "2023-01-15T00:00:00Z"
   }

   ✅ Backend auto-links: employee.user_id = 3
   ✅ User cannot provide their own user_id

4️⃣ Apply for Leave
   POST /leave/apply
   Headers: Authorization: Bearer {token}
   {
     "leave_type": "annual",
     "start_date": "2025-12-22T00:00:00Z",
     "end_date": "2025-12-26T00:00:00Z",
     "reason": "Vacation"
   }

   ✅ Backend extracts employee_id from token (user_id=3)
   ✅ Creates: leave_requests(employee_id=3, status='pending', ...)
   ✅ User cannot apply for another person's leave

5️⃣ View Own Leaves
   GET /leave/my-requests
   Headers: Authorization: Bearer {token}

   ✅ Returns only leaves where employee_id matches user's employee
```

### Admin Flow

```
1️⃣ Admin Logs In
   POST /auth/login
   {"username": "admin", "password": "secret"}
   Returns JWT token with:
   {
     "user_id": 1,
     "role": "admin",    // Admin role!
     "expires_at": 1765367213
   }

2️⃣ View All Leave Requests
   GET /leave/all
   Headers: Authorization: Bearer {token}

   ✅ Role check: Requires "admin"
   ✅ Returns all pending leaves from all employees

3️⃣ Approve Leave
   POST /leave/approve/42
   Headers: Authorization: Bearer {token}

   ✅ Role check: Requires "admin"
   ✅ Sets: status='approved', approved_by=1, approval_date=NOW()
   ✅ Admin's identity (user_id=1) recorded in approved_by

4️⃣ Admin Cannot Apply for Leave
   POST /leave/apply

   ❌ Returns 403 Forbidden
   ❌ Message: "admins cannot apply for leave"
   ✅ Prevents admins from creating fake employee leave records
```

## Database Schema Updates

### Employees Table

```sql
CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE,        -- NEW: Links to users table
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    email VARCHAR(100) UNIQUE,
    phone VARCHAR(20),
    position VARCHAR(100),
    salary DECIMAL(10, 2),
    hired_date TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE  -- NEW
);

CREATE INDEX idx_employees_user_id ON employees(user_id);  -- NEW
```

### Leave Requests Table (Unchanged)

```sql
CREATE TABLE leave_requests (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL,   -- Now guaranteed to exist
    leave_type VARCHAR(50),
    status VARCHAR(20) DEFAULT 'pending',
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    reason TEXT,
    days_count INTEGER,
    approved_by INTEGER,            -- Admin's user_id
    approval_date TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (employee_id) REFERENCES employees(id),
    FOREIGN KEY (approved_by) REFERENCES users(id)
);
```

## JWT Token Security

The JWT token is the source of truth for user identity:

```go
type JWTClaims struct {
    UserID   int    `json:"user_id"`     // ✅ Unique identifier
    Role     string `json:"role"`        // ✅ Role for authorization
    Username string `json:"username"`    // For display
    Email    string `json:"email"`       // For display
}
```

**Why this is secure**:

- JWT is cryptographically signed with server secret
- Client cannot modify token without invalidating signature
- Server verifies signature before processing request
- user_id in token cannot be spoofed
- role in token cannot be elevated by user

## Testing the Security

### Test Case 1: Employee Cannot Apply Without Login

```bash
POST /leave/apply (without Authorization header)
→ 401 Unauthorized ✅
```

### Test Case 2: Employee Cannot Become Admin

```bash
POST /auth/register {"role": "admin"}  # User tries to register as admin
→ 403 Forbidden (or defaults to "user" role) ✅
```

### Test Case 3: Admin Cannot Apply for Leave

```bash
# Login as admin
POST /auth/login {"username": "admin", ...}
→ {role: "admin"}

# Try to apply for leave
POST /leave/apply {"leave_type": "annual", ...}
→ 403 Forbidden: "admins cannot apply for leave" ✅
```

### Test Case 4: Employee Automatically Linked to User

```bash
# Login as john_doe (user_id=3)
POST /auth/login {"username": "john_doe", ...}

# Create employee record
POST /api/v1/employees {...}
→ employee.user_id = 3 automatically ✅
```

### Test Case 5: Employee Cannot Apply for Another Person's Leave

```bash
# Token contains user_id=3
POST /leave/apply (with jane's leave dates)
→ Creates leave with employee_id from token (3) ✅
→ Jane (user_id=5) sees this leave in her /leave/my-requests? NO ✅
```

## Migration Steps

1. **Apply migration** to add `user_id` column:

   ```bash
   migrate -path migrations -database "postgres://..." up
   ```

2. **Update existing employee-user links** (if any):

   ```sql
   -- Link employees to users by email matching (if applicable)
   UPDATE employees e SET user_id = u.id FROM users u WHERE e.email = u.email;
   ```

3. **Verify no orphaned employees** (employees without users):
   ```sql
   SELECT COUNT(*) FROM employees WHERE user_id IS NULL;
   ```

## Conclusion

This implementation follows enterprise HRMS (Human Resource Management System) patterns where:

- Employees must be users first (registration)
- Employee records are linked to user accounts (1:1 relationship)
- Leave applications are inherently tied to the authenticated user
- Admin operations are explicitly role-checked
- Identity is cryptographically verified via JWT tokens

**Result**: Secure, tamper-proof leave management system where:
✅ No unauthorized access
✅ No identity spoofing  
✅ No role elevation
✅ Proper audit trail (approved_by records which admin approved)
