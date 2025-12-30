# Leave Management API - Quick Test Guide

## 🧪 Testing the Leave Management System

This guide provides quick test commands to verify the Leave Management implementation.

---

## Prerequisites

1. Application running on `http://localhost:8080`
2. Valid JWT tokens for testing
3. Sample employees and users in the database

---

## 1️⃣ Setup: Get Admin & Employee Tokens

### Login as Admin

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

**Expected Response:**

```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@company.com",
      "role": "admin"
    },
    "expires_at": 1702156800
  }
}
```

**Save the token as `$ADMIN_TOKEN`**

---

### Login as Employee

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "password123"
  }'
```

**Save the token as `$EMPLOYEE_TOKEN`**

---

## 2️⃣ Employee Tests

### Test 2.1: Apply for Leave

```bash
curl -X POST http://localhost:8080/leave/apply \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "leave_type": "annual",
    "start_date": "2025-12-20T00:00:00Z",
    "end_date": "2025-12-25T00:00:00Z",
    "reason": "Year-end vacation with family"
  }'
```

**Expected Response (201):**

```json
{
  "success": true,
  "message": "Leave request submitted successfully",
  "data": {
    "id": 1,
    "employee_id": 2,
    "leave_type": "annual",
    "status": "pending",
    "start_date": "2025-12-20T00:00:00Z",
    "end_date": "2025-12-25T00:00:00Z",
    "reason": "Year-end vacation with family",
    "days_count": 4,
    "created_at": "2025-12-10T10:30:00Z",
    "updated_at": "2025-12-10T10:30:00Z"
  }
}
```

**Note: Days count = 4 (excludes weekends: Sat 20, Sun 21, Sat 27, Sun 28)**

---

### Test 2.2: View My Leave Requests

```bash
curl -X GET http://localhost:8080/leave/my-requests \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN"
```

**Expected Response (200):**

```json
{
  "success": true,
  "message": "Leave requests retrieved successfully",
  "data": {
    "count": 1,
    "leave_requests": [
      {
        "id": 1,
        "employee_id": 2,
        "leave_type": "annual",
        "status": "pending",
        "start_date": "2025-12-20T00:00:00Z",
        "end_date": "2025-12-25T00:00:00Z",
        "reason": "Year-end vacation with family",
        "days_count": 4,
        "created_at": "2025-12-10T10:30:00Z",
        "updated_at": "2025-12-10T10:30:00Z"
      }
    ]
  }
}
```

---

### Test 2.3: Filter by Status

```bash
curl -X GET "http://localhost:8080/leave/my-requests?status=pending" \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN"
```

---

### Test 2.4: Apply Another Leave (Sick Leave)

```bash
curl -X POST http://localhost:8080/leave/apply \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "leave_type": "sick",
    "start_date": "2025-12-15T00:00:00Z",
    "end_date": "2025-12-15T00:00:00Z",
    "reason": "Doctor appointment"
  }'
```

**Expected: 1 day of leave (15 Dec is Monday)**

---

### Test 2.5: Cancel Leave Request

```bash
curl -X DELETE http://localhost:8080/leave/cancel/1 \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN"
```

**Expected Response (200):**

```json
{
  "success": true,
  "message": "Leave request cancelled successfully",
  "data": null
}
```

---

### Test 2.6: Try Invalid Leave (Future Cancel Failed)

```bash
curl -X DELETE http://localhost:8080/leave/cancel/1 \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN"
```

**Expected Response (400):**

```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "status": "only pending leave requests can be cancelled"
  }
}
```

---

## 3️⃣ Admin Tests

### Test 3.1: View All Leave Requests

```bash
curl -X GET http://localhost:8080/leave/all \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

**Expected Response (200):**

```json
{
  "success": true,
  "message": "Leave requests retrieved successfully",
  "data": {
    "count": 2,
    "leave_requests": [
      {
        "id": 2,
        "employee_id": 2,
        "leave_type": "sick",
        "status": "pending",
        "start_date": "2025-12-15T00:00:00Z",
        "end_date": "2025-12-15T00:00:00Z",
        "reason": "Doctor appointment",
        "days_count": 1,
        "employee_name": "John Doe",
        "employee_email": "john.doe@company.com",
        "created_at": "2025-12-10T10:35:00Z",
        "updated_at": "2025-12-10T10:35:00Z"
      }
    ]
  }
}
```

---

### Test 3.2: Filter Pending Requests

```bash
curl -X GET "http://localhost:8080/leave/all?status=pending" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

---

### Test 3.3: Approve Leave Request

```bash
curl -X POST http://localhost:8080/leave/approve/2 \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json"
```

**Expected Response (200):**

```json
{
  "success": true,
  "message": "Leave request approved successfully",
  "data": null
}
```

---

### Test 3.4: Verify Approved Status

```bash
curl -X GET "http://localhost:8080/leave/all?status=approved" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

**Expected: Leave shows as approved with approval_date**

---

### Test 3.5: Reject a Pending Request

