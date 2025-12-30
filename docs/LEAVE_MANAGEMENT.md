# Leave Management System - Integration Documentation

## 🎯 Overview

A complete Leave Management system integrated into your existing Employee Management System. Employees can apply for leave, admins can approve/reject, and all requests are tracked in the database.

---

## 📊 Database Schema

### leave_requests Table

```sql
CREATE TABLE leave_requests (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL,           -- Links to employees table
    leave_type VARCHAR(50) NOT NULL,        -- Type of leave (annual, sick, etc)
    status VARCHAR(20) DEFAULT 'pending',   -- pending, approved, rejected, cancelled
    start_date TIMESTAMP NOT NULL,          -- Leave start date
    end_date TIMESTAMP NOT NULL,            -- Leave end date
    reason TEXT,                            -- Reason for leave
    days_count INTEGER NOT NULL,            -- Working days (weekdays only)
    approved_by INTEGER,                    -- Admin user ID who approved
    approval_date TIMESTAMP,                -- When it was approved/rejected
    created_at TIMESTAMP NOT NULL,          -- When request was created
    updated_at TIMESTAMP NOT NULL,          -- Last update timestamp
    FOREIGN KEY (employee_id) REFERENCES employees(id),
    FOREIGN KEY (approved_by) REFERENCES users(id)
);
```

**Indexes created for performance:**

- `idx_leave_requests_employee_id` - Fast employee-specific queries
- `idx_leave_requests_status` - Fast status filtering
- `idx_leave_requests_start_date` - Fast date-based queries
- `idx_leave_requests_created_at` - Fast timeline queries

---

## 📁 New Files Created

### Models

- `models/leave/leave.go` - Data models and validation

### Repositories

- `repositories/postgres/leave_repo.go` - Database operations

### Services

- `services/leave/leave_service.go` - Business logic

### Handlers

- `http/handlers/leave_handler.go` - HTTP request handlers

### Migrations

- `migrations/004_create_leave_requests_table.up.sql` - Create table
- `migrations/004_create_leave_requests_table.down.sql` - Rollback

---

## 🔐 API Endpoints

### Employee Endpoints (Authenticated)

#### 1. **Apply for Leave**

```
POST /leave/apply
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

Body:
{
    "leave_type": "annual",
    "start_date": "2025-12-20T00:00:00Z",
    "end_date": "2025-12-25T00:00:00Z",
    "reason": "Year-end vacation"
}

Response (201 Created):
{
    "status": "success",
    "message": "Leave request submitted successfully",
    "data": {
        "id": 1,
        "employee_id": 5,
        "leave_type": "annual",
        "status": "pending",
        "start_date": "2025-12-20T00:00:00Z",
        "end_date": "2025-12-25T00:00:00Z",
        "reason": "Year-end vacation",
        "days_count": 4,
        "created_at": "2025-12-10T10:00:00Z",
        "updated_at": "2025-12-10T10:00:00Z"
    }
}
```

#### 2. **Get My Leave Requests**

```
GET /leave/my-requests
Authorization: Bearer <JWT_TOKEN>
Optional Query: ?status=pending

Response (200 OK):
{
    "status": "success",
    "message": "Leave requests retrieved successfully",
    "data": {
        "count": 2,
        "leave_requests": [
            {
                "id": 1,
                "employee_id": 5,
                "leave_type": "annual",
                "status": "pending",
                "start_date": "2025-12-20T00:00:00Z",
                "end_date": "2025-12-25T00:00:00Z",
                "reason": "Year-end vacation",
                "days_count": 4,
                "created_at": "2025-12-10T10:00:00Z",
                "updated_at": "2025-12-10T10:00:00Z"
            },
            ...
        ]
    }
}
```

#### 3. **Cancel Leave Request**

```
DELETE /leave/cancel/{id}
Authorization: Bearer <JWT_TOKEN>

Response (200 OK):
{
    "status": "success",
    "message": "Leave request cancelled successfully",
    "data": null
}
```

---

### Admin Endpoints (Authenticated + Admin Role Only)

#### 1. **Get All Leave Requests**

```
GET /leave/all
Authorization: Bearer <JWT_TOKEN>
Optional Query: ?status=pending

Response (200 OK):
{
    "status": "success",
    "message": "Leave requests retrieved successfully",
    "data": {
        "count": 5,
        "leave_requests": [
            {
                "id": 1,
                "employee_id": 5,
                "leave_type": "annual",
                "status": "pending",
                "start_date": "2025-12-20T00:00:00Z",
                "end_date": "2025-12-25T00:00:00Z",
                "reason": "Year-end vacation",
                "days_count": 4,
                "employee_name": "John Doe",
                "employee_email": "john.doe@company.com",
                "created_at": "2025-12-10T10:00:00Z",
                "updated_at": "2025-12-10T10:00:00Z"
            },
            ...
        ]
    }
}
```

#### 2. **Approve Leave Request**

```
POST /leave/approve/{id}
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

Response (200 OK):
{
    "status": "success",
    "message": "Leave request approved successfully",
    "data": null
}
```

