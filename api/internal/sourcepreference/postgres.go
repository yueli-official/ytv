package sourcepreference

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type PostgresStore struct{ database *sql.DB }

func NewPostgresStore(database *sql.DB) (*PostgresStore, error) {
	if database == nil {
		return nil, fmt.Errorf("ytv/sourcepreference: database is required")
	}
	return &PostgresStore{database: database}, nil
}

func (store *PostgresStore) Get(ctx context.Context, userKey string) (Preference, error) {
	var value Preference
	var raw []byte
	err := store.database.QueryRowContext(ctx, `SELECT user_key, disabled_source_keys, revision, updated_at FROM ytv_source_preferences WHERE user_key=$1`, userKey).Scan(&value.UserKey, &raw, &value.Revision, &value.UpdatedAt)
	if err == sql.ErrNoRows {
		return Preference{UserKey: userKey, DisabledSourceKeys: []string{}}, nil
	}
	if err != nil {
		return Preference{}, fmt.Errorf("read source preference: %w", err)
	}
	if err := json.Unmarshal(raw, &value.DisabledSourceKeys); err != nil {
		return Preference{}, fmt.Errorf("decode source preference: %w", err)
	}
	value.Exists = true
	return value, nil
}

func (store *PostgresStore) Save(ctx context.Context, value Preference, expectedRevision int64) (Preference, error) {
	raw, err := json.Marshal(value.DisabledSourceKeys)
	if err != nil {
		return Preference{}, err
	}
	var saved Preference
	var savedRaw []byte
	if expectedRevision == 0 {
		err = store.database.QueryRowContext(ctx, `INSERT INTO ytv_source_preferences (user_key, disabled_source_keys, revision, updated_at)
			VALUES ($1, $2::jsonb, 1, $3) ON CONFLICT (user_key) DO NOTHING
			RETURNING user_key, disabled_source_keys, revision, updated_at`, value.UserKey, raw, value.UpdatedAt).Scan(&saved.UserKey, &savedRaw, &saved.Revision, &saved.UpdatedAt)
	} else {
		err = store.database.QueryRowContext(ctx, `UPDATE ytv_source_preferences
			SET disabled_source_keys=$2::jsonb, revision=revision+1, updated_at=$3
			WHERE user_key=$1 AND revision=$4
			RETURNING user_key, disabled_source_keys, revision, updated_at`, value.UserKey, raw, value.UpdatedAt, expectedRevision).Scan(&saved.UserKey, &savedRaw, &saved.Revision, &saved.UpdatedAt)
	}
	if err == sql.ErrNoRows {
		return Preference{}, ErrRevisionConflict
	}
	if err != nil {
		return Preference{}, fmt.Errorf("save source preference: %w", err)
	}
	if err := json.Unmarshal(savedRaw, &saved.DisabledSourceKeys); err != nil {
		return Preference{}, err
	}
	saved.Exists = true
	return saved, nil
}
