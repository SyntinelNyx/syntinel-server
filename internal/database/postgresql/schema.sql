CREATE TYPE VULNSTATE AS ENUM('New', 'Active', 'Resolved', 'Resurfaced');

DO $$ BEGIN CREATE TYPE VULNSTATE AS ENUM('New', 'Active', 'Resolved', 'Resurfaced');
EXCEPTION
WHEN duplicate_object THEN NULL;
END $$;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS root_accounts (
  account_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email VARCHAR(255) NOT NULL UNIQUE,
  username VARCHAR(255) NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  email_verified_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS iam_accounts (
  account_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  root_account_id UUID NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  username VARCHAR(255) NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  account_status VARCHAR(50) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  email_verified_at TIMESTAMPTZ,
  FOREIGN KEY (root_account_id) REFERENCES root_accounts (account_id)
);

CREATE TABLE IF NOT EXISTS permissions (
  permission_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  is_administrator BOOLEAN DEFAULT FALSE,
  view_assets BOOLEAN DEFAULT FALSE,
  manage_assets BOOLEAN DEFAULT FALSE,
  view_modules BOOLEAN DEFAULT FALSE,
  create_modules BOOLEAN DEFAULT FALSE,
  manage_modules BOOLEAN DEFAULT FALSE,
  view_scans BOOLEAN DEFAULT FALSE,
  start_scans BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS roles (
  role_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  role_name VARCHAR(255) NOT NULL UNIQUE,
  is_deleted BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS roles_permissions (
  role_id UUID NOT NULL,
  permission_id UUID NOT NULL,
  PRIMARY KEY (role_id, permission_id),
  FOREIGN KEY (role_id) REFERENCES roles (role_id),
  FOREIGN KEY (permission_id) REFERENCES permissions (permission_id)
);

CREATE TABLE IF NOT EXISTS iam_user_roles (
  iam_account_id UUID NOT NULL,
  role_id UUID NOT NULL,
  PRIMARY KEY (iam_account_id, role_id),
  FOREIGN KEY (iam_account_id) REFERENCES iam_accounts (account_id),
  FOREIGN KEY (role_id) REFERENCES roles (role_id)
);

CREATE TABLE IF NOT EXISTS iam_user_permissions (
  iam_account_id UUID NOT NULL,
  permission_id UUID NOT NULL,
  PRIMARY KEY (iam_account_id, permission_id),
  FOREIGN KEY (iam_account_id) REFERENCES iam_accounts (account_id),
  FOREIGN KEY (permission_id) REFERENCES permissions (permission_id)
);

CREATE TABLE IF NOT EXISTS vulnerability_data (
  vulnerability_data_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  vulnerability_id VARCHAR(50) UNIQUE NOT NULL,
  vulnerability_name VARCHAR(255),
  vulnerability_description TEXT,
  vulnerability_severity VARCHAR(50),
  cvss_score DECIMAL(4, 2),
  created_on TIMESTAMPTZ,
  last_modified TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS vulnerability_state_history (
  history_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  vuln_data_id UUID NOT NULL,
  vulnerability_state VULNSTATE NOT NULL DEFAULT 'New',
  state_changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  root_account_id UUID NOT NULL,
  FOREIGN KEY (vuln_data_id) REFERENCES vulnerability_data (vulnerability_data_id),
  FOREIGN KEY (root_account_id) REFERENCES root_accounts (account_id)
);

-- Convert to hypertable
SELECT create_hypertable(
    'vulnerability_state_history',
    'state_changed_at',
    if_not_exists => TRUE
  );

CREATE TABLE IF NOT EXISTS scans (
  scan_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  root_account_id UUID NOT NULL,
  scanner_name VARCHAR(255) NOT NULL UNIQUE,
  scan_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  FOREIGN KEY (root_account_id) REFERENCES root_accounts (account_id)
);

-- Convert to hypertable
SELECT create_hypertable('scans', 'scan_date', if_not_exists => TRUE);

CREATE TABLE IF NOT EXISTS asset_vulnerability_state (
  scan_result_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  root_account_id UUID NOT NULL,
  scan_id UUID NOT NULL,
  asset_id UUID NOT NULL,
  vulnerability_id UUID NOT NULL,
  scan_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  FOREIGN KEY (root_account_id) REFERENCES root_accounts (account_id),
  FOREIGN KEY (scan_id) REFERENCES scans (scan_id),
  FOREIGN KEY (asset_id) REFERENCES assets (asset_id),
  FOREIGN KEY (vulnerability_id) REFERENCES vulnerability_data (vulnerability_data_id)
);

-- Convert to hypertable
SELECT create_hypertable(
    'asset_vulnerability_state',
    'scan_date',
    if_not_exists => TRUE
  );