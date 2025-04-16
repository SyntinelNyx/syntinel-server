-- name: CreateScanEntry :one
INSERT INTO scans (scanner)
VALUES ($1)
RETURNING scan_id;