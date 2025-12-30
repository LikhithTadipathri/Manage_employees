-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    leave_request_id INTEGER,
    recipient_email VARCHAR(100) NOT NULL,
    recipient_name VARCHAR(100) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    template_name VARCHAR(100) NOT NULL,
    delivery_channel VARCHAR(20) NOT NULL DEFAULT 'SMTP',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    error_message TEXT,
    sent_at TIMESTAMP,
    next_retry_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (leave_request_id) REFERENCES leave_requests(id) ON DELETE SET NULL
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_event_type ON notifications(event_type);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_email ON notifications(recipient_email);
CREATE INDEX IF NOT EXISTS idx_notifications_next_retry_at ON notifications(next_retry_at);
CREATE INDEX IF NOT EXISTS idx_notifications_leave_request_id ON notifications(leave_request_id);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
