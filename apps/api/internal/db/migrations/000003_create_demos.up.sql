CREATE TABLE demos (
    id BIGSERIAL PRIMARY KEY,
    s3_bucket TEXT NOT NULL,
    s3_key TEXT NOT NULL,
    team_t_score INTEGER NOT NULL,
    team_ct_score INTEGER NOT NULL,
    winner TEXT,
    players JSONB NOT NULL,
    kills JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX demos_s3_bucket_s3_key_idx ON demos (s3_bucket, s3_key);
