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

// GetPlatinumCutSQL returns SQL for getting scores at specific ranks for regular raids (총력전)
// ranks should be ordered ascending (e.g., [2000, 4000, 6000, ...])
func GetPlatinumCutSQL(ranks []int) string {
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

// GetEliminationPlatinumCutSQL returns SQL for getting combined scores at specific ranks for elimination raids (대결전)
// It sums all armor type points and ranks by total score
func GetEliminationPlatinumCutSQL(ranks []int) string {
	return fmt.Sprintf(`
WITH ranked AS (
	SELECT
		COALESCE(Light_point, 0) + COALESCE(Heavy_point, 0) +
		COALESCE(Special_point, 0) + COALESCE(Elastic_point, 0) as total_point,
		ROW_NUMBER() OVER (
			ORDER BY COALESCE(Light_point, 0) + COALESCE(Heavy_point, 0) +
			         COALESCE(Special_point, 0) + COALESCE(Elastic_point, 0) DESC
		) as rank
	FROM complete_runs
	WHERE COALESCE(Light_point, 0) + COALESCE(Heavy_point, 0) +
	      COALESCE(Special_point, 0) + COALESCE(Elastic_point, 0) > 0
)
SELECT rank, total_point
FROM ranked
WHERE rank IN (%s)
ORDER BY rank
`, intSliceToSQL(ranks))
}

// intSliceToSQL converts []int to comma-separated string for SQL IN clause
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
