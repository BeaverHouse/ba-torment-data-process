-- name: InsertStudentData :exec
INSERT INTO batorment_v2.students (student_id, name_ko, name_ja, search_keyword, detail, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
ON CONFLICT (student_id) DO UPDATE SET 
    name_ko = EXCLUDED.name_ko,
    name_ja = EXCLUDED.name_ja,
    search_keyword = EXCLUDED.search_keyword,
    detail = EXCLUDED.detail,
    updated_at = NOW();

-- name: UpdateStudentData :exec
UPDATE batorment_v2.students
SET name_ko = $2, name_ja = $3, search_keyword = $4, detail = $5, updated_at = NOW()
WHERE student_id = $1;

-- name: DeleteStudentData :exec
DELETE FROM batorment_v2.students
WHERE student_id = $1;

-- name: ListStudents :many
SELECT student_id, name_ko, name_ja, search_keyword
FROM batorment_v2.students
ORDER BY name_ko;
