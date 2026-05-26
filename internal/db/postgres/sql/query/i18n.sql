-- name: UpsertI18n :exec
INSERT INTO batorment_v3.i18n (category, key, name_ko, name_ja, name_en, name_zh, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
ON CONFLICT (category, key) DO UPDATE SET
    name_ko = EXCLUDED.name_ko,
    name_ja = EXCLUDED.name_ja,
    name_en = EXCLUDED.name_en,
    name_zh = EXCLUDED.name_zh,
    updated_at = NOW();

-- name: GetI18nByCategory :many
SELECT * FROM batorment_v3.i18n WHERE category = $1;

-- name: GetI18n :one
SELECT * FROM batorment_v3.i18n WHERE category = $1 AND key = $2;
