CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE
  IF NOT EXISTS root_accounts (
    account_id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at BIGINT DEFAULT EXTRACT(
      EPOCH
      FROM
        NOW ()
    ),
    updated_at BIGINT DEFAULT EXTRACT(
      EPOCH
      FROM
        NOW ()
    ),
    email_verified_at BIGINT
  );

CREATE TABLE
  IF NOT EXISTS iam_accounts (
    account_id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    root_account_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    account_status VARCHAR(50) NOT NULL,
    created_at BIGINT DEFAULT EXTRACT(
      EPOCH
      FROM
        NOW ()
    ),
    updated_at BIGINT DEFAULT EXTRACT(
      EPOCH
      FROM
        NOW ()
    ),
    email_verified_at BIGINT,
    FOREIGN KEY (root_account_id) REFERENCES root_accounts (account_id)
  );

CREATE TABLE
  IF NOT EXISTS permissions (
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

CREATE TABLE
  IF NOT EXISTS iam_account_permissions (
    iam_account_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    PRIMARY KEY (iam_account_id, permission_id),
    FOREIGN KEY (iam_account_id) REFERENCES iam_accounts (account_id),
    FOREIGN KEY (permission_id) REFERENCES permissions (permission_id)
  );