CREATE TABLE IF NOT EXISTS users (
    username VARCHAR(255) PRIMARY KEY,
    password_hash VARCHAR(255) NOT NULL,
    roles TEXT[] NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on username for faster lookups (though PK already handles this)
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Insert default admin user if not exists (password: admin123)
INSERT INTO users (username, password_hash, roles, active)
VALUES ('admin', '$2a$10$LtlhX7.r1Rf9Fl7XjR9VKeaZvwU7PJK6tlWF5rXdxe1fg55wurAnW', ARRAY['admin'], TRUE)
ON CONFLICT (username) DO NOTHING;
