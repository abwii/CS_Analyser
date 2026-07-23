CREATE TABLE player_positions (
    id BIGSERIAL PRIMARY KEY,
    demo_id BIGINT NOT NULL REFERENCES demos (id) ON DELETE CASCADE,
    steamid BIGINT NOT NULL,
    round INTEGER NOT NULL,
    tick BIGINT NOT NULL,
    team TEXT NOT NULL,
    x DOUBLE PRECISION NOT NULL,
    y DOUBLE PRECISION NOT NULL,
    z DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX player_positions_steamid_idx ON player_positions (steamid);
