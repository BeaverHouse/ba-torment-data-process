-- name: InsertPresent :exec
INSERT INTO batorment_v3.presents (present_id, name_ko, rarity, tags, exp_value, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
ON CONFLICT (present_id) DO UPDATE SET 
    name_ko = EXCLUDED.name_ko,
    rarity = EXCLUDED.rarity,
    tags = EXCLUDED.tags,
    exp_value = EXCLUDED.exp_value,
    updated_at = NOW();