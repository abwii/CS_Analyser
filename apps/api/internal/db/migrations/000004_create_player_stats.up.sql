CREATE TABLE player_stats (
    id BIGSERIAL PRIMARY KEY,
    demo_id BIGINT NOT NULL REFERENCES demos (id) ON DELETE CASCADE,
    steamid BIGINT NOT NULL,
    kpr DOUBLE PRECISION NOT NULL,
    dpr DOUBLE PRECISION NOT NULL,
    kast DOUBLE PRECISION NOT NULL,
    impact DOUBLE PRECISION NOT NULL,
    adr DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX player_stats_demo_id_steamid_idx ON player_stats (demo_id, steamid);
