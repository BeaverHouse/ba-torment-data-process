-- setup_test_env: skip. External tables for sqlc validation only.
-- service-api owns service_tokens; seed-test registers this job's token in it.
CREATE SCHEMA IF NOT EXISTS tinyclover;
CREATE TABLE tinyclover.service_tokens (
    name VARCHAR(100) NOT NULL,
    token VARCHAR(60) NOT NULL UNIQUE,
    scopes TEXT[] NOT NULL DEFAULT '{}'
);
