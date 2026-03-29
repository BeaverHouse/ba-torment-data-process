package party

import (
	"math"
	"sort"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/logic/id"
	"ba-torment-data-process/internal/types"
)

// GetEssentialCharacters returns characters used by 70%+ of users (excluding assists)
func GetEssentialCharacters(partyData *types.BATormentPartyData) (torment []types.EssentialCharacter, lunatic []types.EssentialCharacter) {
	if len(partyData.PartyDetail) == 0 {
		return nil, nil
	}

	isInsane := partyData.PartyDetail[0].Score < constants.TormentMinScore

	tormentCharCount := make(map[int]int)
	lunaticCharCount := make(map[int]int)
	var tormentUsers, lunaticUsers int

	for _, party := range partyData.PartyDetail {
		if party.Rank > constants.PlatinumRankLimit {
			break
		}

		isLunatic := party.Score >= constants.LunaticMinScore
		isTorment := isInsane || party.Score >= constants.TormentMinScore

		if isLunatic {
			lunaticUsers++
		} else if isTorment {
			tormentUsers++
		} else {
			continue
		}

		for _, members := range party.PartyData {
			for _, member := range members {
				if member == 0 {
					continue
				}
				if member%10 == 1 {
					continue
				}
				studentID := id.GetStudentID(member)
				if isLunatic {
					lunaticCharCount[studentID]++
				} else {
					tormentCharCount[studentID]++
				}
			}
		}
	}

	calcEssential := func(charCount map[int]int, totalUsers int) []types.EssentialCharacter {
		if totalUsers == 0 {
			return nil
		}

		threshold := float64(totalUsers) * 0.7
		type charUsage struct {
			studentID int
			count     int
		}
		var usages []charUsage
		for sid, count := range charCount {
			if float64(count) >= threshold {
				usages = append(usages, charUsage{sid, count})
			}
		}

		sort.Slice(usages, func(i, j int) bool {
			return usages[i].count > usages[j].count
		})

		var result []types.EssentialCharacter
		for _, u := range usages {
			ratio := float64(u.count) / float64(totalUsers)
			result = append(result, types.EssentialCharacter{
				StudentID: u.studentID,
				Ratio:     math.Round(ratio*1000) / 1000,
			})
		}
		return result
	}

	torment = calcEssential(tormentCharCount, tormentUsers)
	lunatic = calcEssential(lunaticCharCount, lunaticUsers)

	return torment, lunatic
}

// GetMinUEUsers returns users who cleared with minimum unique equipment usage
func GetMinUEUsers(partyData *types.BATormentPartyData) (torment *types.MinUEUser, lunatic *types.MinUEUser) {
	if len(partyData.PartyDetail) == 0 {
		return nil, nil
	}

	isInsane := partyData.PartyDetail[0].Score < constants.TormentMinScore

	type userUEData struct {
		rank       int
		score      int
		ueCount    int
		partyCount int
		partyData  [][6]int
	}

	var tormentUsers, lunaticUsers []userUEData

	for _, party := range partyData.PartyDetail {
		if party.Rank > constants.PlatinumRankLimit {
			break
		}

		ueCount := 0
		for _, members := range party.PartyData {
			for _, member := range members {
				if member == 0 {
					continue
				}
				if member%10 == 1 {
					continue
				}
				weaponStar := (member % 100) / 10
				if weaponStar > 0 {
					ueCount++
				}
			}
		}

		userData := userUEData{
			rank:       party.Rank,
			score:      party.Score,
			ueCount:    ueCount,
			partyCount: len(party.PartyData),
			partyData:  party.PartyData,
		}

		if party.Score >= constants.LunaticMinScore {
			lunaticUsers = append(lunaticUsers, userData)
		} else if isInsane || party.Score >= constants.TormentMinScore {
			tormentUsers = append(tormentUsers, userData)
		}
	}

	sortFunc := func(users []userUEData) {
		sort.Slice(users, func(i, j int) bool {
			if users[i].ueCount != users[j].ueCount {
				return users[i].ueCount < users[j].ueCount
			}
			if users[i].partyCount != users[j].partyCount {
				return users[i].partyCount < users[j].partyCount
			}
			return users[i].rank < users[j].rank
		})
	}

	if len(tormentUsers) > 0 {
		sortFunc(tormentUsers)
		torment = &types.MinUEUser{
			Rank:      tormentUsers[0].rank,
			Score:     tormentUsers[0].score,
			UECount:   tormentUsers[0].ueCount,
			PartyData: tormentUsers[0].partyData,
		}
	}

	if len(lunaticUsers) > 0 {
		sortFunc(lunaticUsers)
		lunatic = &types.MinUEUser{
			Rank:      lunaticUsers[0].rank,
			Score:     lunaticUsers[0].score,
			UECount:   lunaticUsers[0].ueCount,
			PartyData: lunaticUsers[0].partyData,
		}
	}

	return torment, lunatic
}

