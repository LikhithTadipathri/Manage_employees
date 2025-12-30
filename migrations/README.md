# PostgreSQL Database Migrations

**Date:** December 4, 2025  
**Database:** PostgreSQL  
**Tool:** Plain SQL migration files (no external tool required)

---

## 📋 Migration Structure

```
migrations/
├── 001_create_users_table.up.sql       ← Create users table
├── 001_create_users_table.down.sql     ← Drop users table (rollback)
├── 002_create_employees_table.up.sql   ← Create employees table
├── 002_create_employees_table.down.sql ← Drop employees table (rollback)
├── 003_create_indexes.up.sql           ← Create indexes
└── 003_create_indexes.down.sql         ← Drop indexes (rollback)
```

---

## 🎯 Migrations Overview

### **001_create_users_table**

**Purpose:** Create the users table for authentication

**Table Structure:**

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Columns:**

- `id` - Primary key, auto-increment
- `username` - Unique username for login
- `email` - Unique email address
- `password_hash` - Bcrypt hashed password
- `role` - User role: 'admin' or 'user'
- `created_at` - Record creation time
- `updated_at` - Last update time

---

### **002_create_employees_table**

**Purpose:** Create the employees table for employee records

**Table Structure:**

```sql
CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    phone VARCHAR(20) NOT NULL,
    position VARCHAR(100) NOT NULL,
    salary DECIMAL(10, 2) NOT NULL,
    hired_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Columns:**

- `id` - Primary key, auto-increment
- `first_name` - Employee first name
- `last_name` - Employee last name
- `email` - Unique email address
- `phone` - Contact phone number
- `position` - Job title/position
- `salary` - Annual salary (2 decimals)
- `hired_date` - Employment start date
- `created_at` - Record creation time
- `updated_at` - Last update time

---

### **003_create_indexes**

**Purpose:** Create indexes for query performance

**Indexes Created:**

**Users Table Indexes:**

- `idx_users_username` - On username (frequent login searches)
- `idx_users_email` - On email (unique constraint support)
- `idx_users_role` - On role (filter by admin/user)
- `idx_users_created_at` - On created_at (sorting/filtering)

**Employees Table Indexes:**

- `idx_employees_email` - On email (unique constraint support)
- `idx_employees_created_at` - On created_at (sorting/filtering)
- `idx_employees_position` - On position (search by job title)
- `idx_employees_hired_date` - On hired_date (historical queries)
- `idx_employees_name` - Composite on (last_name, first_name)
- `idx_employees_salary_date` - Composite on (salary, hired_date)

---

## 🚀 How to Run Migrations

### **Option 1: Using psql CLI**

```bash
# Connect to PostgreSQL
psql -h localhost -U postgres -d employee_db -f migrations/001_create_users_table.up.sql
psql -h localhost -U postgres -d employee_db -f migrations/002_create_employees_table.up.sql
psql -h localhost -U postgres -d employee_db -f migrations/003_create_indexes.up.sql
```

### **Option 2: Using DBeaver GUI**

1. Open DBeaver
2. Connect to PostgreSQL database
3. Open SQL Editor
4. Copy content of migration file
5. Execute

### **Option 3: Using Go Application (Recommended)**

See `migration_runner.go` file in this directory.

---

## 🔄 Migration Workflow

### **Apply All Migrations (Forward):**

```bash
# Run in order:
# 1. Users table
# 2. Employees table
# 3. Indexes
```

### **Rollback Migrations (Backward):**

```bash
# Reverse order:
# 1. Drop indexes
# 2. Drop employees table
# 3. Drop users table
```

---

## 📊 Current Schema Summary

| Table     | Columns | Rows | Purpose                        |
| --------- | ------- | ---- | ------------------------------ |
| users     | 7       | 3+   | Authentication & authorization |
| employees | 10      | 10+  | Employee records               |

---

## ✅ Migration Status

**Completed:**

- ✅ Users table created
- ✅ Employees table created
- ✅ Indexes created

**Current State:**

- PostgreSQL database has all tables
- All indexes applied
- Schema is normalized and optimized

---

## 🛠️ Common Migration Tasks

### **Add a New Column to Users:**

```sql
-- migrations/004_add_department_to_users.up.sql
ALTER TABLE users ADD COLUMN department VARCHAR(100);

-- migrations/004_add_department_to_users.down.sql
ALTER TABLE users DROP COLUMN department;
```

### **Add a Foreign Key:**

```sql
-- migrations/005_add_manager_to_employees.up.sql
ALTER TABLE employees ADD COLUMN manager_id INTEGER REFERENCES employees(id);

-- migrations/005_add_manager_to_employees.down.sql
ALTER TABLE employees DROP COLUMN manager_id;
```

### **Modify Column Type:**

```sql
-- migrations/006_increase_salary_precision.up.sql
ALTER TABLE employees ALTER COLUMN salary TYPE DECIMAL(12, 2);

-- migrations/006_increase_salary_precision.down.sql
ALTER TABLE employees ALTER COLUMN salary TYPE DECIMAL(10, 2);
```

---

## 📝 Migration File Naming Convention

**Format:** `NNN_description.{up|down}.sql`

- `NNN` - Sequential number (001, 002, 003, ...)
- `description` - Brief description (snake_case, lowercase)
- `up` - Apply migration (forward)
- `down` - Rollback migration (backward)

**Examples:**

- `001_create_users_table.up.sql`
- `002_create_employees_table.down.sql`
- `005_add_manager_to_employees.up.sql`

---

## ⚠️ Best Practices

✅ **DO:**

- Always create matching `.down.sql` files
- Test migrations on development first
- Keep migrations small and focused
- Add comments explaining the change
- Use transactions in production
- Backup database before migrations

❌ **DON'T:**

- Modify existing migration files (create new ones)
- Lose data in rollback scripts
- Create multiple tables in one migration
- Skip version numbers (001, 002, 004 is wrong)

---

## 🔍 Verify Migrations

### **Check Tables:**

```bash
psql -h localhost -U postgres -d employee_db -c "\dt"
```

### **Check Indexes:**

```bash
psql -h localhost -U postgres -d employee_db -c "\di"
```

### **Check Columns:**

```bash
psql -h localhost -U postgres -d employee_db -c "\d users"
psql -h localhost -U postgres -d employee_db -c "\d employees"
```

---

**All migration files are ready to use! 🎯**