#### 3. **Reject Leave Request**

```
POST /leave/reject/{id}
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

Response (200 OK):
{
    "status": "success",
    "message": "Leave request rejected successfully",
    "data": null
}
```

---

## 🎭 Leave Types Supported

- `annual` - Annual/vacation leave
- `sick` - Sick leave
- `casual` - Casual/personal leave
- `maternity` - Maternity leave
- `paternity` - Paternity leave
- `unpaid` - Unpaid leave

---

## 📋 Leave Status Flow

```
pending → approved/rejected
   ↓
cancelled (employee can cancel only if pending)
```

**Status Meanings:**

- **pending** - Awaiting admin approval
- **approved** - Admin approved the leave
- **rejected** - Admin rejected the leave
- **cancelled** - Employee cancelled pending request

---

## ⚙️ Implementation Details

### Key Features

1. **Automatic Day Calculation**

   - Excludes weekends (only counts Monday-Friday)
   - Prevents invalid date ranges
   - Validates at least one working day

2. **Role-Based Access**

   - Employees: Can only see/manage their own leaves
   - Admins: Can see all leaves and approve/reject

3. **Database Integrity**

   - Foreign keys ensure employee exists
   - Cascade delete on employee removal
   - Null handling for optional fields

4. **Timestamp Tracking**
   - `created_at` - When request was made
   - `updated_at` - Last modification
   - `approval_date` - When approved/rejected

---

## 🚀 Usage Examples

### Create New Leave Request

```bash
curl -X POST http://localhost:8080/leave/apply \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "leave_type": "annual",
    "start_date": "2025-12-20T00:00:00Z",
    "end_date": "2025-12-25T00:00:00Z",
    "reason": "Holiday vacation"
  }'
```

### Get My Leave Requests

```bash
curl http://localhost:8080/leave/my-requests \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Get My Pending Leaves Only

```bash
curl "http://localhost:8080/leave/my-requests?status=pending" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Cancel a Leave (as Employee)

```bash
curl -X DELETE http://localhost:8080/leave/cancel/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Admin: Get All Leave Requests

```bash
curl http://localhost:8080/leave/all \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

### Admin: Approve Leave

```bash
curl -X POST http://localhost:8080/leave/approve/1 \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

### Admin: Reject Leave

```bash
curl -X POST http://localhost:8080/leave/reject/1 \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

---

## 🔄 Integration with Existing System

### Modified Files

- `http/server.go` - Added leave routes and handler initialization
- `utils/helpers/db.go` - Added schema for both PostgreSQL and SQLite
- `errors/common.go` - Added NotFoundError and ForbiddenError types
- `errors/validation.go` - Added AddField method for error chaining

### New Dependencies

- Uses existing `JWTMiddleware` for authentication
- Uses existing `GetUserFromContext` to extract user info
- Uses existing error handling patterns
- Follows existing response format

---

## 📝 Database Migration Files

### Up Migration (004_create_leave_requests_table.up.sql)

Creates the leave_requests table with all necessary indexes and constraints.

### Down Migration (004_create_leave_requests_table.down.sql)

Drops the leave_requests table for rollback.

---

## ✅ Validation Rules

### Apply Leave Request

- `leave_type` - Required, non-empty string
- `start_date` - Required, valid timestamp
- `end_date` - Required, valid timestamp, must be after start_date
- `reason` - Required, max 500 characters
- Must include at least one working day

### Error Responses

```json
{
  "status": "error",
  "message": "Validation failed",
  "data": {
    "leave_type": "leave type is required",
    "start_date": "start date is required",
    "dates": "start date cannot be after end date",
    "reason": "reason cannot exceed 500 characters"
  }
}
```

---

## 🔍 Query Examples

### Get all pending leaves from a specific employee

```sql
SELECT * FROM leave_requests
WHERE employee_id = 5 AND status = 'pending'
ORDER BY created_at DESC;
```

### Get all approved leaves for this month

```sql
SELECT * FROM leave_requests
WHERE status = 'approved'
AND EXTRACT(MONTH FROM start_date) = EXTRACT(MONTH FROM CURRENT_DATE)
ORDER BY start_date;
```

### Get leave statistics

```sql
SELECT
    status,
    COUNT(*) as count,
    SUM(days_count) as total_days
FROM leave_requests
GROUP BY status;
```

---

## 🛠️ Configuration

No additional configuration needed. The system uses:

- Existing JWT configuration for authentication
- Existing database connection settings
- Existing user roles (admin/user)
- Existing employee ID mapping

---

## 📚 Related Documentation

- Database: See `migrations/SCHEMA.md`
- Employee Management: See employee models and handlers
- Authentication: See `http/middlewares/jwt.go`
- Response Format: See `http/response/response.go`

---

## 🎯 Next Steps

1. Run migrations: `migrations/004_create_leave_requests_table.up.sql`
2. Rebuild/restart the application
3. Test endpoints with provided examples
4. Import Postman collection for full API testing

---
