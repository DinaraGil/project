package postgres

import (
	"context"
	"errors"
	"fmt"
	"projectOzonBank/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	const schema = `
		CREATE TABLE IF NOT EXISTS links (
		short_code   VARCHAR(10) CONSTRAINT links_short_code_pk PRIMARY KEY,
		original_url TEXT NOT NULL CONSTRAINT links_original_url_uq UNIQUE,
		created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	`
	if _, err := pool.Exec(ctx, schema); err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}
func (s *Storage) Save(ctx context.Context, code, originalURL string) error {
	const query = `
        INSERT INTO links (short_code, original_url)
        VALUES ($1, $2)
    `

	_, err := s.pool.Exec(ctx, query, code, originalURL)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "links_original_url_uq":
			const query = `
                SELECT short_code
                FROM links
                WHERE original_url = $1
            `

			var existingCode string

			err := s.pool.QueryRow(
				ctx,
				query,
				originalURL,
			).Scan(&existingCode)

			if err != nil {
				return fmt.Errorf("get existing code: %w", err)
			}

			return &domain.AlreadyExistsError{
				ExistingCode: existingCode,
			}

		case "links_short_code_pk":
			return domain.ErrCodeAlreadyTaken
		}
	}

	return fmt.Errorf("save link: %w", err)
}
func (s *Storage) Get(ctx context.Context, code string) (string, error) {
	const query = `
        SELECT original_url
        FROM links
        WHERE short_code = $1
    `

	var originalURL string

	err := s.pool.QueryRow(
		ctx,
		query,
		code,
	).Scan(&originalURL)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrNotFound
		}

		return "", fmt.Errorf("get link: %w", err)
	}

	return originalURL, nil
}
