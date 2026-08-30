package progress

import (
	"context"
	"database/sql"
	"fmt"
)

type PostgresStore struct {
	database *sql.DB
}

func NewPostgresStore(database *sql.DB) (*PostgresStore, error) {
	if database == nil {
		return nil, fmt.Errorf("ytv/progress: database is required")
	}
	return &PostgresStore{database: database}, nil
}

func (store *PostgresStore) Save(ctx context.Context, record Record) (SaveResult, error) {
	var saved Record
	err := store.database.QueryRowContext(ctx, `INSERT INTO ytv_playback_progress (
		user_key, source_key, external_id, episode_key, title_name, episode_title,
		position_seconds, duration_seconds, completed, revision, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 1, $10)
	ON CONFLICT (user_key, source_key, external_id, episode_key) DO UPDATE SET
		title_name=EXCLUDED.title_name,
		episode_title=EXCLUDED.episode_title,
		position_seconds=EXCLUDED.position_seconds,
		duration_seconds=EXCLUDED.duration_seconds,
		completed=EXCLUDED.completed,
		revision=ytv_playback_progress.revision+1,
		updated_at=EXCLUDED.updated_at
	WHERE ytv_playback_progress.updated_at < EXCLUDED.updated_at
	RETURNING user_key, source_key, external_id, episode_key, title_name, episode_title,
		position_seconds, duration_seconds, completed, revision, updated_at`,
		record.UserKey, record.SourceKey, record.ExternalID, record.EpisodeKey,
		record.TitleName, record.EpisodeTitle, record.PositionSeconds, record.DurationSeconds,
		record.Completed, record.UpdatedAt).Scan(
		&saved.UserKey, &saved.SourceKey, &saved.ExternalID, &saved.EpisodeKey,
		&saved.TitleName, &saved.EpisodeTitle, &saved.PositionSeconds, &saved.DurationSeconds,
		&saved.Completed, &saved.Revision, &saved.UpdatedAt,
	)
	if err == nil {
		return SaveResult{Record: saved, Applied: true}, nil
	}
	if err != sql.ErrNoRows {
		return SaveResult{}, fmt.Errorf("save playback progress: %w", err)
	}
	err = store.database.QueryRowContext(ctx, `SELECT
		user_key, source_key, external_id, episode_key, title_name, episode_title,
		position_seconds, duration_seconds, completed, revision, updated_at
	FROM ytv_playback_progress
	WHERE user_key=$1 AND source_key=$2 AND external_id=$3 AND episode_key=$4`,
		record.UserKey, record.SourceKey, record.ExternalID, record.EpisodeKey).Scan(
		&saved.UserKey, &saved.SourceKey, &saved.ExternalID, &saved.EpisodeKey,
		&saved.TitleName, &saved.EpisodeTitle, &saved.PositionSeconds, &saved.DurationSeconds,
		&saved.Completed, &saved.Revision, &saved.UpdatedAt,
	)
	if err != nil {
		return SaveResult{}, fmt.Errorf("read retained playback progress: %w", err)
	}
	return SaveResult{Record: saved, Applied: false}, nil
}

func (store *PostgresStore) List(ctx context.Context, userKey string, limit int) ([]Record, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT
		user_key, source_key, external_id, episode_key, title_name, episode_title,
		position_seconds, duration_seconds, completed, revision, updated_at
	FROM ytv_playback_progress
	WHERE user_key=$1
	ORDER BY updated_at DESC, source_key, external_id, episode_key
	LIMIT $2`, userKey, limit)
	if err != nil {
		return nil, fmt.Errorf("list playback progress: %w", err)
	}
	defer rows.Close()
	result := make([]Record, 0)
	for rows.Next() {
		var record Record
		if err := rows.Scan(
			&record.UserKey, &record.SourceKey, &record.ExternalID, &record.EpisodeKey,
			&record.TitleName, &record.EpisodeTitle, &record.PositionSeconds, &record.DurationSeconds,
			&record.Completed, &record.Revision, &record.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan playback progress: %w", err)
		}
		result = append(result, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate playback progress: %w", err)
	}
	return result, nil
}
