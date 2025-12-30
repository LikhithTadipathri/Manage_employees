# Updated API Documentation - Secure Leave Management

## Authentication & Authorization

### JWT Token Structure

```json
{
  "user_id": 3,
  "username": "john_doe",
  "email": "john@company.com",
  "role": "user",
  "exp": 1765367213
}
```

**Roles**:

- `admin` - Can manage leave requests, view all leaves
- `user` - Can apply for own leave, view own leaves

---

## Employee Management APIs

### Create Employee (Auto-Linked to User)

```http
POST /api/v1/employees
Authorization: Bearer {token}
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@company.com",
  "phone": "555-0123",
  "position": "Software Engineer",
  "salary": 85000,
  "hired_date": "2023-01-15T00:00:00Z"
}
```

**✅ Security Features**:

- Requires valid JWT token
- Auto-links to authenticated user (`user_id` from token)
- User cannot override user_id
- Unique constraint prevents multiple employee records per user
- Response includes auto-assigned `user_id`

**Response**:

```json
{
  "success": true,
  "message": "Employee created successfully",
  "data": {
    "id": 1,
    "user_id": 3,
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@company.com",
    "position": "Software Engineer",
    "salary": 85000,
    "hired_date": "2023-01-15T00:00:00Z",
    "created_at": "2025-12-10T09:10:41Z",
    "updated_at": "2025-12-10T09:10:41Z"
  }
}
```

**Error Cases**:

```json
// 401: No valid token
{"success": false, "message": "unauthorized"}

// 409: User already has employee record
{"success": false, "message": "Email address already exists or you already have an employee record"}

// 400: Validation error
{"success": false, "message": "Validation failed", "fields": {...}}
```

---

## Leave Management APIs

### 1. Apply for Leave (Employees Only)

```http
POST /leave/apply
Authorization: Bearer {employee_token}
Content-Type: application/json

{
  "leave_type": "annual",
  "start_date": "2025-12-22T00:00:00Z",
  "end_date": "2025-12-26T00:00:00Z",
  "reason": "Year-end vacation"
}
```

**✅ Security Features**:

- Requires `user` role (rejects `admin` role)
- Employee_id extracted from JWT token (cannot be spoofed)
- Validates leave dates
- Calculates working days (Mon-Fri only)

**Valid Leave Types**:

- `annual` - Annual vacation
- `sick` - Sick leave
- `casual` - Casual leave
- `maternity` - Maternity leave
- `paternity` - Paternity leave
- `unpaid` - Unpaid leave

**Response** (201 Created):

```json
{
  "success": true,
  "message": "Leave request submitted successfully",
  "data": {
    "id": 1,
    "employee_id": 1,
    "leave_type": "annual",
    "status": "pending",
    "start_date": "2025-12-22T00:00:00Z",
    "end_date": "2025-12-26T00:00:00Z",
    "reason": "Year-end vacation",
    "days_count": 5,
    "created_at": "2025-12-10T09:15:00Z",
    "updated_at": "2025-12-10T09:15:00Z"
  }
}
```

**Error Cases**:

```json
// 401: Not authenticated
{"success": false, "message": "unauthorized"}

// 403: Admin trying to apply
{"success": false, "message": "admins cannot apply for leave"}

// 404: Employee record not found
{"success": false, "message": "employee record not found. Please contact HR to create your employee profile."}

// 400: Validation errors
{
  "success": false,
  "message": "Validation failed",
  "fields": {
    "leave_type": "leave type is required",
    "dates": "leave period must include at least one working day"
  }
}
```

---

### 2. Get My Leave Requests (Employees Only)

```http
GET /leave/my-requests?status=pending
Authorization: Bearer {employee_token}
```

**Query Parameters**:

- `status` (optional) - Filter by status: `pending`, `approved`, `rejected`, `cancelled`

**Response**:

```json
{
  "success": true,
  "message": "Leave requests retrieved successfully",
  "data": {
    "count": 2,
    "leave_requests": [
      {
        "id": 1,
        "employee_id": 1,
        "leave_type": "annual",
        "status": "pending",
        "start_date": "2025-12-22T00:00:00Z",
        "end_date": "2025-12-26T00:00:00Z",
        "reason": "Year-end vacation",
        "days_count": 5,
        "created_at": "2025-12-10T09:15:00Z",
        "updated_at": "2025-12-10T09:15:00Z"
      },
      {
        "id": 2,
        "employee_id": 1,
        "leave_type": "sick",
        "status": "approved",
        "start_date": "2025-12-18T00:00:00Z",
        "end_date": "2025-12-19T00:00:00Z",
        "reason": "Medical appointment",
        "days_count": 2,
        "approved_by": 5,
        "approval_date": "2025-12-11T10:30:00Z",
        "created_at": "2025-12-10T09:20:00Z",
        "updated_at": "2025-12-11T10:30:00Z"
      }
    ]
  }
}
```

**Security**:

- Only returns leaves for authenticated user's employee record
- Cannot view other employees' leaves

---

### 3. Cancel Leave Request (Employees Only)

```http
DELETE /leave/cancel/{id}
Authorization: Bearer {employee_token}
```

**Path Parameters**:

- `id` - Leave request ID to cancel

**Response** (200 OK):

```json
{
  "success": true,
  "message": "Leave request cancelled successfully"
}
```

**Error Cases**:

```json
// 401: Not authenticated
{"success": false, "message": "unauthorized"}

// 403: Trying to cancel someone else's leave
{"success": false, "message": "you don't have permission to cancel this leave request"}

// 404: Leave request not found
{"success": false, "message": "Leave request not found"}

// 400: Cannot cancel non-pending leave
{"success": false, "message": "Validation failed", "fields": {"status": "Only pending leaves can be cancelled"}}
```

