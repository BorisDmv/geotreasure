-- Idempotent schema. Safe to run on every startup.

CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    username      TEXT        NOT NULL UNIQUE,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS treasures (
    id          BIGSERIAL PRIMARY KEY,
    owner_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    hint        TEXT        NOT NULL DEFAULT '',
    lat         DOUBLE PRECISION NOT NULL,
    lng         DOUBLE PRECISION NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes on lat/lng make the bounding-box pre-filter for "near" searches fast.
CREATE INDEX IF NOT EXISTS idx_treasures_lat ON treasures (lat);
CREATE INDEX IF NOT EXISTS idx_treasures_lng ON treasures (lng);
CREATE INDEX IF NOT EXISTS idx_treasures_owner ON treasures (owner_id);

-- A "find" records that a user located someone's treasure.
CREATE TABLE IF NOT EXISTS finds (
    treasure_id BIGINT      NOT NULL REFERENCES treasures(id) ON DELETE CASCADE,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    found_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (treasure_id, user_id)
);
