CREATE DATABASE everylog;
\c everylog

-- Create table for locations first as it's referenced in other tables
CREATE TABLE IF NOT EXISTS location (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    address1 VARCHAR(255),
    address2 VARCHAR(255),
    city VARCHAR(255),
    state VARCHAR(255),
    country VARCHAR(255),
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    CONSTRAINT location_unique UNIQUE (address1, address2, city, state, country)
);

-- Create table for single_users
CREATE TABLE IF NOT EXISTS single_user (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    pii_id UUID,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    token TEXT
);


-- Create table for organizations
CREATE TABLE IF NOT EXISTS org (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    owner UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    location_id UUID,
    CONSTRAINT org_unique_for_user UNIQUE (owner, name),
    FOREIGN KEY (owner) REFERENCES single_user(id),
    FOREIGN KEY (location_id) REFERENCES location(id)
);

CREATE TABLE IF NOT EXISTS user_org_level (
    id INT PRIMARY KEY NOT NULL,
    value VARCHAR(255)
);

-- Create table for single_user-organization relationships (many-to-many)
CREATE TABLE IF NOT EXISTS user_org (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    user_id UUID,
    org_id UUID,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    level INT,
    FOREIGN KEY (user_id) REFERENCES single_user(id),
    FOREIGN KEY (org_id) REFERENCES org(id),
    FOREIGN KEY (level) REFERENCES user_org_level(id),
    CONSTRAINT user_org_unique UNIQUE (user_id, org_id)
);

-- Create table for single_user PII
CREATE TABLE IF NOT EXISTS user_pii (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    user_id UUID,
    email VARCHAR(255) UNIQUE,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    location_id UUID,
    mobile_number VARCHAR(20),
    password VARCHAR(255), -- todo: hash this or use a better method
    FOREIGN KEY (user_id) REFERENCES single_user(id),
    FOREIGN KEY (location_id) REFERENCES location(id)
);

-- Create table for projects
CREATE TABLE IF NOT EXISTS project (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    user_id UUID,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    description TEXT,
    FOREIGN KEY (user_id) REFERENCES single_user(id),
    CONSTRAINT project_unique UNIQUE (name, user_id)
);

-- Create table for project-organization relationships (many-to-many)
CREATE TABLE IF NOT EXISTS project_org (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    org_id UUID,
    project_id UUID,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES org(id),
    FOREIGN KEY (project_id) REFERENCES project(id)
);

-- Create table for permitted projects (many-to-many)
-- Must have a user_id and a project_id
-- Server should ensure user_id has permission to be added to project_id
-- This should be via an invite
CREATE TABLE IF NOT EXISTS permitted_project (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    project_id UUID NOT NULL,
    user_id UUID NOT NULL,
    FOREIGN KEY (user_id) REFERENCES single_user(id),
    FOREIGN KEY (project_id) REFERENCES project(id)
);

-- Create table for invites
CREATE TABLE IF NOT EXISTS project_invite (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    from_user_id UUID NOT NULL,
    to_user_id UUID NOT NULL,
    status VARCHAR(10) NOT NULL DEFAULT 'PENDING',
    project_id UUID NOT NULL,
    FOREIGN KEY (from_user_id) REFERENCES single_user(id),
    FOREIGN KEY (to_user_id) REFERENCES single_user(id),
    FOREIGN KEY (project_id) REFERENCES project(id),
    CONSTRAINT project_invite_unique UNIQUE (from_user_id, to_user_id, project_id)
);

-- Create table for invites
CREATE TABLE IF NOT EXISTS org_invite (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    from_user_id UUID NOT NULL,
    to_user_id UUID NOT NULL,
    status VARCHAR(10) NOT NULL DEFAULT 'PENDING',
    org_id UUID NOT NULL,
    FOREIGN KEY (from_user_id) REFERENCES single_user(id),
    FOREIGN KEY (to_user_id) REFERENCES single_user(id),
    FOREIGN KEY (org_id) REFERENCES org(id)
);


-- Create table for API keys
CREATE TABLE IF NOT EXISTS api_key (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    permitted_project_id UUID UNIQUE NOT NULL,
    key TEXT NOT NULL,
    FOREIGN KEY (permitted_project_id) REFERENCES permitted_project(id)
);

-- Create table for processes
CREATE TABLE IF NOT EXISTS process (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    project_id UUID,
    name VARCHAR(255),
    FOREIGN KEY (project_id) REFERENCES project(id),
    CONSTRAINT process_unique UNIQUE (name, project_id)
);

-- Create table for log levels
CREATE TABLE IF NOT EXISTS log_level (
    id INT PRIMARY KEY NOT NULL,
    value VARCHAR(255)
);


-- Create table for logs
CREATE TABLE IF NOT EXISTS log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL,
    project_id UUID NOT NULL,
    level_id INT NOT NULL,
    process_id UUID,
    message TEXT,
    traceback TEXT,
    FOREIGN KEY (user_id) REFERENCES single_user(id),
    FOREIGN KEY (project_id) REFERENCES project(id),
    FOREIGN KEY (level_id) REFERENCES log_level(id),
    FOREIGN KEY (process_id) REFERENCES process(id)
);