**Security**:

- Only employee who created the leave can cancel it
- Cannot cancel already approved/rejected leaves
- Identity verified via JWT token

---

### 4. Get All Leave Requests (Admin Only)

```http
GET /leave/all?status=pending
Authorization: Bearer {admin_token}
```

**Query Parameters**:

- `status` (optional) - Filter: `pending`, `approved`, `rejected`, `cancelled`

**Response**:

```json
{
  "success": true,
  "message": "Leave requests retrieved successfully",
  "data": {
    "count": 5,
    "leave_requests": [
      {
        "id": 1,
        "employee_id": 1,
        "leave_type": "annual",
        "status": "pending",
        "start_date": "2025-12-22T00:00:00Z",
        "end_date": "2025-12-26T00:00:00Z",
        "reason": "Year-end vacation",
        "days_count": 5,
        "created_at": "2025-12-10T09:15:00Z",
        "updated_at": "2025-12-10T09:15:00Z"
      },
      {
        "id": 2,
        "employee_id": 2,
        "leave_type": "sick",
        "status": "pending",
        "start_date": "2025-12-18T00:00:00Z",
        "end_date": "2025-12-19T00:00:00Z",
        "reason": "Medical appointment",
        "days_count": 2,
        "created_at": "2025-12-10T09:20:00Z",
        "updated_at": "2025-12-10T09:20:00Z"
      }
    ]
  }
}
```

**Error Cases**:

```json
// 401: Not authenticated
{"success": false, "message": "unauthorized"}

// 403: Not an admin
{"success": false, "message": "admin access required"}
```

**Security**:

- Requires `admin` role
- Non-admin users get 403 Forbidden
- Returns all leaves for HR/admin review

---

### 5. Approve Leave Request (Admin Only)

```http
POST /leave/approve/{id}
Authorization: Bearer {admin_token}
Content-Type: application/json

{}
```

**Path Parameters**:

- `id` - Leave request ID to approve

**Response** (200 OK):

```json
{
  "success": true,
  "message": "Leave request approved successfully"
}
```

**Updated Leave Record**:

```
status → 'approved'
approved_by → {admin_user_id}
approval_date → NOW()
updated_at → NOW()
```

**Error Cases**:

```json
// 401: Not authenticated
{"success": false, "message": "unauthorized"}

// 403: Not an admin
{"success": false, "message": "admin access required"}

// 404: Leave not found
{"success": false, "message": "Leave request not found"}

// 400: Cannot approve non-pending leave
{"success": false, "message": "Validation failed", "fields": {"status": "Only pending leaves can be approved"}}
```

**Security**:

- Requires `admin` role
- Admin's user_id recorded in `approved_by`
- Provides audit trail of who approved the leave

---

### 6. Reject Leave Request (Admin Only)

```http
POST /leave/reject/{id}
Authorization: Bearer {admin_token}
Content-Type: application/json

{}
```

**Path Parameters**:

- `id` - Leave request ID to reject

**Response** (200 OK):

```json
{
  "success": true,
  "message": "Leave request rejected successfully"
}
```

**Updated Leave Record**:

```
status → 'rejected'
approved_by → {admin_user_id}
approval_date → NOW()
updated_at → NOW()
```

**Error Cases**: Same as Approve

**Security**:

- Requires `admin` role
- Admin's user_id recorded in `approved_by`
- Clear audit trail

---

## Leave Status Workflow

```
┌─────────────┐
│   Pending   │  ← User applies for leave
└──────┬──────┘
       │
       ├─→ [Admin Approves] → Approved (approved_by set, approval_date set)
       │
       ├─→ [Admin Rejects] → Rejected (approved_by set, approval_date set)
       │
       └─→ [User Cancels] → Cancelled (while pending only)
```

---

## Comparative Examples

### ❌ BEFORE (Insecure)

```bash
# Employee could apply without login
curl -X POST http://localhost:8080/leave/apply

# Admin could apply for leave (unintended)
curl -X POST http://localhost:8080/leave/apply \
  -H "Authorization: Bearer {admin_token}"

# No validation of employee_id source
# Employee IDs could be spoofed
```

### ✅ AFTER (Secure)

```bash
# Must authenticate first
curl -X POST http://localhost:8080/auth/login
  → Returns token with user_id and role

# Admin gets explicit rejection
curl -X POST http://localhost:8080/leave/apply \
  -H "Authorization: Bearer {admin_token}"
  → 403 Forbidden: "admins cannot apply for leave"

# Employee identity comes from JWT
# Cannot be spoofed or forged
```

---

## Key Security Points

1. **Authentication is Required**: All leave endpoints require valid JWT
2. **Role is Verified**: Admin endpoints check for `admin` role
3. **Identity is Token-Based**: Employee ID derived from JWT (cannot be overridden)
4. **Audit Trail**: Admin actions recorded with `approved_by` field
5. **Unique Employee Records**: User-Employee 1:1 relationship enforced
6. **Data Integrity**: Foreign keys prevent orphaned records

---

## Error Response Format

All errors follow this format:

```json
{
  "success": false,
  "message": "Human-readable error message",
  "fields": {
    "field_name": "Specific field error (if validation error)"
  }
}
```

**HTTP Status Codes**:

- `200` - Success (GET, POST with no creation)
- `201` - Created (POST with new resource)
- `400` - Bad request (validation error)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not found (resource doesn't exist)
- `409` - Conflict (unique constraint violated)
- `500` - Internal server error
