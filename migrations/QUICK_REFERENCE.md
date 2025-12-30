# Migration Quick Reference

**Fast lookup guide for database migrations**

---

## 📁 File Structure

```
migrations/
├── 001_create_users_table.up.sql       ✅ Users table
├── 001_create_users_table.down.sql     ✅ Rollback users
├── 002_create_employees_table.up.sql   ✅ Employees table
├── 002_create_employees_table.down.sql ✅ Rollback employees
├── 003_create_indexes.up.sql           ✅ All indexes
├── 003_create_indexes.down.sql         ✅ Drop indexes
├── migration_runner.go                 ✅ Go migration runner
├── README.md                           ✅ Full documentation
└── SCHEMA.md                           ✅ Schema details
```

---

## 🎯 Quick Commands

### Run All Migrations (PostgreSQL)

```bash
cd migrations

# Using psql
psql -h localhost -U postgres -d employee_db -f 001_create_users_table.up.sql
psql -h localhost -U postgres -d employee_db -f 002_create_employees_table.up.sql
psql -h localhost -U postgres -d employee_db -f 003_create_indexes.up.sql
```

### Rollback Migrations (Reverse Order)

```bash
psql -h localhost -U postgres -d employee_db -f 003_create_indexes.down.sql
psql -h localhost -U postgres -d employee_db -f 002_create_employees_table.down.sql
psql -h localhost -U postgres -d employee_db -f 001_create_users_table.down.sql
```

### Using Go Migration Runner

```go
// In your application
runner := NewMigrationRunner(db, "./migrations")

// Run all
runner.RunAllMigrations()

// Rollback all
runner.RollbackAllMigrations()

// List all
runner.PrintMigrations()
```

---

## 📋 Tables at a Glance

### users

- 7 columns
- 4 indexes
- Stores: admin credentials, user authentication
- Key fields: username (unique), email (unique), role, password_hash

### employees

- 10 columns
- 6 indexes
- Stores: employee records
- Key fields: email (unique), position, salary, hired_date

---

## 🔍 Verify Migrations

```bash
# Connect to PostgreSQL
psql -h localhost -U postgres -d employee_db

# List tables
\dt

# List indexes
\di

# Describe users table
\d users

# Describe employees table
\d employees

# Row counts
SELECT 'users' as table, COUNT(*) FROM users
UNION ALL
SELECT 'employees' as table, COUNT(*) FROM employees;

# Exit
\q
```

---

## ✅ Checklist After Running Migrations

- [ ] users table exists with 7 columns
- [ ] employees table exists with 10 columns
- [ ] 4 indexes on users table
- [ ] 6 indexes on employees table
- [ ] No errors during migration execution
- [ ] Can query tables successfully

---

## 🚨 Troubleshooting

| Issue                    | Solution                                                  |
| ------------------------ | --------------------------------------------------------- |
| "Permission denied"      | Check PostgreSQL user has create permissions              |
| "Table already exists"   | Drop tables first or use IF NOT EXISTS (already in files) |
| "Column type mismatch"   | Ensure PostgreSQL version supports DECIMAL                |
| "Timestamp format error" | Check PostgreSQL date format settings                     |

---

## 📌 Important Notes

- ✅ All migrations include `.up.sql` and `.down.sql` files
- ✅ All tables use `IF NOT EXISTS` (safe to run multiple times)
- ✅ Timestamps automatically set to CURRENT_TIMESTAMP
- ✅ Primary keys are auto-increment (SERIAL)
- ✅ Email and username fields are unique

---

**Ready to migrate! 🚀**
