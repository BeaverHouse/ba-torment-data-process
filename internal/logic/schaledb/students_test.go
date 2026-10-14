package schaledb

import "testing"

// Without this, a form tag that has no localization row (<s:FormChange='잠수'>)
// reaches the reader as raw markup, as it did in 코코로's skill text.
func TestReplaceBuffTagsResolvesFormTag(t *testing.T) {
	buffNames := map[string]string{"Buff_ATK": "공격력"}
	got := replaceBuffTags("<s:FormChange='잠수'> 상태에서 <b:ATK> 증가", buffNames)
	want := "잠수 상태에서 공격력 증가"
	if got != want {
		t.Fatalf("replaceBuffTags = %q, want %q", got, want)
	}
	unknown := replaceBuffTags("<s:Nothing> 유지", buffNames)
	if unknown != "<s:Nothing> 유지" {
		t.Fatalf("unknown tag must stay visible, got %q", unknown)
	}
}
