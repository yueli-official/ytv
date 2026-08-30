package searchanalytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PostgresStore struct {
	database *sql.DB
}

func NewPostgresStore(database *sql.DB) (*PostgresStore, error) {
	if database == nil {
		return nil, fmt.Errorf("ytv/searchanalytics: database is required")
	}
	return &PostgresStore{database: database}, nil
}

func (store *PostgresStore) Record(ctx context.Context, event Event) (bool, error) {
	result, err := store.database.ExecContext(ctx, `INSERT INTO ytv_search_events (
		event_id, normalized_query, result_count, success_source_count,
		failed_source_count, duration_ms, occurred_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (event_id) DO NOTHING`, event.ID, event.Query, event.ResultCount,
		event.SuccessfulSourceCount, event.FailedSourceCount, event.Duration.Milliseconds(), event.OccurredAt)
	if err != nil {
		return false, fmt.Errorf("record search event: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("inspect search event result: %w", err)
	}
	return affected == 0, nil
}

func (store *PostgresStore) Overview(ctx context.Context, since time.Time, limit int) (Overview, error) {
	var overview Overview
	if err := store.database.QueryRowContext(ctx, `SELECT
		COUNT(*),
		COUNT(*) FILTER (WHERE result_count = 0),
		COUNT(*) FILTER (WHERE failed_source_count > 0),
		COALESCE(ROUND(AVG(duration_ms)), 0)
	FROM ytv_search_events WHERE occurred_at >= $1`, since).Scan(
		&overview.Summary.Searches,
		&overview.Summary.ZeroResultSearches,
		&overview.Summary.IncompleteSearches,
		&overview.Summary.AverageDurationMS,
	); err != nil {
		return Overview{}, fmt.Errorf("summarize search events: %w", err)
	}

	rows, err := store.database.QueryContext(ctx, `SELECT
		normalized_query,
		COUNT(*) AS searches,
		COUNT(*) FILTER (WHERE result_count = 0) AS zero_result_searches,
		COALESCE(AVG(result_count), 0),
		COALESCE(ROUND(AVG(duration_ms)), 0),
		MAX(occurred_at)
	FROM ytv_search_events
	WHERE occurred_at >= $1
	GROUP BY normalized_query
	ORDER BY searches DESC, MAX(occurred_at) DESC, normalized_query
	LIMIT $2`, since, limit)
	if err != nil {
		return Overview{}, fmt.Errorf("read top search queries: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item QueryStat
		if err := rows.Scan(&item.Query, &item.Searches, &item.ZeroResultSearches, &item.AverageResultCount, &item.AverageDurationMS, &item.LastSearchedAt); err != nil {
			return Overview{}, fmt.Errorf("scan top search query: %w", err)
		}
		overview.TopQueries = append(overview.TopQueries, item)
	}
	if err := rows.Err(); err != nil {
		return Overview{}, fmt.Errorf("iterate top search queries: %w", err)
	}

	seriesRows, err := store.database.QueryContext(ctx, `SELECT
		TO_CHAR((occurred_at AT TIME ZONE 'Asia/Shanghai')::date, 'YYYY-MM-DD'),
		COUNT(*),
		COUNT(*) FILTER (WHERE result_count = 0)
	FROM ytv_search_events
	WHERE occurred_at >= $1
	GROUP BY (occurred_at AT TIME ZONE 'Asia/Shanghai')::date
	ORDER BY (occurred_at AT TIME ZONE 'Asia/Shanghai')::date`, since)
	if err != nil {
		return Overview{}, fmt.Errorf("read search series: %w", err)
	}
	defer seriesRows.Close()
	for seriesRows.Next() {
		var point DayStat
		if err := seriesRows.Scan(&point.Day, &point.Searches, &point.ZeroResultSearches); err != nil {
			return Overview{}, fmt.Errorf("scan search series: %w", err)
		}
		overview.Series = append(overview.Series, point)
	}
	if err := seriesRows.Err(); err != nil {
		return Overview{}, fmt.Errorf("iterate search series: %w", err)
	}
	return overview, nil
}
