package parse

import (
	"fmt"
	"io"
	"strconv"

	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/data"
	"ba-torment-data-process/app/logic"
	"ba-torment-data-process/app/types"
)

// Parses the Google API's CSV data into BA Torment's party data format.
func ParsePartyDataFromGoogleAPI(seasonString string) (*types.BATormentPartyData, error) {

	rankData, err := GetRankData(seasonString)
	if err != nil {
		return nil, common.WrapErrorWithContext("ParsePartyDataFromGoogleAPI", err)
	}

	// 유저 ID와 랭킹 매핑
	userRankMap := make(map[int64]int)
	for _, data := range rankData {
		userRankMap[data.UserID] = data.FinalRank
	}

	reader, err := data.GetPartyCSVFromGoogleAPI(seasonString)
	if err != nil {
		return nil, common.WrapErrorWithContext("ParsePartyDataFromGoogleAPI", err)
	}

	// Skip header
	reader.Read()

	filters := make(map[string][]int)
	assistFilters := make(map[string][]int)
	partyDetail := make([]types.BATormentPartyDetail, 0)

	// 파티 수 계산
	maxParties := 0
	minParties := 100

	// 각 행 처리
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		// 기본 정보 추출 (인덱스 0: Rank, 1: BestRankingPoint)
		rank, _ := strconv.Atoi(record[0])
		score, _ := strconv.ParseInt(record[1], 10, 64)
		userId, _ := strconv.ParseInt(record[2], 10, 64)
		finalRank := userRankMap[userId]

		// 파티 데이터 추출
		partyData := make(map[string][]int)
		partyCount := 0

		// 파티 처리 함수
		processParty := func(startIdx int) []int {
			partyMembers := make([]int, 6)

			// 6명의 캐릭터 처리
			for memberCount := 0; memberCount < 6; memberCount++ {
				// 각 캐릭터의 7개 필드 인덱스 계산
				baseIdx := startIdx + memberCount*7

				// 인덱스가 레코드 길이를 벗어나지 않는지 확인
				if baseIdx+6 >= len(record) {
					break
				}

				// 필요한 정보만 추출 (UniqueId, StarGrade, WeaponGrade, IsAssist)
				uniqueIdIdx := baseIdx + 1
				starGradeIdx := baseIdx + 2
				weaponGradeIdx := baseIdx + 5
				isAssistIdx := baseIdx + 6

				// 캐릭터 정보 추출
				uniqueId, err := strconv.Atoi(record[uniqueIdIdx])
				if err != nil || uniqueId <= 0 {
					continue // 유효하지 않은 캐릭터 ID는 건너뛰기
				}

				starGrade, _ := strconv.Atoi(record[starGradeIdx])
				if starGrade <= 0 {
					continue // 유효하지 않은 성급은 건너뛰기
				}

				weaponGrade, _ := strconv.Atoi(record[weaponGradeIdx])
				if weaponGrade < 0 {
					weaponGrade = 0
				}
				isAssist := record[isAssistIdx] == "True"

				// 캐릭터 ID 생성 (8자리)
				studentDetailID := logic.GetStudentDetailIDInt(uniqueId, starGrade, weaponGrade, isAssist)
				partyMembers[memberCount] = studentDetailID

				logic.UpdatePartyFilters(filters, assistFilters, studentDetailID)
			}

			return partyMembers
		}

		i := 4 // First party's base index

		for i < len(record) {
			partyMembers := processParty(i)

			if len(partyMembers) > 0 {
				partyCount++
				partyData[fmt.Sprintf("party_%d", partyCount)] = partyMembers
			}

			// Next party's base index
			i += 44
		}

		// 파티 수 업데이트
		if partyCount > 0 {
			if partyCount > maxParties {
				maxParties = partyCount
			}
			if partyCount < minParties {
				minParties = partyCount
			}
		}

		level := logic.GetLevelFromScore(int(score))

		// 파티 정보 추가
		partyInfo := types.BATormentPartyDetail{
			FinalRank:   finalRank,
			TormentRank: rank,
			Score:       score,
			UserID:      userId,
			Level:       level,
			PartyData:   partyData,
		}

		// 유저 ID가 있는 경우 추가
		for _, data := range rankData {
			if data.FinalRank == rank && data.Score == score && data.UserID == userId {
				partyInfo.Score = data.PartScore
				partyInfo.Level = logic.GetLevelFromScore(int(data.PartScore))
				break
			}
		}

		partyDetail = append(partyDetail, partyInfo)
	}

	return &types.BATormentPartyData{
		Filters:       filters,
		AssistFilters: assistFilters,
		MinPartys:     minParties,
		MaxPartys:     maxParties,
		PartyDetail:   partyDetail,
	}, nil
}

// Parse Arona AI data into BA Torment website's party data format.
//
// It's a secondary option if Google API is not available.
func ParsePartyDataFromAronaAI(seasonString string) (*types.BATormentPartyData, error) {

	aronaAIData, err := data.GetDataFromAronaAI(seasonString)
	if err != nil {
		return nil, common.WrapErrorWithContext("ParsePartyDataFromAronaAI", err)
	}

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
			Score:       int64(score),
			UserID:      int64(-(idx + 1)),
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
