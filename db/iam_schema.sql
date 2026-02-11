-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Client table (owners of applications)
CREATE TABLE client (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- App table (applications)
CREATE TABLE app (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner UUID NOT NULL REFERENCES client(id) ON DELETE CASCADE,
    app_id VARCHAR(255) NOT NULL UNIQUE,
    app_secret VARCHAR(255) NOT NULL,
    origin VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Member table (users of applications)
CREATE TABLE member (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    two_factor BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Member password table (passwords of users for specific applications)
CREATE TABLE member_pwd (
    member_id UUID NOT NULL REFERENCES member(id) ON DELETE CASCADE,
    app_id UUID NOT NULL REFERENCES app(id) ON DELETE CASCADE,
    hashed_pwd VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (member_id, app_id)
);

-- Indexes for read-heavy operations

-- Client indexes
CREATE INDEX idx_client_email ON client(email);

-- App indexes
CREATE INDEX idx_app_owner ON app(owner);
CREATE INDEX idx_app_app_id ON app(app_id);

-- Member indexes (read-heavy table)
CREATE INDEX idx_member_email ON member(email);
CREATE INDEX idx_member_email_verified ON member(email_verified);

-- Member_pwd indexes (read-heavy table)
CREATE INDEX idx_member_pwd_member_id ON member_pwd(member_id);
CREATE INDEX idx_member_pwd_app_id ON member_pwd(app_id);

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers to auto-update updated_at on each table
CREATE TRIGGER update_client_updated_at BEFORE UPDATE ON client
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_app_updated_at BEFORE UPDATE ON app
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_member_updated_at BEFORE UPDATE ON member
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_member_pwd_updated_at BEFORE UPDATE ON member_pwd
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
