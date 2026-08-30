CREATE TABLE ytv_source_settings (
    source_key TEXT PRIMARY KEY,
    enabled BOOLEAN NOT NULL,
    priority INTEGER NOT NULL CHECK (priority >= 0),
    revision BIGINT NOT NULL DEFAULT 1 CHECK (revision >= 1),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ytv_title_snapshots (
    source_key TEXT NOT NULL,
    external_id TEXT NOT NULL,
    name TEXT NOT NULL,
    year TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (source_key, external_id)
);

CREATE TABLE ytv_playback_progress (
    user_key TEXT NOT NULL,
    source_key TEXT NOT NULL,
    external_id TEXT NOT NULL,
    episode_key TEXT NOT NULL,
    position_seconds INTEGER NOT NULL CHECK (position_seconds >= 0),
    duration_seconds INTEGER NOT NULL CHECK (duration_seconds > 0),
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    revision BIGINT NOT NULL DEFAULT 1 CHECK (revision >= 1),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_key, source_key, external_id, episode_key)
);

CREATE INDEX ytv_playback_progress_user_updated
    ON ytv_playback_progress (user_key, updated_at DESC);

CREATE TABLE ytv_search_events (
    event_id TEXT PRIMARY KEY,
    normalized_query TEXT NOT NULL CHECK (normalized_query <> '' AND length(normalized_query) <= 200),
    result_count INTEGER NOT NULL CHECK (result_count >= 0),
    success_source_count INTEGER NOT NULL CHECK (success_source_count >= 0),
    failed_source_count INTEGER NOT NULL CHECK (failed_source_count >= 0),
    duration_ms BIGINT NOT NULL CHECK (duration_ms >= 0),
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ytv_search_events_occurred
    ON ytv_search_events (occurred_at DESC);

CREATE INDEX ytv_search_events_query_occurred
    ON ytv_search_events (normalized_query, occurred_at DESC);

