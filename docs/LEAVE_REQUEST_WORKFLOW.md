# How Employees Can Request Leave - Complete Workflow

## Current Status

✅ **Leave Request System is Ready** - The API endpoints and security are implemented
⏳ **Blocked by Migration** - Migration 005 must be applied to enable leave requests
❌ **No Employee Records** - Employee records need to be created first

---

## Prerequisites for Leave Requests

### 1. **Migration 005 Must Be Applied**

The `user_id` column must be added to the employees table first:

```powershell
migrate -path migrations -database "postgres://user:password@localhost:5432/employee_db?sslmode=disable" up
```

Expected output:

```
2025/12/10 17:45:00 [postgresql] Reading migration files from migrations
2025/12/10 17:45:00 [postgresql] Migrating up to 005_add_user_id_to_employees
2025/12/10 17:45:00 [postgresql] [OK] Migration 005_add_user_id_to_employees completed
```

### 2. **Employee Record Must Be Created**

Before an employee can apply for leave, they need an employee record linked to their user account.

**What happens currently:**

- Users are seeded: `admin`, `john_doe`, `jane_smith`
- No employee records exist in the employees table
- When john_doe tries to apply for leave: Error "employee record not found"

---

## Step-by-Step: How to Request Leave

### Step 1: Create Employee Record (First Time Only)

**Endpoint:** `POST /employees`

**Headers:**

```
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json
```

**Request Body:**

```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@company.com",
  "position": "Software Engineer",
  "department": "Engineering",
  "hired_date": "2023-01-15"
}
```

**What Happens:**

- ✅ User must be authenticated (JWT token required)
- ✅ Automatically links employee to the authenticated user
- ✅ Sets `user_id` = current user's ID (cannot be spoofed)
- ✅ Admin users can create employee records for themselves
- ✅ Regular users can only create records for themselves

**Success Response (201 Created):**

```json
{
  "success": true,
  "message": "Employee created successfully",
  "data": {
    "id": 2,
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@company.com",
    "position": "Software Engineer",
    "department": "Engineering",
    "user_id": 2,
    "hired_date": "2023-01-15T00:00:00Z",
    "created_at": "2025-12-10T18:00:00Z"
  }
}
```

---

### Step 2: Apply for Leave

**Endpoint:** `POST /leave/apply`

**Headers:**

```
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json
```

**Request Body:**

```json
{
  "leave_type": "annual",
  "start_date": "2025-12-15T00:00:00Z",
  "end_date": "2025-12-17T00:00:00Z",
  "reason": "Vacation time with family"
}
```

**Valid Leave Types:**

- `annual` - Annual/vacation leave
- `sick` - Sick leave
- `casual` - Casual leave
- `maternity` - Maternity leave
- `paternity` - Paternity leave
- `unpaid` - Unpaid leave

**Date Format Requirements:**

- Must be RFC3339 format with timezone: `2025-12-15T00:00:00Z`
- Alternative: ISO 8601 standard datetime format
- Start date must be before or equal to end date
- Must include at least 1 working day (Monday-Friday)

**Validations Performed:**

- ✅ `leave_type` - Required, must be valid type
- ✅ `start_date` - Required, must be valid date
- ✅ `end_date` - Required, must be valid date
- ✅ `reason` - Required, max 500 characters
- ✅ Date range - At least 1 working day required
- ✅ Admin prevention - Admins **CANNOT** apply for leave (403 Forbidden)

**Success Response (201 Created):**

```json
{
  "success": true,
  "message": "Leave request submitted successfully",
  "data": {
    "id": 5,
    "employee_id": 2,
    "leave_type": "annual",
    "status": "pending",
    "start_date": "2025-12-15T00:00:00Z",
    "end_date": "2025-12-17T00:00:00Z",
    "reason": "Vacation time with family",
    "days_count": 2,
    "created_at": "2025-12-10T18:05:00Z",
    "updated_at": "2025-12-10T18:05:00Z"
  }
}
```

**Error Responses:**

❌ `400 Bad Request` - Invalid JSON format

```json
{ "success": false, "message": "Invalid request body" }
```

❌ `400 Bad Request` - Validation failed

```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "leave_type": "leave type is required",
    "start_date": "start date is required"
  }
}
```

❌ `403 Forbidden` - Admin cannot apply for leave

```json
{ "success": false, "message": "admins cannot apply for leave" }
```

❌ `404 Not Found` - No employee record found

```json
{
  "success": false,
  "message": "employee record not found. Please contact HR to create your employee profile."
}
```

❌ `401 Unauthorized` - Invalid or missing JWT token

```json
{ "success": false, "message": "unauthorized" }
```

---

## Step 3: View My Leave Requests

**Endpoint:** `GET /leave/my-requests`

