-- name: TruncateSeedTables :exec
TRUNCATE batorment_v3.contents, batorment_v3.youtube_metadata, batorment_v3.youtube_analysis;

-- name: RegisterServiceToken :exec
-- A fresh test database has no token for this job, and the file manager
-- refuses its uploads without one.
INSERT INTO tinyclover.service_tokens (name, token, scopes)
VALUES ('e2e-batorment', $1, '{batorment,file}')
ON CONFLICT (token) DO UPDATE SET scopes = EXCLUDED.scopes;

-- name: ResetVideoIDs :exec
-- Restored ids keep their values, so the next insert continues after them.
SELECT
    setval(pg_get_serial_sequence('batorment_v3.youtube_metadata', 'id'),
        COALESCE((SELECT MAX(id) FROM batorment_v3.youtube_metadata), 0) + 1, false),
    setval(pg_get_serial_sequence('batorment_v3.youtube_analysis', 'id'),
        COALESCE((SELECT MAX(id) FROM batorment_v3.youtube_analysis), 0) + 1, false);
