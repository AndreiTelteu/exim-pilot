package database

// Schema contains all the SQL statements for creating tables and indexes
const Schema = `
-- Messages table for tracking all messages
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    timestamp DATETIME NOT NULL,
    sender TEXT NOT NULL,
    size INTEGER,
    status TEXT NOT NULL CHECK (status IN ('received', 'queued', 'delivered', 'deferred', 'bounced', 'frozen')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Recipients table for message recipients
CREATE TABLE IF NOT EXISTS recipients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id TEXT NOT NULL,
    recipient TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('delivered', 'deferred', 'bounced', 'pending')),
    delivered_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
);

-- Delivery attempts table
CREATE TABLE IF NOT EXISTS delivery_attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id TEXT NOT NULL,
    recipient TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    host TEXT,
    ip_address TEXT,
    status TEXT NOT NULL CHECK (status IN ('success', 'defer', 'bounce', 'timeout')),
    smtp_code TEXT,
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
);

-- Log entries table for searchable log history
CREATE TABLE IF NOT EXISTS log_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL,
    message_id TEXT,
    log_type TEXT NOT NULL CHECK (log_type IN ('main', 'reject', 'panic')),
    event TEXT NOT NULL,
    host TEXT,
    sender TEXT,
    recipients TEXT, -- JSON array
    size INTEGER,
    status TEXT,
    error_code TEXT,
    error_text TEXT,
    raw_line TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Audit log for administrative actions
CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    action TEXT NOT NULL,
    message_id TEXT,
    user_id TEXT,
    details TEXT, -- JSON
    ip_address TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Queue snapshots for historical tracking
CREATE TABLE IF NOT EXISTS queue_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    total_messages INTEGER NOT NULL DEFAULT 0,
    deferred_messages INTEGER NOT NULL DEFAULT 0,
    frozen_messages INTEGER NOT NULL DEFAULT 0,
    oldest_message_age INTEGER, -- seconds
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);

CREATE INDEX IF NOT EXISTS idx_recipients_message_id ON recipients(message_id);
CREATE INDEX IF NOT EXISTS idx_recipients_status ON recipients(status);
CREATE INDEX IF NOT EXISTS idx_recipients_recipient ON recipients(recipient);

CREATE INDEX IF NOT EXISTS idx_delivery_attempts_message_id ON delivery_attempts(message_id);
CREATE INDEX IF NOT EXISTS idx_delivery_attempts_timestamp ON delivery_attempts(timestamp);
CREATE INDEX IF NOT EXISTS idx_delivery_attempts_recipient ON delivery_attempts(recipient);
CREATE INDEX IF NOT EXISTS idx_delivery_attempts_status ON delivery_attempts(status);

CREATE INDEX IF NOT EXISTS idx_log_entries_timestamp ON log_entries(timestamp);
CREATE INDEX IF NOT EXISTS idx_log_entries_message_id ON log_entries(message_id);
CREATE INDEX IF NOT EXISTS idx_log_entries_event ON log_entries(event);
CREATE INDEX IF NOT EXISTS idx_log_entries_log_type ON log_entries(log_type);
CREATE INDEX IF NOT EXISTS idx_log_entries_sender ON log_entries(sender);

CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_log_message_id ON audit_log(message_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id);

CREATE INDEX IF NOT EXISTS idx_queue_snapshots_timestamp ON queue_snapshots(timestamp);

-- Message notes table for operator notes
CREATE TABLE IF NOT EXISTS message_notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    note TEXT NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
);

-- Message tags table for categorization
CREATE TABLE IF NOT EXISTS message_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    color TEXT, -- hex color code
    user_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    UNIQUE(message_id, tag) -- prevent duplicate tags per message
);

-- Indexes for notes and tags
CREATE INDEX IF NOT EXISTS idx_message_notes_message_id ON message_notes(message_id);
CREATE INDEX IF NOT EXISTS idx_message_notes_user_id ON message_notes(user_id);
CREATE INDEX IF NOT EXISTS idx_message_notes_created_at ON message_notes(created_at);

CREATE INDEX IF NOT EXISTS idx_message_tags_message_id ON message_tags(message_id);
CREATE INDEX IF NOT EXISTS idx_message_tags_tag ON message_tags(tag);
CREATE INDEX IF NOT EXISTS idx_message_tags_user_id ON message_tags(user_id);

-- Users table for authentication
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    email TEXT,
    full_name TEXT,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    last_login_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table for session management
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    expires_at DATETIME NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for authentication tables
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at);
`
