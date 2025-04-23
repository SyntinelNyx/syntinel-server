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
INSERT INTO asset_vulnerability_scan (root_account_id, scan_id, asset_id, vuln_id)
SELECT s.root_account_id,
    $1 AS scan_id,
    $2 AS asset_id,
    vd.vulnerability_id AS vuln_id
FROM unnest(@vuln_list::text []) AS id
    JOIN vulnerability_data vd ON vd.vulnerability_id = id
    JOIN scans s ON s.scan_id = $1;