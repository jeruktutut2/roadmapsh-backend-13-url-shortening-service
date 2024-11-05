package modelentities

import "github.com/jackc/pgx/v5/pgtype"

type Shorten struct {
	Id          pgtype.Int4
	Url         pgtype.Text
	ShortCode   pgtype.Text
	CreatedAt   pgtype.Int8
	UpdatedAt   pgtype.Int8
	AccessCount pgtype.Int4
}
