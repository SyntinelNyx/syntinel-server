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
  is_deleted BOOLEAN DEFAULT FALSE,
  FOREIGN KEY (root_account_id) REFERENCES root_accounts (account_id)
);

CREATE TABLE IF NOT EXISTS system_information (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  hostname TEXT,
  uptime BIGINT,
  boot_time BIGINT,
  procs BIGINT,
  os TEXT,
  platform TEXT,
  platform_family TEXT,
  platform_version TEXT,
  kernel_version TEXT,
  kernel_arch TEXT,
  virtualization_system TEXT,
  virtualization_role TEXT,
  host_id TEXT,
  cpu_vendor_id TEXT,
  cpu_cores INTEGER,
  cpu_model_name TEXT,
  cpu_mhz DOUBLE PRECISION,
  cpu_cache_size INTEGER,
  memory BIGINT,
  disk BIGINT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS assets (
  asset_id UUID PRIMARY KEY,
  ip_address INET NOT NULL,
  sysinfo_id UUID NOT NULL,
  root_account_id UUID NOT NULL,
  registered_at TIMESTAMPTZ DEFAULT NOW(),
  FOREIGN KEY (sysinfo_id) REFERENCES system_information (id),
  FOREIGN KEY (root_account_id) REFERENCES root_accounts (account_id)
);

CREATE TABLE IF NOT EXISTS actions (
  action_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  action_name TEXT NOT NULL,
  action_type VARCHAR(10) NOT NULL,
  action_payload TEXT NOT NULL,
  action_note TEXT NOT NULL,
  root_account_id UUID NOT NULL,
  created_by TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  FOREIGN KEY (root_account_id) REFERENCES root_accounts (account_id)
);

CREATE TABLE IF NOT EXISTS environments (
  environment_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  environment_name TEXT NOT NULL,
  prev_env_id UUID,
  next_env_id UUID,
  root_account_id UUID NOT NULL,
  FOREIGN KEY (root_account_id) REFERENCES root_accounts (account_id),
  FOREIGN KEY (prev_env_id) REFERENCES environments (environment_id),
  FOREIGN KEY (next_env_id) REFERENCES environments (environment_id)
);

CREATE TABLE IF NOT EXISTS environment_assets (
  environment_id UUID NOT NULL,
  asset_id UUID NOT NULL,
  PRIMARY KEY (environment_id, asset_id),
  FOREIGN KEY (environment_id) REFERENCES environments (environment_id),
  FOREIGN KEY (asset_id) REFERENCES assets (asset_id)
);

CREATE TABLE IF NOT EXISTS permission_templates (
  template_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  component_name VARCHAR(50) UNIQUE,
  capabilities TEXT []
);
INSERT INTO permission_templates (component_name, capabilities)
VALUES ('Overview', ARRAY ['View']),
  ('Assets', ARRAY ['View', 'Create', 'Manage']),
  ('Vulnerabilities', ARRAY ['View', 'Manage']),
  (
    'Environments',
    ARRAY ['View', 'Create', 'Manage']
  ),
  ('Actions', ARRAY ['View', 'Create', 'Manage']),
  ('Snapshots', ARRAY ['View', 'Create', 'Manage']),
  ('Scans', ARRAY ['View', 'Create', 'Manage']),
  (
    'UserManagement',
    ARRAY ['View', 'Create', 'Manage']
  ),
  (
    'RoleManagement',
    ARRAY ['View', 'Create', 'Manage']
  ),
  ('ApplicationConfig', ARRAY ['View', 'Manage']),
  ('Logs', ARRAY ['View']) ON CONFLICT (component_name) DO NOTHING;

CREATE TABLE IF NOT EXISTS permissions_new (
  permission_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  permission_name TEXT UNIQUE
);
INSERT INTO permissions_new (permission_name)
SELECT pt.component_name || '.' || cap_name AS permission_name
FROM permission_templates pt
  CROSS JOIN unnest(pt.capabilities) AS cap_name ON CONFLICT (permission_name) DO NOTHING;

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
  reference TEXT [],
  cvss_score DECIMAL(4, 2),
  created_on TIMESTAMPTZ,
  last_modified TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS vulnerability_state_history (
  history_id UUID DEFAULT uuid_generate_v4(),
  vuln_data_id UUID NOT NULL,
  vulnerability_state VULNSTATE NOT NULL,
  state_changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  root_account_id UUID NOT NULL,
  PRIMARY KEY (history_id, state_changed_at),
  FOREIGN KEY (vuln_data_id) REFERENCES vulnerability_data (vulnerability_data_id),
  FOREIGN KEY (root_account_id) REFERENCES root_accounts (account_id)
);

-- Convert to hypertable
SELECT create_hypertable(
    'vulnerability_state_history',
    by_range('state_changed_at'),
    if_not_exists => TRUE,
    migrate_data => TRUE
  );


CREATE TABLE IF NOT EXISTS scans (
  scan_id UUID DEFAULT uuid_generate_v4(),
  root_account_id UUID NOT NULL,
  scanned_by_user UUID,
  scanner_name VARCHAR(255) NOT NULL,
  scan_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  notes TEXT,
  PRIMARY KEY (scan_id),
  FOREIGN KEY (root_account_id) REFERENCES root_accounts (account_id)
);


CREATE TABLE IF NOT EXISTS asset_vulnerability_scan (
  scan_result_id UUID DEFAULT uuid_generate_v4(),
  root_account_id UUID NOT NULL,
  scan_id UUID NOT NULL,
  asset_id UUID NOT NULL,
  vulnerability_id UUID NOT NULL,
  scan_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (scan_result_id, scan_date),
  FOREIGN KEY (root_account_id) REFERENCES root_accounts (account_id),
  FOREIGN KEY (scan_id) REFERENCES scans (scan_id),
  FOREIGN KEY (asset_id) REFERENCES assets (asset_id),
  FOREIGN KEY (vulnerability_id) REFERENCES vulnerability_data (vulnerability_data_id)
);

-- Convert to hypertable
SELECT create_hypertable(
    'asset_vulnerability_scan',
    by_range('scan_date'),
    if_not_exists => TRUE,
    migrate_data => TRUE
  );

CREATE TABLE IF NOT EXISTS telemetry (
  telemetry_id UUID DEFAULT uuid_generate_v4(),
  telemetry_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  cpu_usage FLOAT NOT NULL,
  mem_total BIGINT NOT NULL,
  mem_available BIGINT NOT NULL,
  mem_used BIGINT NOT NULL,
  mem_used_percent FLOAT NOT NULL,
  disk_total BIGINT NOT NULL,
  disk_free BIGINT NOT NULL,
  disk_used BIGINT NOT NULL,
  disk_used_percent FLOAT NOT NULL,
  PRIMARY KEY (telemetry_time, telemetry_id)
);
SELECT create_hypertable(
    'telemetry',
    by_range('telemetry_time'),
    if_not_exists => TRUE,
    migrate_data => TRUE
);

CREATE TABLE IF NOT EXISTS telemetry_asset (
  telemetry_id UUID NOT NULL,
  asset_id UUID NOT NULL,
  root_account_id UUID NOT NULL,
  PRIMARY KEY (telemetry_id, asset_id),
  FOREIGN KEY (asset_id) REFERENCES assets (asset_id),
  FOREIGN KEY (root_account_id) REFERENCES root_accounts (account_id)
);
