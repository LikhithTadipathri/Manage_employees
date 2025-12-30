# Leave Policies & Balances - Implementation Summary

## Overview

This feature adds leave balance management and auto-deduction functionality to the employee leave system. Employees now have limited leave balances that are checked and automatically deducted when leave is approved.

## Changes Made

### 1. Database Schema

**New Migration: `008_create_leave_balances_table.up.sql`**

- Created `leave_balances` table with:
  - `id` (Primary Key)
  - `employee_id` (Foreign Key to employees)
  - `leave_type` (VARCHAR 50)
  - `balance` (INTEGER)
  - Unique constraint on (employee_id, leave_type)
  - Indexes for faster lookups
  - Auto-initialization of default balances for existing employees

**Down Migration: `008_create_leave_balances_table.down.sql`**

- Drops the leave_balances table

### 2. Models

**`models/leave/leave.go` - Added:**

- `LeaveBalance` struct:
  - `ID`, `EmployeeID`, `LeaveType`, `Balance`, `CreatedAt`, `UpdatedAt`
- `DefaultLeaveBalances()` function:
  - Returns map with default balances:
    - ANNUAL: 10 days
    - SICK: 15 days
    - CASUAL: 10 days

### 3. Repository Layer

**`repositories/postgres/leave_repo.go` - Added Methods:**

- `GetLeaveBalance(employeeID int, leaveType LeaveType)` - Get balance for specific leave type
- `GetEmployeeLeaveBalances(employeeID int)` - Get all balances for an employee
- `InitializeLeaveBalances(employeeID int)` - Initialize default balances for new employee
- `UpdateLeaveBalance(employeeID int, leaveType LeaveType, newBalance int)` - Update balance
- `DeductLeaveBalance(employeeID int, leaveType LeaveType, days int)` - Deduct days from balance with validation

### 4. Service Layer

**`services/leave/leave_service.go` - Updated & Added:**

**Updated Methods:**

- `ApplyLeave()`:

  - Added balance check before creating leave request
  - Auto-rejects if balance is insufficient
  - Only checks balance for ANNUAL, SICK, CASUAL leave types
  - Other types (PERSONAL, MATERNITY, UNPAID) have no balance limits

- `ApproveLeave()`:
  - Added auto-deduction from balance when leave is approved
  - Deduction only happens for ANNUAL, SICK, CASUAL leave types

**New Methods:**

- `GetEmployeeLeaveBalance(userID, leaveType)` - Get balance by type for current employee
- `GetEmployeeLeaveBalances(userID)` - Get all balances for current employee
- `InitializeLeaveBalances(employeeID)` - Initialize default balances

### 5. HTTP Handlers

**`http/handlers/employee_handler.go` - Updated:**

- Modified constructor to accept leave service
- `CreateEmployee()` now initializes leave balances for new employees
- Added `NewEmployeeHandlerWithLeave()` constructor

**`http/handlers/leave_handler.go` - Added:**

- `GetMyLeaveBalance()` (GET `/leave/balance`) - Get all leave balances for current employee
- `GetMyLeaveBalanceByType()` (GET `/leave/balance/{type}`) - Get specific leave type balance

### 6. API Routes

**`http/server.go` - Updated:**

- Added new routes under `/leave`:
  - `GET /leave/balance` - Get all leave balances
  - `GET /leave/balance/{type}` - Get balance by type

## Feature Behavior

### Balance Check & Auto-Rejection

1. When employee applies for leave (ANNUAL, SICK, or CASUAL):
   - System checks current balance
   - If requested days > available balance: Request is rejected with error message
   - If balance insufficient: HTTP 400 with validation error

### Auto-Deduction

1. When admin approves leave (ANNUAL, SICK, or CASUAL):
   - Requested days are automatically deducted from balance
   - If deduction fails (balance became insufficient): Approval fails
   - No double-deduction: Only happens once on approval

### Leave Types Without Balance Limits

- PERSONAL, MATERNITY, UNPAID leaves have no balance limits
- These can be requested freely without balance checks
- No deduction on approval

## Testing Endpoints

### Get Leave Balance

```
GET /leave/balance
Authorization: Bearer <token>
Response: List of all leave balances with available days for each type
```

### Get Specific Leave Type Balance

```
GET /leave/balance/ANNUAL
Authorization: Bearer <token>
Response: LeaveBalance object with current balance for ANNUAL leave
```

### Apply for Leave (with auto-validation)

```
POST /leave/apply
Authorization: Bearer <token>
{
  "leave_type": "ANNUAL",
  "start_date": "2025-12-20",
  "end_date": "2025-12-22",
  "reason": "Holiday"
}
Response: Will reject if balance is insufficient
```

### Approve Leave (with auto-deduction)

```
POST /leave/approve/{id}
Authorization: Bearer <token> (Admin only)
Response: Leaves balance is automatically deducted
```

## Default Balances

- **Annual Leave**: 10 days
- **Sick Leave**: 15 days
- **Casual Leave**: 10 days

## Migration Instructions

1. Run migration `008_create_leave_balances_table.up.sql` to create the table
2. Migration auto-initializes balances for all existing employees
3. New employees get default balances initialized when their employee record is created via API

## Notes

- Leave balances are checked when applying but the request is still created in PENDING status
- Balance deduction happens only when the leave is APPROVED
- If a rejected leave request existed before, it doesn't restore the balance (since no balance was deducted)
- Cancelled approved leaves do NOT restore balance (as per typical HR policies)
