# PostgreSQL Schema Documentation

**Date:** December 4, 2025  
**Database:** PostgreSQL  
**Version:** 1.0

---

## 📊 Database Overview

**Database Name:** `employee_db`  
**Tables:** 2  
**Indexes:** 10  
**Total Columns:** 17

---

## 🗂️ Table Schemas

### Table 1: `users`

**Purpose:** Store user authentication and authorization data

**DDL:**

```sql
CREATE TABLE IF NOT EXISTS users (
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

| Column          | Type         | Constraints              | Description                    |
| --------------- | ------------ | ------------------------ | ------------------------------ |
| `id`            | SERIAL       | PRIMARY KEY              | Auto-increment user identifier |
| `username`      | VARCHAR(50)  | NOT NULL, UNIQUE         | Login username (unique)        |
| `email`         | VARCHAR(100) | NOT NULL, UNIQUE         | Email address (unique)         |
| `password_hash` | VARCHAR(255) | NOT NULL                 | Bcrypt hashed password         |
| `role`          | VARCHAR(20)  | NOT NULL, DEFAULT 'user' | Role: 'admin' or 'user'        |
| `created_at`    | TIMESTAMP    | NOT NULL                 | Creation timestamp (auto)      |
| `updated_at`    | TIMESTAMP    | NOT NULL                 | Last update timestamp (auto)   |

**Sample Data:**

```sql
INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
VALUES
  ('admin', 'admin@example.com', '$2a$10$...', 'admin', NOW(), NOW()),
  ('john_doe', 'john@example.com', '$2a$10$...', 'user', NOW(), NOW()),
  ('jane_smith', 'jane@example.com', '$2a$10$...', 'user', NOW(), NOW());
