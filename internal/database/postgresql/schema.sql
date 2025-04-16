CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS root_accounts (
  account_id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
  email VARCHAR(255) NOT NULL UNIQUE,
  username VARCHAR(255) NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at BIGINT DEFAULT EXTRACT(
    EPOCH
    FROM NOW ()
  ),
  updated_at BIGINT DEFAULT EXTRACT(
    EPOCH
    FROM NOW ()
  ),
  email_verified_at BIGINT
);

CREATE TABLE IF NOT EXISTS iam_accounts (
  account_id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
  root_account_id UUID NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  username VARCHAR(255) NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  account_status VARCHAR(50) NOT NULL,
  created_at BIGINT DEFAULT EXTRACT(
    EPOCH
    FROM NOW ()
  ),
  updated_at BIGINT DEFAULT EXTRACT(
    EPOCH
    FROM NOW ()
  ),
  email_verified_at BIGINT,
  FOREIGN KEY (root_account_id) REFERENCES root_accounts (account_id)
);

CREATE TABLE IF NOT EXISTS permissions (
  permission_id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
  is_administrator BOOLEAN DEFAULT FALSE,
  view_assets BOOLEAN DEFAULT FALSE,
  manage_assets BOOLEAN DEFAULT FALSE,
  view_modules BOOLEAN DEFAULT FALSE,
  create_modules BOOLEAN DEFAULT FALSE,
  manage_modules BOOLEAN DEFAULT FALSE,
  view_scans BOOLEAN DEFAULT FALSE,
  start_scans BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS roles(
  role_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  role_name VARCHAR(255) NOT NULL UNIQUE,
  is_deleted BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS roles_permissions(
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

CREATE TABLE IF NOT EXISTS assets (
  asset_id UUID NOT NULL DEFAULT uuid_generate_v4(),
  asset_name VARCHAR(255) NOT NULL,
  asset_OS VARCHAR(255) NOT NULL,
  PRIMARY KEY (asset_id)
);

CREATE TABLE IF NOT EXISTS vulnerabilities (
  vulnerability_id UUID NOT NULL DEFAULT uuid_generate_v4(),
  cve_id VARCHAR(50) NOT NULL,
  vulnerability_name VARCHAR(255) NOT NULL,
  vulnerability_description TEXT,
  vulnerability_severity VARCHAR(50),
  cvss_score DECIMAL(3, 2),
  reference TEXT [],
  PRIMARY KEY(vulnerability_id)
);

CREATE TABLE IF NOT EXISTS scans (
  scan_id UUID NOT NULL DEFAULT uuid_generate_v4(),
  scan_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  scanner VARCHAR(255),
  PRIMARY KEY(scan_id)
);

CREATE TYPE vuln_state AS ENUM('New', 'Active', 'Resolved', 'Resurfaced');

CREATE TABLE IF NOT EXISTS asset_vulnerability_state (
  scan_result_id UUID NOT NULL DEFAULT uuid_generate_v4(),
  scan_id UUID REFERENCES scans(scan_id),
  asset_id UUID REFERENCES assets(asset_id),
  vulnerability_id UUID REFERENCES vulnerabilities(vulnerability_id),
  vulnerability_state vuln_state,
  PRIMARY KEY (scan_result_id),
  FOREIGN KEY (scan_id) REFERENCES scans (scan_id),
  FOREIGN KEY (asset_id) REFERENCES assets (asset_id),
  FOREIGN KEY (vulnerability_id) REFERENCES vulnerabilities (vulnerability_id)
);