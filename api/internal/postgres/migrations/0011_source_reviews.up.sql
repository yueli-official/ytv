CREATE TABLE ytv_source_reviews (
 user_key TEXT NOT NULL,
 source_key TEXT NOT NULL,
 ads TEXT NOT NULL DEFAULT '' CHECK (ads IN ('','few','many')),
 quality TEXT NOT NULL DEFAULT '' CHECK (quality IN ('','hd','normal','blurry')),
 updated_at TIMESTAMPTZ NOT NULL,
 PRIMARY KEY (user_key,source_key),
 CHECK (ads <> '' OR quality <> '')
);
CREATE INDEX ytv_source_reviews_source ON ytv_source_reviews(source_key);