// GetMaxPartyUsers returns users who cleared with maximum party count
func GetMaxPartyUsers(partyData *types.BATormentPartyData) (torment *types.MaxPartyUser, lunatic *types.MaxPartyUser) {
	if len(partyData.PartyDetail) == 0 {
		return nil, nil
	}

	isInsane := partyData.PartyDetail[0].Score < constants.TormentMinScore

	var tormentMaxCount, lunaticMaxCount int

	for _, party := range partyData.PartyDetail {
		if party.Rank > constants.PlatinumRankLimit {
			break
		}

		partyCount := len(party.PartyData)

		if party.Score >= constants.LunaticMinScore {
			if partyCount > lunaticMaxCount {
				lunaticMaxCount = partyCount
				lunatic = &types.MaxPartyUser{
					Rank:      party.Rank,
					Score:     party.Score,
					PartyData: party.PartyData,
				}
			}
		} else if isInsane || party.Score >= constants.TormentMinScore {
			if partyCount > tormentMaxCount {
				tormentMaxCount = partyCount
				torment = &types.MaxPartyUser{
					Rank:      party.Rank,
					Score:     party.Score,
					PartyData: party.PartyData,
				}
			}
		}
	}

	return torment, lunatic
}

// GetHighImpactCharacters returns top 3 characters with the biggest rank gap when missing
func GetHighImpactCharacters(partyData *types.BATormentPartyData) (torment []types.HighImpactCharacter, lunatic []types.HighImpactCharacter) {
	if len(partyData.PartyDetail) == 0 {
		return nil, nil
	}

	isInsane := partyData.PartyDetail[0].Score < constants.TormentMinScore

	type partyInfo struct {
		rank      int
		usedChars map[int]bool
	}

	var tormentParties, lunaticParties []partyInfo

	for _, party := range partyData.PartyDetail {
		if party.Rank > constants.PlatinumRankLimit {
			break
		}

		usedChars := make(map[int]bool)
		for _, members := range party.PartyData {
			for _, member := range members {
				if member == 0 {
					continue
				}
				usedChars[id.GetStudentID(member)] = true
			}
		}

		info := partyInfo{rank: party.Rank, usedChars: usedChars}

		if party.Score >= constants.LunaticMinScore {
			lunaticParties = append(lunaticParties, info)
		} else if isInsane || party.Score >= constants.TormentMinScore {
			tormentParties = append(tormentParties, info)
		}
	}

	findBestRankWithout := func(parties []partyInfo, charID int) int {
		bestRank := 0
		for _, p := range parties {
			if !p.usedChars[charID] {
				if bestRank == 0 || p.rank < bestRank {
					bestRank = p.rank
				}
			}
		}
		return bestRank
	}

	calcHighImpact := func(parties []partyInfo, fallbackParties []partyInfo, top100Limit int) []types.HighImpactCharacter {
		if len(parties) == 0 {
			return nil
		}

		topRank := parties[0].rank

		top100Chars := make(map[int]bool)
		for i, p := range parties {
			if i >= top100Limit {
				break
			}
			for charID := range p.usedChars {
				top100Chars[charID] = true
			}
		}

		type charGap struct {
			studentID       int
			rankGap         int
			withoutBestRank int
		}
		var gaps []charGap

		for charID := range top100Chars {
			withoutBestRank := findBestRankWithout(parties, charID)

			if withoutBestRank == 0 && len(fallbackParties) > 0 {
				withoutBestRank = findBestRankWithout(fallbackParties, charID)
			}

			var rankGap int
			if withoutBestRank > 0 {
				rankGap = withoutBestRank - topRank
			} else {
				rankGap = -1
			}

			gaps = append(gaps, charGap{charID, rankGap, withoutBestRank})
		}

		sort.Slice(gaps, func(i, j int) bool {
			if gaps[i].rankGap == -1 && gaps[j].rankGap != -1 {
				return true
			}
			if gaps[i].rankGap != -1 && gaps[j].rankGap == -1 {
				return false
			}
			return gaps[i].rankGap > gaps[j].rankGap
		})

		var result []types.HighImpactCharacter
		for i := 0; i < 3 && i < len(gaps); i++ {
			result = append(result, types.HighImpactCharacter{
				StudentID:       gaps[i].studentID,
				RankGap:         gaps[i].rankGap,
				TopRank:         topRank,
				WithoutBestRank: gaps[i].withoutBestRank,
			})
		}
		return result
	}

	torment = calcHighImpact(tormentParties, nil, 100)
	lunatic = calcHighImpact(lunaticParties, tormentParties, 100)

	return torment, lunatic
}
