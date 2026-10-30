package types

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestStudentDataUnmarshalsFractionalSkillEffectValue(t *testing.T) {
	const data = `{
		"Skills": {
			"Ex": {
				"Desc": "",
				"Effects": [{
					"Type": "Buff",
					"Value": [[3197, 3676], [3836.3999999999996, 4411.2]]
				}]
			}
		}
	}`

	var student StudentData
	if err := json.Unmarshal([]byte(data), &student); err != nil {
		t.Fatalf("unmarshal student data: %v", err)
	}

	got := student.Skills.Ex.Effects[0].Value[1][0]
	if got != 3836.3999999999996 {
		t.Fatalf("fractional effect value = %v, want 3836.3999999999996", got)
	}
}

// Raw-source calculator tests cannot detect metadata discarded while serializing a student's DB detail.
func TestStudentDataPreservesCalculationMetadata(t *testing.T) {
	const skills = `{
		"Ex": {
			"Id": "setup", "Type": "Ex", "Desc": "",
			"Effects": [{"Type": "Special", "Key": "duel", "Channel": 90111}],
			"GroupLabel": {"1": {"Icon": "stack.webp", "Label": "x2", "LabelTranslated": "setting_on", "Disabled": true}},
			"ExtraSkills": [{
				"Id": "attack", "Name": "Attack", "Desc": "",
				"GroupLabel": {"1": {"Label": "x5"}},
				"Effects": [{
					"Type": "Damage", "Scale": [10000], "Hits": [10000], "Group": 1,
					"Condition": {"Type": "Special", "Parameter": "duel", "Operand": "Exists", "Value": false},
					"CanEvade": false, "ApplyStability": false,
					"IgnoreDef": [5000], "HitFrames": [0, 10],
					"StackLabel": [{"Label": "x2"}], "StackSame": 2, "OverrideSlot": "Ex"
				}]
			}]
		},
		"Public": {"Desc": "", "Effects": [{"Type": "Damage", "Scale": [76009], "SourceStat": "DefensePower"}]}
	}`
	var student StudentData
	if err := json.Unmarshal([]byte(`{"Skills":`+skills+`}`), &student); err != nil {
		t.Fatal(err)
	}
	stored, err := json.Marshal(student)
	if err != nil {
		t.Fatal(err)
	}
	var want map[string]any
	if err := json.Unmarshal([]byte(skills), &want); err != nil {
		t.Fatal(err)
	}
	var got struct{ Skills map[string]any }
	if err := json.Unmarshal(stored, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Skills, want) {
		t.Fatalf("calculation metadata changed in stored detail: %s", stored)
	}
}
