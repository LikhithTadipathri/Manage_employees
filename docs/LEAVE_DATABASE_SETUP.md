# Leave Management - Database Setup Guide

## 📋 Overview

This guide explains the database changes made to support Leave Management.

---

## 🗄️ New Table: leave_requests

### Purpose

Stores all leave requests submitted by employees with their approval/rejection status.

### Schema

```sql
CREATE TABLE leave_requests (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL,
    leave_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    reason TEXT,
    days_count INTEGER NOT NULL,
    approved_by INTEGER,
    approval_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL
);
```

### Field Descriptions

| Field         | Type        | Description                                              |
| ------------- | ----------- | -------------------------------------------------------- |
| id            | SERIAL      | Primary key, auto-increment                              |
| employee_id   | INTEGER     | FK to employees table (mandatory)                        |
| leave_type    | VARCHAR(50) | Type: annual, sick, casual, maternity, paternity, unpaid |
| status        | VARCHAR(20) | pending, approved, rejected, cancelled                   |
| start_date    | TIMESTAMP   | Leave start date (inclusive)                             |
| end_date      | TIMESTAMP   | Leave end date (inclusive)                               |
| reason        | TEXT        | Reason for leave                                         |
| days_count    | INTEGER     | Working days (weekdays only)                             |
| approved_by   | INTEGER     | FK to users table (admin who approved)                   |
| approval_date | TIMESTAMP   | When approved/rejected                                   |
| created_at    | TIMESTAMP   | When request was created                                 |
| updated_at    | TIMESTAMP   | Last modification time                                   |

---

## 🔑 Foreign Key Relationships

### employee_id → employees(id)

- **Cascade Delete**: YES
- **Null Allowed**: NO
- **Ensures**: Leave requests cannot exist without employee

When an employee is deleted, all their leave requests are deleted.

### approved_by → users(id)

- **Set Null on Delete**: YES
- **Null Allowed**: YES
- **Default**: NULL (until approved/rejected)

When an admin user is deleted, the approved_by reference becomes NULL.

---

## 📑 Indexes

Four indexes created for query performance:

### 1. idx_leave_requests_employee_id

```sql
CREATE INDEX idx_leave_requests_employee_id ON leave_requests(employee_id);
```

- **Purpose**: Fast queries by employee
- **Used by**: `/leave/my-requests` endpoint
- **Query**: `SELECT * FROM leave_requests WHERE employee_id = ?`

### 2. idx_leave_requests_status

```sql
CREATE INDEX idx_leave_requests_status ON leave_requests(status);
```

- **Purpose**: Fast filtering by status
- **Used by**: `/leave/all?status=pending` queries
- **Query**: `SELECT * FROM leave_requests WHERE status = ?`

### 3. idx_leave_requests_start_date

```sql
CREATE INDEX idx_leave_requests_start_date ON leave_requests(start_date);
```

- **Purpose**: Fast queries by date range
- **Used by**: Calendar views, date-based reports
- **Query**: `SELECT * FROM leave_requests WHERE start_date BETWEEN ? AND ?`

### 4. idx_leave_requests_created_at

```sql
CREATE INDEX idx_leave_requests_created_at ON leave_requests(created_at);
```

- **Purpose**: Fast sorting by creation date
- **Used by**: Timeline views, recent requests
- **Query**: `SELECT * FROM leave_requests ORDER BY created_at DESC`

---

## 🔄 PostgreSQL vs SQLite

### PostgreSQL Schema (Production)

```sql
CREATE TABLE leave_requests (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL,
    leave_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    reason TEXT,
    days_count INTEGER NOT NULL,
    approved_by INTEGER,
    approval_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL
);
```

### SQLite Schema (Development Fallback)

```sql
CREATE TABLE leave_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    employee_id INTEGER NOT NULL,
    leave_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    reason TEXT,
    days_count INTEGER NOT NULL,
    approved_by INTEGER,
    approval_date DATETIME,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL
);
```

**Differences:**

- SERIAL → INTEGER PRIMARY KEY AUTOINCREMENT
- VARCHAR → TEXT
- TIMESTAMP → DATETIME

Both support the same queries through placeholder conversion.

---

## 📊 Relationships Diagram

```
users
  ├── id (PK)
  └── role
      │
      └── admin

employees
  ├── id (PK)
  └── FK to users (for login)

leave_requests (NEW)
  ├── id (PK)
  ├── employee_id (FK → employees.id) ON DELETE CASCADE
  └── approved_by (FK → users.id) ON DELETE SET NULL
```

---

## 🗄️ Migration Files

### Up Migration: 004_create_leave_requests_table.up.sql

Creates the table with all necessary constraints and indexes.

**Location**: `migrations/004_create_leave_requests_table.up.sql`

**File Contents**:

```sql
CREATE TABLE IF NOT EXISTS leave_requests (...)
CREATE INDEX IF NOT EXISTS idx_leave_requests_employee_id (...)
CREATE INDEX IF NOT EXISTS idx_leave_requests_status (...)
CREATE INDEX IF NOT EXISTS idx_leave_requests_start_date (...)
CREATE INDEX IF NOT EXISTS idx_leave_requests_created_at (...)
```

### Down Migration: 004_create_leave_requests_table.down.sql

Rolls back the table creation.

**Location**: `migrations/004_create_leave_requests_table.down.sql`

**File Contents**:

```sql
DROP TABLE IF EXISTS leave_requests CASCADE;
```

