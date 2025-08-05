package parse

import (
	"encoding/json"
	"fmt"
	"os"

	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/logic"
	"ba-torment-data-process/app/types"
)

// Parse Arona AI data into BA Torment website's party data format.
func ParsePartyDataFromAronaAI(seasonString string) (*types.BATormentPartyData, error) {

	var aronaAIData *types.AronaAIData
	fileName := fmt.Sprintf("data/%s.json", seasonString)
	jsonBytes, err := os.ReadFile(fileName)
	if err != nil {
		return nil, common.WrapErrorWithContext("ParsePartyDataFromAronaAI", err)
	}
	json.Unmarshal(jsonBytes, &aronaAIData)

	filters := make(map[string][]int)
	assistFilters := make(map[string][]int)
	var parties []types.BATormentPartyDetail

	for idx, rankData := range aronaAIData.D {
		rank := rankData.R
		score := rankData.S

		partyData := make(map[string][]int)

		for i, party := range rankData.T {
			partyMembers := make([]int, 6)

			for memberIdx := range 6 {
				var char types.AronaAICharacter
				// First 4 students are strikers, and others are supports
				if memberIdx < 4 {
					char = party.M[memberIdx]
				} else {
					char = party.S[memberIdx-4]
				}
				if char.StudentID == 0 {
					continue
				}

				star := char.Star
				weaponStar := 0
				if char.HasWeapon {
					weaponStar = char.WeaponStar
				}

				// 캐릭터 ID 생성 (8자리)
				studentDetailID := logic.GetStudentDetailIDInt(char.StudentID, star, weaponStar, char.IsAssist)
				partyMembers[memberIdx] = studentDetailID

				logic.UpdatePartyFilters(filters, assistFilters, studentDetailID)
			}

			partyData[fmt.Sprintf("party_%d", i+1)] = partyMembers
		}

		level := logic.GetLevelFromScore(score)

		partyInfo := types.BATormentPartyDetail{
			FinalRank:   rank,
			Score:       score,
			UserID:      -(idx + 1),
			Level:       level,
			PartyData:   partyData,
			TormentRank: rank,
		}
		parties = append(parties, partyInfo)
	}

	// 최종 데이터 구성
	result := types.BATormentPartyData{
		Filters:       filters,
		AssistFilters: assistFilters,
		MinPartys:     1,
		MaxPartys:     15,
		PartyDetail:   parties,
	}

	return &result, nil
}
