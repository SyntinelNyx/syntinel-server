package response

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

func UuidToString(u pgtype.UUID) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}

func UuidToStringPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := UuidToString(u)
	return &s
}