---

## 🔍 Common Queries

### Get all pending leave requests

```sql
SELECT * FROM leave_requests
WHERE status = 'pending'
ORDER BY created_at DESC;
```

### Get employee's leave history

```sql
SELECT * FROM leave_requests
WHERE employee_id = 5
ORDER BY created_at DESC;
```

### Get approved leaves for a date range

```sql
SELECT * FROM leave_requests
WHERE status = 'approved'
AND start_date >= '2025-12-01'
AND end_date <= '2025-12-31'
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

### Get admin's approvals

```sql
SELECT * FROM leave_requests
WHERE approved_by = 1
AND status IN ('approved', 'rejected')
ORDER BY approval_date DESC;
```

### Check overlapping leaves

```sql
SELECT e1.id, e1.employee_id, e1.start_date, e1.end_date,
       e2.id, e2.employee_id, e2.start_date, e2.end_date
FROM leave_requests e1
JOIN leave_requests e2 ON e1.employee_id = e2.employee_id
WHERE e1.id < e2.id
AND e1.status = 'approved'
AND e2.status = 'approved'
AND e1.start_date <= e2.end_date
AND e1.end_date >= e2.start_date;
```

---

## 🚀 Auto-Initialization

The table is **automatically created** when the application starts if it doesn't exist.

### Flow:

1. Application connects to database
2. `InitializeSchema()` is called
3. If table doesn't exist, it's created
4. Indexes are created
5. Both PostgreSQL and SQLite supported

### Code Location:

`utils/helpers/db.go` → `InitializeSchema()` function

---

## 🔐 Data Integrity

### Constraints Enforced

1. **NOT NULL Constraints**

   - employee_id must have value
   - leave_type must have value
   - status has default: 'pending'
   - start_date, end_date must have values
   - days_count must have value
   - created_at, updated_at must have values

2. **Foreign Key Constraints**

   - employee_id must exist in employees table
   - approved_by (if set) must exist in users table
   - Cascade delete on employee deletion
   - Set null on admin user deletion

3. **Status Values**

   - Limited to: pending, approved, rejected, cancelled
   - Enforced in application layer

4. **Days Count**
   - Only working days (Mon-Fri)
   - Calculated by application
   - Must be positive integer

---

## 📈 Performance Considerations

### Index Coverage

All common query patterns are indexed:

✅ **Filter by employee**: idx_leave_requests_employee_id  
✅ **Filter by status**: idx_leave_requests_status  
✅ **Filter by date**: idx_leave_requests_start_date  
✅ **Sort by creation**: idx_leave_requests_created_at

### Query Performance

- Employee requests: **O(1)** with index
- Status filtering: **O(1)** with index
- Date range queries: **O(1)** with index
- Timeline sorting: **O(1)** with index

### Storage

Expected table size (estimates):

- 1,000 employees × 10 years × 2 requests/month = 240,000 records
- ~30KB per record (approximate)
- Total: ~7.2 MB (manageable)

---

## 🛠️ Maintenance

### Regular Tasks

1. **Backup**: Include leave_requests table in backups
2. **Monitor**: Check table/index sizes regularly
3. **Archive**: Consider archiving old approved/rejected leaves
4. **Analyze**: Run ANALYZE command periodically

### PostgreSQL Commands

```sql
-- Analyze table
ANALYZE leave_requests;

-- Check table size
SELECT pg_size_pretty(pg_total_relation_size('leave_requests'));

-- Reindex if needed
REINDEX TABLE leave_requests;

-- Vacuum to recover space
VACUUM leave_requests;
```

### SQLite Commands

```sql
-- Analyze
ANALYZE leave_requests;

-- Check size
SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size();

-- Vacuum
VACUUM;
```

---

## ⚠️ Important Notes

1. **Cascade Delete**: Deleting an employee deletes all their leaves
2. **Null Handling**: approved_by is NULL until approved/rejected
3. **Status Immutable**: Once approved/rejected, cannot change to pending
4. **Days Calculation**: Weekends not counted, holidays not subtracted
5. **Timezone**: All timestamps should be in UTC

---

## 🔄 Rollback Procedure

If needed to remove the leave management system:

### Step 1: Run Down Migration

```sql
DROP TABLE IF EXISTS leave_requests CASCADE;
```

### Step 2: Remove Application Code

- Delete `models/leave/` directory
- Delete `services/leave/` directory
- Remove leave handler from handlers
- Remove leave routes from server.go

### Step 3: Restart Application

- Application will work without leave management
- No errors if code is properly removed

---

## ✅ Verification

### Verify Table Exists (PostgreSQL)

```sql
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name = 'leave_requests';
```

### Verify Table Exists (SQLite)

```sql
SELECT name FROM sqlite_master
WHERE type='table'
AND name='leave_requests';
```

### Verify Indexes

```sql
-- PostgreSQL
SELECT indexname FROM pg_indexes
WHERE tablename = 'leave_requests';

-- SQLite
SELECT name FROM sqlite_master
WHERE type='index'
AND tbl_name='leave_requests';
```

---

## 📞 Support

For issues or questions:

- Check migration files: `migrations/004_*.sql`
- Review schema in: `docs/LEAVE_MANAGEMENT.md`
- Check auto-init code: `utils/helpers/db.go`
- Review tests: `docs/LEAVE_API_TEST_GUIDE.md`

---

All done! The database is ready for Leave Management. 🎉
