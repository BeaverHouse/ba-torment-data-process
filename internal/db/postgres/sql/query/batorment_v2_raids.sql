-- name: SoftDeleteOldRaids :execresult
UPDATE batorment_v2.raids 
SET deleted_at = NOW(), updated_at = NOW() 
WHERE created_at < $1 
AND deleted_at IS NULL;