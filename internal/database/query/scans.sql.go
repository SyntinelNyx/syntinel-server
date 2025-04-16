// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: scans.sql

package query

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createScanEntry = `-- name: CreateScanEntry :one
INSERT INTO scans (scanner)
VALUES ($1)
RETURNING scan_id
`

func (q *Queries) CreateScanEntry(ctx context.Context, scanner pgtype.Text) (pgtype.UUID, error) {
	row := q.db.QueryRow(ctx, createScanEntry, scanner)
	var scan_id pgtype.UUID
	err := row.Scan(&scan_id)
	return scan_id, err
}
