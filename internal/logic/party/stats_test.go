package party

import (
	"testing"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/types"
)

func TestSpecialClearUsersPreserveStartingSkillOrder(t *testing.T) {
	wantOrders := [][6]int{{2, 3, 5, 1, 4, 0}}
	data := &types.BATormentPartyData{
		PartyDetail: []types.BATormentPartyDetail{{
			Rank:        1,
			Score:       constants.TormentMinScore,
			PartyData:   [][6]int{{10089520, 10128520, 10148541, 10105520, 20041540, 20039540}},
			SkillOrders: wantOrders,
		}},
	}

	minUE, _ := GetMinUEUsers(data)
	if minUE == nil || len(minUE.SkillOrders) != 1 || minUE.SkillOrders[0] != wantOrders[0] {
		t.Fatalf("minimum UE skill orders = %v, want %v", minUE, wantOrders)
	}
	maxParty, _ := GetMaxPartyUsers(data)
	if maxParty == nil || len(maxParty.SkillOrders) != 1 || maxParty.SkillOrders[0] != wantOrders[0] {
		t.Fatalf("maximum party skill orders = %v, want %v", maxParty, wantOrders)
	}
}
