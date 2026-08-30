ALTER TABLE ytv_playback_progress
    ADD COLUMN title_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN episode_title TEXT NOT NULL DEFAULT '';
