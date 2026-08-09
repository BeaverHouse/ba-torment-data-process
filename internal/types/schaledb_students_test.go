package types

import (
	"encoding/json"
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
