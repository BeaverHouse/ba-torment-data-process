-- setup_test_env: skip. External tables for sqlc validation only.
-- ba-analyzer owns youtube_metadata; seed-test only truncates it and resets its id.
CREATE TABLE batorment_v3.youtube_metadata (
    id SERIAL PRIMARY KEY
);

-- service-api owns service_tokens; seed-test registers this job's token in it.
CREATE SCHEMA IF NOT EXISTS tinyclover;
CREATE TABLE tinyclover.service_tokens (
    name VARCHAR(100) NOT NULL,
    token VARCHAR(60) NOT NULL UNIQUE,
    scopes TEXT[] NOT NULL DEFAULT '{}'
);
