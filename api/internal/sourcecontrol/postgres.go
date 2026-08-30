package sourcecontrol

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
		return nil, fmt.Errorf("ytv/sourcecontrol: database is required")
	}
	return &PostgresStore{database: database}, nil
}

func (store *PostgresStore) LoadAll(ctx context.Context) ([]Setting, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT source_key, display_name, installed, enabled, home_visible, priority, revision, updated_at
		FROM ytv_source_settings ORDER BY priority, source_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Setting
	for rows.Next() {
		var setting Setting
		if err := rows.Scan(&setting.SourceKey, &setting.DisplayName, &setting.Installed, &setting.Enabled, &setting.HomeVisible, &setting.Priority, &setting.Revision, &setting.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, setting)
	}
	return result, rows.Err()
}

func (store *PostgresStore) Save(ctx context.Context, setting Setting, expectedRevision int64) (Setting, error) {
	var row *sql.Row
	if expectedRevision == 0 {
		row = store.database.QueryRowContext(ctx, `INSERT INTO ytv_source_settings (
			source_key, display_name, installed, enabled, home_visible, priority, revision, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, 1, $7)
		ON CONFLICT (source_key) DO NOTHING
		RETURNING revision, updated_at`, setting.SourceKey, setting.DisplayName, setting.Installed, setting.Enabled, setting.HomeVisible, setting.Priority, setting.UpdatedAt)
	} else {
		row = store.database.QueryRowContext(ctx, `UPDATE ytv_source_settings
		SET display_name=$2, installed=$3, enabled=$4, home_visible=$5, priority=$6, revision=revision+1, updated_at=$7
		WHERE source_key=$1 AND revision=$8
		RETURNING revision, updated_at`, setting.SourceKey, setting.DisplayName, setting.Installed, setting.Enabled, setting.HomeVisible, setting.Priority, setting.UpdatedAt, expectedRevision)
	}
	if err := row.Scan(&setting.Revision, &setting.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Setting{}, ErrRevisionConflict
		}
		return Setting{}, fmt.Errorf("save source setting: %w", err)
	}
	return setting, nil
}