**Headers:**

```
Authorization: Bearer {JWT_TOKEN}
```

**Optional Query Parameters:**

- `status=pending` - Filter by status (pending, approved, rejected, cancelled)

**Success Response (200 OK):**

```json
{
  "success": true,
  "message": "Leave requests retrieved successfully",
  "data": [
    {
      "id": 5,
      "employee_id": 2,
      "leave_type": "annual",
      "status": "pending",
      "start_date": "2025-12-15T00:00:00Z",
      "end_date": "2025-12-17T00:00:00Z",
      "reason": "Vacation time with family",
      "days_count": 2,
      "created_at": "2025-12-10T18:05:00Z",
      "updated_at": "2025-12-10T18:05:00Z"
    }
  ]
}
```

---

## Step 4: Admin Approves/Rejects Leave

**Endpoint:** `POST /leave/{id}/approve`

**Headers:**

```
Authorization: Bearer {ADMIN_JWT_TOKEN}
Content-Type: application/json
```

**Request Body:**

```json
{
  "approved": true,
  "remarks": "Approved for end-of-year vacation"
}
```

**Success Response (200 OK):**

```json
{
  "success": true,
  "message": "Leave request updated successfully",
  "data": {
    "id": 5,
    "status": "approved",
    "approved_by": 1,
    "approval_date": "2025-12-10T18:10:00Z"
  }
}
```

---

## Security Features

### ✅ Admin Prevention

- **Admins CANNOT apply for leave** (returns 403 Forbidden)
- Only regular employees (non-admin users) can apply
- This prevents management from abusing the system

### ✅ User-Employee Linkage

- When creating an employee record, `user_id` is automatically set
- **Cannot be manually overridden** - enforced at handler level
- Prevents users from linking to other users' employee records

### ✅ JWT Identity Verification

- `user_id` extracted from JWT token (cryptographically verified)
- Cannot be spoofed by modifying request body
- Token expiration: 1 hour

### ✅ Database Constraints

- UNIQUE constraint on `user_id` in employees table
- FOREIGN KEY constraint ensures data integrity
- CASCADE DELETE: If user is deleted, employee record is deleted

---

## Complete Example: Employee Leave Request Workflow

```powershell
# 1. LOGIN
$login = curl.exe -s -X POST http://localhost:8080/auth/login `
  -H "Content-Type: application/json" `
  -d '{"username":"john_doe","password":"john123"}'
$token = ($login | ConvertFrom-Json).data.token

# 2. CREATE EMPLOYEE RECORD (First time only)
curl.exe -s -X POST http://localhost:8080/employees `
  -H "Authorization: Bearer $token" `
  -H "Content-Type: application/json" `
  -d '{
    "first_name":"John",
    "last_name":"Doe",
    "email":"john@company.com",
    "position":"Engineer",
    "department":"IT",
    "hired_date":"2023-01-15"
  }'

# 3. APPLY FOR LEAVE
curl.exe -s -X POST http://localhost:8080/leave/apply `
  -H "Authorization: Bearer $token" `
  -H "Content-Type: application/json" `
  -d '{
    "leave_type":"annual",
    "start_date":"2025-12-15T00:00:00Z",
    "end_date":"2025-12-17T00:00:00Z",
    "reason":"Vacation"
  }'

# 4. VIEW MY LEAVE REQUESTS
curl.exe -s -X GET "http://localhost:8080/leave/my-requests" `
  -H "Authorization: Bearer $token"

# 5. ADMIN APPROVES (using admin token)
curl.exe -s -X POST http://localhost:8080/leave/5/approve `
  -H "Authorization: Bearer $adminToken" `
  -H "Content-Type: application/json" `
  -d '{"approved":true,"remarks":"Approved"}'
```

---

## Current Blockers

| Step            | Status | Blocker                   | Solution                         |
| --------------- | ------ | ------------------------- | -------------------------------- |
| Create employee | ⏳     | Migration 005 not applied | Run: `migrate up`                |
| Apply for leave | ⏳     | No employee records       | Create via `/employees` endpoint |
| View requests   | ✅     | None                      | Works after employee created     |
| Admin approval  | ✅     | None                      | Works with admin token           |

---

## Next Steps to Enable Full Workflow

1. **Apply Migration 005**

   ```powershell
   migrate -path migrations -database "postgres://..." up
   ```

2. **Restart Application**

   ```powershell
   Stop-Process -Name employee-service
   .\employee-service.exe
   ```

3. **Create Employee Records**

   ```powershell
   POST /employees with authenticated user token
   ```

4. **Apply for Leave**

   ```powershell
   POST /leave/apply with leave details
   ```

5. **View and Manage Requests**
   - Employees: View their own requests
   - Admin: View all, approve/reject
