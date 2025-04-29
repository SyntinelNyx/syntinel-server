-- name: CreateScanEntryRoot :one
INSERT INTO scans (scanner_name, root_account_id)
VALUES ($1, $2)
RETURNING scan_id;

-- name: CreateScanEntryIAMUser :one
INSERT INTO scans (scanner_name, root_account_id, scanned_by_user)
VALUES (
        $1,
        (
            SELECT root_account_id
            FROM iam_accounts
            WHERE account_id = $2
        ),
        $2
    )
RETURNING scan_id;

-- name: BatchUpdateAVS :exec
INSERT INTO asset_vulnerability_scan (
        root_account_id,
        scan_id,
        asset_id,
        vulnerability_id
    )
SELECT s.root_account_id,
    $1 AS scan_id,
    $2 AS asset_id,
    vd.vulnerability_data_id AS vulnerability_id
FROM unnest(@vuln_list::text []) AS id
    JOIN vulnerability_data vd ON vd.vulnerability_id = id
    JOIN scans s ON s.scan_id = $1;

-- name: RetrieveScans :many
SELECT s.scan_id,
    ra.username AS root_account_username,
    s.root_account_id,
    s.scanner_name,
    s.scan_date,
    s.notes
FROM scans s
    JOIN root_accounts ra ON s.root_account_id = ra.account_id
WHERE root_account_id = $1
ORDER BY s.scan_date DESC;

-- name: UpdateScanNotes :exec
UPDATE scans
SET notes = $1
WHERE scan_id = $2;