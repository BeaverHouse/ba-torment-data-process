-- name: GetContents :one
SELECT content_id, start_date FROM batorment_v3.contents WHERE content_id = $1;
