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
func ParsePartyDataFromAronaAI(seasonString string) (*types.BATormentPartyData, *types.BATormentFilter, error) {

	var aronaAIData *types.AronaAIData
	fileName := fmt.Sprintf("data/%s.json", seasonString)
	jsonBytes, err := os.ReadFile(fileName)
	if err != nil {
		return nil, nil, common.WrapErrorWithContext("ParsePartyDataFromAronaAI", err)
	}
	json.Unmarshal(jsonBytes, &aronaAIData)

	filters := make(map[string](map[string]int))
	assistFilters := make(map[string](map[string]int))
	var parties []types.BATormentPartyDetail

	minPartys := 99
	maxPartys := 0

	for _, rankData := range aronaAIData.D {
		rank := rankData.R
		score := rankData.S

		partyData := make([][6]int, len(rankData.T))

		for i, party := range rankData.T {
			partyMembers := [6]int{}

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

			partyData[i] = partyMembers
		}

		partyInfo := types.BATormentPartyDetail{
			Rank:      rank,
			Score:     score,
			PartyData: partyData,
		}
		parties = append(parties, partyInfo)

		if len(partyData) < minPartys {
			minPartys = len(partyData)
		}
		if len(partyData) > maxPartys {
			maxPartys = len(partyData)
		}
	}

	filterResult := types.BATormentFilter{
		Filters:       filters,
		AssistFilters: assistFilters,
	}

	// 최종 데이터 구성
	result := types.BATormentPartyData{
		MinPartys:   minPartys,
		MaxPartys:   maxPartys,
		PartyDetail: parties,
	}

	return &result, &filterResult, nil
}
