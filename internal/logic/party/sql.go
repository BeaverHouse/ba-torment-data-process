package party

import (
	"database/sql"
	"fmt"
	"strings"
)

func getCompleteRunIDAndScoreSQL(armorType string) string {
	columnName := "point"
	if armorType != "" {
		columnName = armorType + "_point"
	}
	return fmt.Sprintf(`
SELECT crunid, cr.%s
FROM complete_runs cr
ORDER BY cr.%s DESC
`, columnName, columnName)
}

func getRunIDsByCompleteRunIDSQL(armorType string, completeRunID int) string {
	tableName := "runs"
	if armorType != "" {
		tableName = "runs_" + armorType
	}
	return fmt.Sprintf(`
SELECT r.runid
FROM %s r
WHERE r.crunid = %d
ORDER BY r.runid
`, tableName, completeRunID)
}

func getStudentsTableName(armorType string) string {
	if armorType == "" {
		return "students"
	}
	return "students_" + armorType
}

func getPartyInfoByRunIDSQL(armorType string, runID int, hasSkillOrder bool) string {
	mulliganColumn := "0 AS mulligan"
	if hasSkillOrder {
		mulliganColumn = "COALESCE(mulligan, 0) AS mulligan"
	}
	return fmt.Sprintf(`
SELECT sid, build, level, slot, assist, %s
FROM %s
WHERE runid = %d
`, mulliganColumn, getStudentsTableName(armorType), runID)
}

func hasMulliganColumn(db *sql.DB, armorType string) (bool, error) {
	const query = `
SELECT EXISTS (
	SELECT 1
	FROM information_schema.columns
	WHERE table_name = ? AND column_name = 'mulligan'
)`

	var exists bool
	err := db.QueryRow(query, getStudentsTableName(armorType)).Scan(&exists)
	return exists, err
}

func getPlatinumCutSQL(ranks []int) string {
	return fmt.Sprintf(`
WITH ranked AS (
	SELECT point, ROW_NUMBER() OVER (ORDER BY point DESC) as rank
	FROM complete_runs
	WHERE point > 0
)
SELECT rank, point
FROM ranked
WHERE rank IN (%s)
ORDER BY rank
`, intSliceToSQL(ranks))
}

func getPartialPlatinumCutSQL(ranks []int, existingColumns []string) string {
	if len(existingColumns) == 0 {
		return ""
	}

	sumParts := make([]string, len(existingColumns))
	for i, col := range existingColumns {
		sumParts[i] = fmt.Sprintf("COALESCE(%s, 0)", col)
	}
	sumExpr := strings.Join(sumParts, " + ")

	return fmt.Sprintf(`
WITH ranked AS (
	SELECT
		%s as total_point,
		ROW_NUMBER() OVER (ORDER BY %s DESC) as rank
	FROM complete_runs
	WHERE %s > 0
)
SELECT rank, total_point
FROM ranked
WHERE rank IN (%s)
ORDER BY rank
`, sumExpr, sumExpr, sumExpr, intSliceToSQL(ranks))
}

func getCompleteRunsColumnsSQL() string {
	return `SELECT column_name FROM information_schema.columns WHERE table_name = 'complete_runs'`
}

func intSliceToSQL(nums []int) string {
	if len(nums) == 0 {
		return "0"
	}
	result := fmt.Sprintf("%d", nums[0])
	for i := 1; i < len(nums); i++ {
		result += fmt.Sprintf(", %d", nums[i])
	}
	return result
}

func getExistingPointColumns(db *sql.DB) ([]string, error) {
	rows, err := db.Query(getCompleteRunsColumnsSQL())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	armorTypes := []string{"Light_point", "Heavy_point", "Special_point", "Elastic_point"}
	armorTypeSet := make(map[string]bool)
	for _, at := range armorTypes {
		armorTypeSet[at] = true
	}

	var existingColumns []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		if armorTypeSet[columnName] {
			existingColumns = append(existingColumns, columnName)
		}
	}

	return existingColumns, nil
}