```bash
# First apply a new leave as employee
curl -X POST http://localhost:8080/leave/apply \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "leave_type": "casual",
    "start_date": "2025-12-24T00:00:00Z",
    "end_date": "2025-12-24T00:00:00Z",
    "reason": "Personal work"
  }'

# Then reject it as admin (replace ID with the new leave ID)
curl -X POST http://localhost:8080/leave/reject/3 \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

**Expected Response (200):**

```json
{
  "success": true,
  "message": "Leave request rejected successfully",
  "data": null
}
```

---

## 4️⃣ Error Test Cases

### Test 4.1: Missing Authorization Header

```bash
curl -X GET http://localhost:8080/leave/my-requests
```

**Expected Response (401):**

```json
{
  "success": false,
  "message": "missing authorization header"
}
```

---

### Test 4.2: Invalid Token

```bash
curl -X GET http://localhost:8080/leave/my-requests \
  -H "Authorization: Bearer invalid_token_123"
```

**Expected Response (401):**

```json
{
  "success": false,
  "message": "invalid or expired token"
}
```

---

### Test 4.3: Employee Accessing Admin Endpoint

```bash
curl -X GET http://localhost:8080/leave/all \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN"
```

**Expected Response (403):**

```json
{
  "success": false,
  "message": "admin access required"
}
```

---

### Test 4.4: Invalid Leave Request (Invalid Dates)

```bash
curl -X POST http://localhost:8080/leave/apply \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "leave_type": "annual",
    "start_date": "2025-12-25T00:00:00Z",
    "end_date": "2025-12-20T00:00:00Z",
    "reason": "Invalid date range"
  }'
```

**Expected Response (400):**

```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "dates": "start date cannot be after end date"
  }
}
```

---

### Test 4.5: Missing Required Field

```bash
curl -X POST http://localhost:8080/leave/apply \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "leave_type": "annual",
    "start_date": "2025-12-25T00:00:00Z"
  }'
```

**Expected Response (400):**

```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "end_date": "end date is required",
    "reason": "reason is required"
  }
}
```

---

### Test 4.6: Leave on Weekend Only

```bash
curl -X POST http://localhost:8080/leave/apply \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "leave_type": "annual",
    "start_date": "2025-12-20T00:00:00Z",
    "end_date": "2025-12-21T00:00:00Z",
    "reason": "Weekend only"
  }'
```

**Expected Response (400):**

```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "dates": "leave period must include at least one working day"
  }
}
```

---

### Test 4.7: Invalid Leave ID

```bash
curl -X DELETE http://localhost:8080/leave/cancel/999 \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN"
```

**Expected Response (404):**

```json
{
  "success": false,
  "message": "Leave request not found"
}
```

---

## 5️⃣ Leave Type Tests

Test different leave types:

```bash
# Maternity Leave
curl -X POST http://localhost:8080/leave/apply \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "leave_type": "maternity",
    "start_date": "2025-03-01T00:00:00Z",
    "end_date": "2025-05-31T00:00:00Z",
    "reason": "Maternity leave"
  }'

# Unpaid Leave
curl -X POST http://localhost:8080/leave/apply \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "leave_type": "unpaid",
    "start_date": "2026-01-05T00:00:00Z",
    "end_date": "2026-01-09T00:00:00Z",
    "reason": "Personal reasons"
  }'
```

---

## 📊 Test Summary Table

| Test Case        | Method | Endpoint                          | Expected Status | Notes                 |
| ---------------- | ------ | --------------------------------- | --------------- | --------------------- |
| Apply Leave      | POST   | /leave/apply                      | 201             | Valid dates required  |
| View My Requests | GET    | /leave/my-requests                | 200             | Returns user's leaves |
| Filter Status    | GET    | /leave/my-requests?status=pending | 200             | Query parameter       |
| Cancel Leave     | DELETE | /leave/cancel/{id}                | 200             | Only pending status   |
| View All (Admin) | GET    | /leave/all                        | 200             | Admin only            |
| Approve (Admin)  | POST   | /leave/approve/{id}               | 200             | Admin only            |
| Reject (Admin)   | POST   | /leave/reject/{id}                | 200             | Admin only            |
| Unauthorized     | ANY    | /\*                               | 401             | Missing token         |
| Forbidden        | GET    | /leave/all                        | 403             | Non-admin user        |
| Not Found        | DELETE | /leave/cancel/999                 | 404             | Invalid ID            |

---

## 🎯 Validation Rules to Test

✅ **Date Validation**

- Start date cannot be after end date
- At least one working day required
- Weekends excluded from count

✅ **Field Validation**

- leave_type: Required
- start_date: Required
- end_date: Required
- reason: Required, max 500 chars

✅ **Status Validation**

- Only pending can be cancelled
- Only pending can be approved/rejected

✅ **Permission Validation**

- Employees only see their own
- Only admins access /leave/all
- Only admins approve/reject

---

## 🚀 Troubleshooting

**Issue: "unauthorized" response**

- Check token is valid
- Check Authorization header format: `Bearer <token>`

**Issue: "admin access required"**

- Use admin token, not employee token
- Verify user role in database

**Issue: "validation failed"**

- Check all required fields provided
- Validate date formats (ISO 8601)
- Ensure dates are valid

**Issue: "employee_id not found"**

- Ensure employee record exists
- Match UserID with employee_id in database

---

## 📝 Notes

- All timestamps should be in UTC (ISO 8601 format)
- Day count calculation excludes weekends (Mon-Fri only)
- Leave requests are immutable once approved/rejected
- Admin user is determined by role = "admin"

---

Done! Use these tests to verify the Leave Management system. 🎉
