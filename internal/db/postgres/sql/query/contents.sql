-- name: GetContentByID :one
SELECT content_id, start_date FROM batorment_v3.contents WHERE content_id = $1;

-- name: ListContentIDs :many
SELECT content_id FROM batorment_v3.contents WHERE deleted_at IS NULL;

-- name: ListContentIDsWithStartDate :many
SELECT content_id, start_date FROM batorment_v3.contents WHERE deleted_at IS NULL ORDER BY start_date ASC;

-- name: ListContentsForRaidList :many
SELECT content_id, title, top_level FROM batorment_v3.contents WHERE deleted_at IS NULL ORDER BY start_date ASC;
