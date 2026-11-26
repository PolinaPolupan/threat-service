package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"threat-service/api"
)

type PostgresThreatRepository struct {
	pool *pgxpool.Pool
}

var _ ThreatRepository = (*PostgresThreatRepository)(nil)

func NewPostgresThreatRepository(
	ctx context.Context,
	databaseURL string,
) (*PostgresThreatRepository, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return &PostgresThreatRepository{
		pool: pool,
	}, nil
}

func (r *PostgresThreatRepository) Close() {
	r.pool.Close()
}

func (r *PostgresThreatRepository) Create(threat api.Threat) error {
	query := `
        INSERT INTO threats (
            md5, name, description, file_name, file_size,
            first_seen, last_seen, severity, sha256, source,
            tags, type
        )
        VALUES (
            $1, $2, $3, $4, $5,
            $6, $7, $8, $9, $10,
            $11, $12
        )
    `

	_, err := r.pool.Exec(
		context.Background(),
		query,
		threat.Md5,
		threat.Name,
		threat.Description,
		threat.FileName,
		threat.FileSize,
		threat.FirstSeen,
		threat.LastSeen,
		threat.Severity,
		threat.Sha256,
		threat.Source,
		threat.Tags,
		threat.Type,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExists
		}

		return err
	}

	return nil
}

func (r *PostgresThreatRepository) Get(md5 string) (api.Threat, error) {
	query := `
        SELECT
            md5, name, description, file_name, file_size,
            first_seen, last_seen, severity, sha256, source,
            tags, type, created_at, updated_at
        FROM threats
        WHERE md5 = $1
    `

	var threat api.Threat

	err := r.pool.QueryRow(context.Background(), query, md5).Scan(
		&threat.Md5,
		&threat.Name,
		&threat.Description,
		&threat.FileName,
		&threat.FileSize,
		&threat.FirstSeen,
		&threat.LastSeen,
		&threat.Severity,
		&threat.Sha256,
		&threat.Source,
		&threat.Tags,
		&threat.Type,
		&threat.CreatedAt,
		&threat.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.Threat{}, ErrNotFound
		}
		return api.Threat{}, err
	}

	return threat, nil
}

func (r *PostgresThreatRepository) Update(md5 string, threat api.Threat) error {
	query := `
        UPDATE threats SET
            name = $2,
            description = $3,
            file_name = $4,
            file_size = $5,
            first_seen = $6,
            last_seen = $7,
            severity = $8,
            sha256 = $9,
            source = $10,
            tags = $11,
            type = $12,
	    updated_at = NOW()
        WHERE md5 = $1
    `

	result, err := r.pool.Exec(
		context.Background(),
		query,
		md5,
		threat.Name,
		threat.Description,
		threat.FileName,
		threat.FileSize,
		threat.FirstSeen,
		threat.LastSeen,
		threat.Severity,
		threat.Sha256,
		threat.Source,
		threat.Tags,
		threat.Type,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
