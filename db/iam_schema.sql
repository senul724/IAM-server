-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Drop tables if they exist (for clean reinstall)
DROP TABLE IF EXISTS user_site CASCADE;
DROP TABLE IF EXISTS "user" CASCADE;
DROP TABLE IF EXISTS site CASCADE;

-- Create user table
CREATE TABLE "user" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255),
    photo_url VARCHAR(500)
);

-- Create site table
CREATE TABLE site (
    domain VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    photo_url VARCHAR(500)
);

-- Create user_site junction table (many-to-many relationship)
CREATE TABLE user_site (
    user_id UUID NOT NULL,
    site_domain VARCHAR(255) NOT NULL,
    hashed_password VARCHAR(255),
    last_login BIGINT NOT NULL,
    token VARCHAR(500),
    PRIMARY KEY (user_id, site_domain),
    FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE,
    FOREIGN KEY (site_domain) REFERENCES site(domain) ON DELETE CASCADE
);

-- Create indexes
-- Index on user email (as specified)
CREATE INDEX idx_user_email ON "user"(email);

-- Index on user_site user_id for faster lookups
CREATE INDEX idx_user_site_user_id ON user_site(user_id);

-- Index on user_site site_domain for faster lookups
CREATE INDEX idx_user_site_domain ON user_site(site_domain);

-- Index on last_login for queries filtering by login time
CREATE INDEX idx_user_site_last_login ON user_site(last_login);

-- Index on token for token-based lookups
CREATE INDEX idx_user_site_token ON user_site(token) WHERE token IS NOT NULL;

-- Comments for documentation
COMMENT ON TABLE "user" IS 'User accounts in the IAM system';
COMMENT ON TABLE site IS 'Sites/applications that users can authenticate to';
COMMENT ON TABLE user_site IS 'Many-to-many relationship between users and sites with authentication details';
COMMENT ON COLUMN user_site.last_login IS 'Unix timestamp of last login';
COMMENT ON COLUMN user_site.hashed_password IS 'Bcrypt hashed password for this user-site combination';
COMMENT ON COLUMN user_site.token IS 'Authentication token for this user-site session';
