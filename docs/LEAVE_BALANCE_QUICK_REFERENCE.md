# Leave Balance Feature - Quick Reference

## Summary of Changes

### Default Leave Balances

| Leave Type | Default Balance |
| ---------- | --------------- |
| Annual     | 10 days         |
| Sick       | 15 days         |
| Casual     | 10 days         |

### How It Works

**1. Balance Checking (On Leave Application)**

- Employee requests leave for ANNUAL/SICK/CASUAL type
- System checks if employee has sufficient balance
- If balance < requested days → Request REJECTED with error
- If balance >= requested days → Request CREATED (PENDING status)

**2. Balance Deduction (On Leave Approval)**

- Admin approves a pending leave request
- System automatically deducts requested days from balance
- Other leave types (PERSONAL, MATERNITY, UNPAID) are not deducted

**3. Balance Viewing**

- Employees can check their balance anytime
- Shows all leave type balances
- Shows balance for specific leave type

### New API Endpoints

| Method | Endpoint                | Description                      | Auth     |
| ------ | ----------------------- | -------------------------------- | -------- |
| GET    | `/leave/balance`        | Get all leave balances           | Employee |
| GET    | `/leave/balance/{type}` | Get balance for specific type    | Employee |
| POST   | `/leave/apply`          | Apply for leave (checks balance) | Employee |
| POST   | `/leave/approve/{id}`   | Approve leave (deducts balance)  | Admin    |

### Correction/Notes

None - implementation is complete and correct. All requirements have been implemented:

✅ Annual leave balance: 10 days
✅ Sick leave balance: 15 days  
✅ Casual leave balance: 10 days
✅ Auto-deduction from balance on approval
✅ Auto-rejection if balance is insufficient
✅ New table: leave_balances (employee_id, leave_type, balance)

### Database Table Structure

```sql
CREATE TABLE leave_balances (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL REFERENCES employees(id),
    leave_type VARCHAR(50) NOT NULL,
    balance INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(employee_id, leave_type)
);
```

### Leave Type Classification

**Managed (Balance Required):**

- ANNUAL
- SICK
- CASUAL

**Unmanaged (No Balance Limits):**

- PERSONAL
- MATERNITY
- UNPAID
