CREATE TABLE IF NOT EXISTS monitors (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT        NOT NULL,
    url             TEXT        NOT NULL,
    interval_s      INTEGER     NOT NULL DEFAULT 60,
    status          TEXT        NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_checked_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_monitors_active
    ON monitors(status) WHERE status = 'active';

CREATE TABLE IF NOT EXISTS checks (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id  UUID        NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    checked_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status_code INTEGER     NOT NULL,
    response_ms INTEGER     NOT NULL,
    error       TEXT,
    CONSTRAINT fk_monitor FOREIGN KEY (monitor_id) REFERENCES monitors(id)
);

CREATE INDEX IF NOT EXISTS idx_checks_monitor_id
    ON checks(monitor_id);

CREATE INDEX IF NOT EXISTS idx_checks_checked_at
    ON checks(monitor_id, checked_at DESC);