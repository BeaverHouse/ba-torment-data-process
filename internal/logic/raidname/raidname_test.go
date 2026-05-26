package raidname

import "testing"

func TestTranslate(t *testing.T) {
	cases := []struct {
		in       string
		lang     Lang
		expected string
	}{
		// Korean passes through unchanged.
		{"총력전 S88 시가지 카이텐", LangKO, "총력전 S88 시가지 카이텐"},

		// Total Assault (no armor / difficulty parens).
		{"총력전 S88 시가지 카이텐", LangEN, "Total Assault S88 Street KAITEN FX Mk.0"},
		{"총력전 S88 시가지 카이텐", LangZH, "总力战 S88 街区 KAITEN FX 0型"},
		{"총력전 S85 야외 호버크래프트", LangEN, "Total Assault S85 Outdoor Hovercraft"},
		{"총력전 S89 시가지 드럼바르카", LangZH, "总力战 S89 街区 鼓波卡"},

		// Grand Assault with armor + difficulty.
		{"대결전 S33 시가지 쿠로카게 (탄력장갑, 인세인)", LangEN, "Grand Assault S33 Street Kurokage (Elastic Armor, Insane)"},
		{"대결전 S33 시가지 쿠로카게 (탄력장갑, 인세인)", LangZH, "大决战 S33 街区 黑影 (弹力装甲, 疯狂)"},

		// Tight comma (no space after) — historical entries use this form.
		{"대결전 S32 야외 페로로지라 (중장갑,토먼트)", LangEN, "Grand Assault S32 Outdoor Perorodzilla (Heavy Armor, Torment)"},

		// Double-space variant seen in S30 entry.
		{"대결전 S30 시가지 시로쿠로  (탄력장갑,토먼트)", LangEN, "Grand Assault S30 Street Shiro & Kuro (Elastic Armor, Torment)"},

		// Unknown title falls back to the input.
		{"unknown garbage", LangEN, "unknown garbage"},
	}
	for _, c := range cases {
		got := Translate(c.in, c.lang)
		if got != c.expected {
			t.Errorf("Translate(%q, %q):\n  got  %q\n  want %q", c.in, c.lang, got, c.expected)
		}
	}
}
