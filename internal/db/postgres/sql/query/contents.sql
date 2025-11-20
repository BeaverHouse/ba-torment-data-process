-- name: GetContentByID :one
SELECT content_id, start_date FROM batorment_v3.contents WHERE content_id = $1;

-- name: ListContentIDs :many
SELECT content_id FROM batorment_v3.contents WHERE deleted_at IS NULL;
