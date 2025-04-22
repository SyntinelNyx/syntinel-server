-- name: CreateScanEntry :one
INSERT INTO scans (scanner)
VALUES ($1)
RETURNING scan_id;

-- name: BatchUpdateAVS :exec
INSERT INTO asset_vulnerability_state (scan_id, asset_id, vuln_id)
SELECT $1 AS scan_id,
    $2 AS asset_id,
    v.id AS vuln_id
FROM unnest(@vuln_list::text []) AS id
    JOIN vulnerabilities v ON v.id = id;