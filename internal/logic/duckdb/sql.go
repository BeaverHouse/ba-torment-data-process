package logic_duckdb

import "fmt"

func GetCompleteRunIDAndScoreSQL(armorType string) string {
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

func GetRunIDsByCompleteRunIDSQL(armorType string, completeRunID int) string {
	tableName := "runs"
	if armorType != "" {
		tableName = "runs_" + armorType
	}
	return fmt.Sprintf(`
SELECT r.runid
FROM %s r
WHERE r.crunid = %d
`, tableName, completeRunID)
}

func GetPartyInfoByRunIDSQL(armorType string, runID int) string {
	tableName := "students"
	if armorType != "" {
		tableName = "students_" + armorType
	}
	return fmt.Sprintf(`
SELECT sid, build, level, slot, assist
FROM %s
WHERE runid = %d
`, tableName, runID)
}
