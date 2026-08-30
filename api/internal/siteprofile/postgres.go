package siteprofile

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PostgresStore struct {
	database *sql.DB
}

func NewPostgresStore(database *sql.DB) (*PostgresStore, error) {
	if database == nil {
		return nil, fmt.Errorf("ytv/siteprofile: database is required")
	}
	return &PostgresStore{database: database}, nil
}

func (store *PostgresStore) Get(ctx context.Context, siteKey string) (Profile, bool, error) {
	var profile Profile
	err := store.database.QueryRowContext(ctx, `SELECT site_key, name, description, revision, updated_at
		FROM ytv_site_profiles WHERE site_key=$1`, siteKey).Scan(
		&profile.SiteKey, &profile.Name, &profile.Description, &profile.Revision, &profile.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Profile{}, false, nil
	}
	if err != nil {
		return Profile{}, false, fmt.Errorf("read site profile: %w", err)
	}
	return profile, true, nil
}

func (store *PostgresStore) Save(ctx context.Context, profile Profile, expectedRevision int64) (Profile, error) {
	var row *sql.Row
	if expectedRevision == 0 {
		row = store.database.QueryRowContext(ctx, `INSERT INTO ytv_site_profiles (
			site_key, name, description, revision, updated_at
		) VALUES ($1, $2, $3, 1, $4)
		ON CONFLICT (site_key) DO NOTHING
		RETURNING revision, updated_at`, profile.SiteKey, profile.Name, profile.Description, profile.UpdatedAt)
	} else {
		row = store.database.QueryRowContext(ctx, `UPDATE ytv_site_profiles
		SET name=$2, description=$3, revision=revision+1, updated_at=$4
		WHERE site_key=$1 AND revision=$5
		RETURNING revision, updated_at`, profile.SiteKey, profile.Name, profile.Description, profile.UpdatedAt, expectedRevision)
	}
	if err := row.Scan(&profile.Revision, &profile.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Profile{}, ErrRevisionConflict
		}
		return Profile{}, fmt.Errorf("save site profile: %w", err)
	}
	return profile, nil
}
