CREATE TABLE ytv_source_preferences (
    user_key TEXT PRIMARY KEY,
    disabled_source_keys JSONB NOT NULL DEFAULT '[]'::jsonb,
    revision BIGINT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

