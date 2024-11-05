package repositories

import (
	"context"
	modelentities "url-shortening-service/models/entities"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ShortenRepository interface {
	Create(tx pgx.Tx, ctx context.Context, shorten modelentities.Shorten) (lastInsertedId int, err error)
	GetByShortCode(pool *pgxpool.Pool, ctx context.Context, shortCode string) (shorten modelentities.Shorten, err error)
	GetByShortCodeTx(tx pgx.Tx, ctx context.Context, shortCode string) (shorten modelentities.Shorten, err error)
	UpdateUrl(tx pgx.Tx, ctx context.Context, shorten modelentities.Shorten) (rowsAffected int64, err error)
	DeleteById(tx pgx.Tx, ctx context.Context, id int) (rowsAffected int64, err error)
	UpdateAccessCount(tx pgx.Tx, ctx context.Context, accessCount int, id int) (rowsAffected int64, err error)
}

type ShortenRepositoryImplementation struct {
}

func NewShortenRepository() ShortenRepository {
	return &ShortenRepositoryImplementation{}
}

func (repository *ShortenRepositoryImplementation) Create(tx pgx.Tx, ctx context.Context, shorten modelentities.Shorten) (lastInsertedId int, err error) {
	query := `INSERT INTO shortens(url, short_code, created_at, updated_at, access_count) VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	err = tx.QueryRow(ctx, query, shorten.Url, shorten.ShortCode, shorten.CreatedAt, shorten.UpdatedAt, shorten.AccessCount).Scan(&lastInsertedId)
	return
}

func (repository *ShortenRepositoryImplementation) GetByShortCode(pool *pgxpool.Pool, ctx context.Context, shortCode string) (shorten modelentities.Shorten, err error) {
	query := `SELECT id, url, short_code, created_at, updated_at, access_count 
			FROM shortens 
			WHERE short_code = $1;`
	err = pool.QueryRow(ctx, query, shortCode).Scan(&shorten.Id, &shorten.Url, &shorten.ShortCode, &shorten.CreatedAt, &shorten.UpdatedAt, &shorten.AccessCount)
	return
}

func (repository *ShortenRepositoryImplementation) GetByShortCodeTx(tx pgx.Tx, ctx context.Context, shortCode string) (shorten modelentities.Shorten, err error) {
	query := `SELECT id, url, short_code, created_at, updated_at, access_count
			FROM shortens
			WHERE short_code = $1;`
	err = tx.QueryRow(ctx, query, shortCode).Scan(&shorten.Id, &shorten.Url, &shorten.ShortCode, &shorten.CreatedAt, &shorten.UpdatedAt, &shorten.AccessCount)
	return
}

func (repository *ShortenRepositoryImplementation) UpdateUrl(tx pgx.Tx, ctx context.Context, shorten modelentities.Shorten) (rowsAffected int64, err error) {
	query := `UPDATE shortens SET url = $1, updated_at = $2 WHERE id = $3;`
	result, err := tx.Exec(ctx, query, shorten.Url, shorten.UpdatedAt, shorten.Id)
	if err != nil {
		return
	}
	rowsAffected = result.RowsAffected()
	return
}

func (repository *ShortenRepositoryImplementation) DeleteById(tx pgx.Tx, ctx context.Context, id int) (rowsAffected int64, err error) {
	query := `DELETE FROM shortens WHERE id = $1;`
	result, err := tx.Exec(ctx, query, id)
	if err != nil {
		return
	}
	rowsAffected = result.RowsAffected()
	return
}

func (repository *ShortenRepositoryImplementation) UpdateAccessCount(tx pgx.Tx, ctx context.Context, accessCount int, id int) (rowsAffected int64, err error) {
	query := `UPDATE shortens SET access_count = $1 WHERE id = $2;`
	result, err := tx.Exec(ctx, query, accessCount, id)
	if err != nil {
		return
	}
	rowsAffected = result.RowsAffected()
	return
}
