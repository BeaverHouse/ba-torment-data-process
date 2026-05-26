-- name: InsertStudentData :exec
INSERT INTO batorment_v3.students (student_id, name_ko, name_ja, name_en, name_zh, search_keyword, detail, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
ON CONFLICT (student_id) DO UPDATE SET
    name_ko = EXCLUDED.name_ko,
    name_ja = EXCLUDED.name_ja,
    name_en = EXCLUDED.name_en,
    name_zh = EXCLUDED.name_zh,
    search_keyword = EXCLUDED.search_keyword,
    detail = EXCLUDED.detail,
    updated_at = NOW();