```

**Indexes:**

- `idx_users_username` - ON (username)
- `idx_users_email` - ON (email)
- `idx_users_role` - ON (role)
- `idx_users_created_at` - ON (created_at)

---

### Table 2: `employees`

**Purpose:** Store employee information and details

**DDL:**

```sql
CREATE TABLE IF NOT EXISTS employees (
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

| Column       | Type           | Constraints      | Description                           |
| ------------ | -------------- | ---------------- | ------------------------------------- |
| `id`         | SERIAL         | PRIMARY KEY      | Auto-increment employee ID            |
| `first_name` | VARCHAR(100)   | NOT NULL         | Employee first name                   |
| `last_name`  | VARCHAR(100)   | NOT NULL         | Employee last name                    |
| `email`      | VARCHAR(100)   | NOT NULL, UNIQUE | Work email (unique)                   |
| `phone`      | VARCHAR(20)    | NOT NULL         | Contact phone number                  |
| `position`   | VARCHAR(100)   | NOT NULL         | Job title/position                    |
| `salary`     | DECIMAL(10, 2) | NOT NULL         | Annual salary (10 digits, 2 decimals) |
| `hired_date` | TIMESTAMP      | NOT NULL         | Employment start date                 |
| `created_at` | TIMESTAMP      | NOT NULL         | Record creation (auto)                |
| `updated_at` | TIMESTAMP      | NOT NULL         | Last update (auto)                    |

**Sample Data:**

```sql
INSERT INTO employees (first_name, last_name, email, phone, position, salary, hired_date, created_at, updated_at)
VALUES
  ('Alice', 'Johnson', 'alice@company.com', '555-0101', 'Software Engineer', 95000.00, '2023-01-15', NOW(), NOW()),
  ('Bob', 'Smith', 'bob@company.com', '555-0102', 'Product Manager', 105000.00, '2023-02-20', NOW(), NOW());
```

**Indexes:**

- `idx_employees_email` - ON (email)
- `idx_employees_created_at` - ON (created_at)
- `idx_employees_position` - ON (position)
- `idx_employees_hired_date` - ON (hired_date)
- `idx_employees_name` - ON (last_name, first_name)
- `idx_employees_salary_date` - ON (salary, hired_date)

---

## 🔍 Indexes Summary

### Purpose and Performance Impact

**Single Column Indexes (8):**

```
idx_users_username        | Fast login lookups
idx_users_email           | Email verification
idx_users_role            | Filter by role
idx_users_created_at      | Historical queries
idx_employees_email       | Unique constraint support
idx_employees_created_at  | Sorting/filtering
idx_employees_position    | Search by job title
idx_employees_hired_date  | Date range queries
```

**Composite Indexes (2):**

```
idx_employees_name        | Fast name searches (last_name + first_name)
idx_employees_salary_date | Salary band with date range queries
```

---

## 📈 Data Relationships

### Current Schema (No Foreign Keys)

```
┌──────────────┐
│    users     │
├──────────────┤
│ id (PK)      │
│ username     │
│ email        │
│ role         │
│ password_hash│
│ created_at   │
│ updated_at   │
└──────────────┘

        (Independent)

┌──────────────┐
│  employees   │
├──────────────┤
│ id (PK)      │
│ first_name   │
│ last_name    │
│ email        │
│ phone        │
│ position     │
│ salary       │
│ hired_date   │
│ created_at   │
│ updated_at   │
└──────────────┘
```

**Note:** Currently no foreign key relationships. Can be added in future migrations if employees need to link to users.

---

## 🔐 Data Integrity

### Unique Constraints

```
users.username  - UNIQUE (no duplicate usernames)
users.email     - UNIQUE (no duplicate emails)
employees.email - UNIQUE (no duplicate employee emails)
```

### Primary Keys

```
users.id        - SERIAL PRIMARY KEY (auto-increment)
employees.id    - SERIAL PRIMARY KEY (auto-increment)
```

### Default Values

```
users.role              - DEFAULT 'user'
users.created_at        - DEFAULT CURRENT_TIMESTAMP
users.updated_at        - DEFAULT CURRENT_TIMESTAMP
employees.created_at    - DEFAULT CURRENT_TIMESTAMP
employees.updated_at    - DEFAULT CURRENT_TIMESTAMP
```

---

## 📝 SQL Queries Reference

### Users Table Queries

**Get user by username:**

```sql
SELECT * FROM users WHERE username = 'admin';
```

**Get user by ID:**

```sql
SELECT * FROM users WHERE id = 1;
```

**Get all users:**

```sql
SELECT * FROM users ORDER BY created_at DESC;
```

**Create user:**

```sql
INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
VALUES ('newuser', 'user@example.com', '$2a$10$...', 'user', NOW(), NOW())
RETURNING id;
```

**Update user role:**

```sql
UPDATE users SET role = 'admin', updated_at = NOW() WHERE id = 1;
```

**Delete user:**

```sql
DELETE FROM users WHERE id = 1;
```

---

### Employees Table Queries

**Get all employees (paginated):**

```sql
SELECT * FROM employees ORDER BY id DESC LIMIT 10 OFFSET 0;
```

**Get employee by ID:**

```sql
SELECT * FROM employees WHERE id = 1;
```

**Get employee by email:**

```sql
SELECT * FROM employees WHERE email = 'alice@company.com';
```

**Search employees:**

```sql
SELECT * FROM employees
WHERE first_name ILIKE '%alice%'
   OR last_name ILIKE '%johnson%'
   OR email ILIKE '%alice%'
   OR phone ILIKE '%555%'
   OR position ILIKE '%engineer%'
ORDER BY last_name, first_name
LIMIT 10;
```

**Get employees by position:**

```sql
SELECT * FROM employees
WHERE position = 'Software Engineer'
ORDER BY hired_date DESC;
```

**Get employees hired in date range:**

```sql
SELECT * FROM employees
WHERE hired_date BETWEEN '2023-01-01' AND '2023-12-31'
ORDER BY hired_date;
```

**Get employees by salary range:**

```sql
SELECT * FROM employees
WHERE salary BETWEEN 90000 AND 110000
ORDER BY salary DESC;
```

**Count total employees:**

```sql
SELECT COUNT(*) FROM employees;
```

**Count by position:**

```sql
SELECT position, COUNT(*)
FROM employees
GROUP BY position
ORDER BY COUNT(*) DESC;
```

**Average salary:**

```sql
SELECT AVG(salary) as avg_salary FROM employees;
```

**Create employee:**

```sql
INSERT INTO employees (first_name, last_name, email, phone, position, salary, hired_date, created_at, updated_at)
VALUES ('John', 'Doe', 'john@company.com', '555-1234', 'Developer', 95000, NOW(), NOW(), NOW())
RETURNING id;
```

**Update employee:**

```sql
UPDATE employees
SET first_name = 'Jonathan', position = 'Senior Developer', salary = 110000, updated_at = NOW()
WHERE id = 1;
```

**Delete employee:**

```sql
DELETE FROM employees WHERE id = 1;
```

---

## 📊 Statistics

### Current Database State

```
Total Users: 3
  - Admin: 1
  - Regular: 2

Total Employees: 10+

Storage Estimate:
  - users table: ~5 KB
  - employees table: ~50 KB
  - indexes: ~20 KB
  - Total: ~75 KB
```

---

## 🔧 Maintenance

### Recommended Maintenance Tasks

**Regular (Weekly):**

```sql
-- Vacuum to reclaim space
VACUUM ANALYZE;

-- Check for dead rows
SELECT * FROM pg_stat_user_tables;
```

**Periodic (Monthly):**

```sql
-- Reindex tables
REINDEX TABLE users;
REINDEX TABLE employees;

-- Update statistics
ANALYZE users;
ANALYZE employees;
```

### Backup

```bash
# Full database backup
pg_dump -h localhost -U postgres -d employee_db > employee_db_backup.sql

# Restore from backup
psql -h localhost -U postgres -d employee_db < employee_db_backup.sql
```

---

## 📋 Schema Version Control

| Version | Date       | Changes                                                 |
| ------- | ---------- | ------------------------------------------------------- |
| 1.0     | 2025-12-04 | Initial schema: users and employees tables with indexes |

---

**Schema is production-ready! 🎯**
