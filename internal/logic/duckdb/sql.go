package logic_duckdb

import "fmt"

const duckDBName = "20250813.main"

func GetCompleteRunIDAndScoreSQL(armorType string) string {
	return fmt.Sprintf(`
SELECT crunid, cr.%s_point
FROM complete_runs cr
ORDER BY cr.%s_point DESC
`, armorType, armorType)
}

func GetRunIDsByCompleteRunIDSQL(armorType string, completeRunID int) string {
	return fmt.Sprintf(`
SELECT r.runid
FROM runs_%s r
WHERE r.crunid = %d
`, armorType, completeRunID)
}

func GetPartyInfoByRunIDSQL(armorType string, runID int) string {
	return fmt.Sprintf(`
SELECT sid, build, level, slot, assist
FROM students_%s
WHERE runid = %d
`, armorType, runID)
}
