// Package raidname renders raid titles in multiple languages from the
// Korean source string. The dictionaries are hand-maintained — when a new
// raid lands, add its boss to bossMap. Unknown components fall back to the
// Korean text so the output is never empty.
package raidname

import (
	"fmt"
	"regexp"
)

type Lang string

const (
	LangKO Lang = "ko"
	LangEN Lang = "en"
	LangZH Lang = "zh"
)

var raidTypes = map[string]map[Lang]string{
	"총력전": {LangEN: "Total Assault", LangZH: "总力战"},
	"대결전": {LangEN: "Grand Assault", LangZH: "大决战"},
}

var locations = map[string]map[Lang]string{
	"시가지": {LangEN: "Street", LangZH: "街区"},
	"야외":  {LangEN: "Outdoor", LangZH: "户外"},
	"실내":  {LangEN: "Indoor", LangZH: "室内"},
}

// Boss names use the global EN release / common CN community translations.
var bosses = map[string]map[Lang]string{
	"호드":     {LangEN: "Hod", LangZH: "Hod"},
	"비나":     {LangEN: "Binah", LangZH: "Binah"},
	"페로로지라":  {LangEN: "Perorodzilla", LangZH: "佩洛洛斯拉"},
	"헤세드":    {LangEN: "Chesed", LangZH: "Chesed"},
	"게부라":    {LangEN: "Geburah", LangZH: "Geburah"},
	"시로쿠로":   {LangEN: "Shiro & Kuro", LangZH: "白＆黑"},
	"예소드":    {LangEN: "Yesod", LangZH: "Yesod"},
	"예로니무스":  {LangEN: "Hieronymus", LangZH: "希罗尼穆斯"},
	"쿠로카게":   {LangEN: "Kurokage", LangZH: "黑影"},
	"카이텐":    {LangEN: "KAITEN FX Mk.0", LangZH: "KAITEN FX 0型"},
	"호버크래프트": {LangEN: "Hovercraft", LangZH: "灾厄之狐"},
	"고즈":     {LangEN: "Goz", LangZH: "戈兹"},
	"드럼바르카":  {LangEN: "Drum Barca", LangZH: "鼓波卡"},
}

var armors = map[string]map[Lang]string{
	"경장갑":  {LangEN: "Light Armor", LangZH: "轻装甲"},
	"중장갑":  {LangEN: "Heavy Armor", LangZH: "重装甲"},
	"특수장갑": {LangEN: "Special Armor", LangZH: "特殊装甲"},
	"탄력장갑": {LangEN: "Elastic Armor", LangZH: "弹力装甲"},
}

var difficulties = map[string]map[Lang]string{
	"토먼트": {LangEN: "Torment", LangZH: "折磨"},
	"인세인": {LangEN: "Insane", LangZH: "疯狂"},
}

// Matches "총력전 S88 시가지 카이텐" and
// "대결전 S33 시가지 쿠로카게 (탄력장갑, 인세인)". Trailing parens are optional;
// whitespace around the comma is tolerated (some DB entries have it, others
// don't).
var titleRe = regexp.MustCompile(`^(총력전|대결전)\s+S(\d+)\s+(시가지|야외|실내)\s+(\S+?)\s*(?:\(\s*(\S+?)\s*,\s*(\S+?)\s*\))?\s*$`)

// Translate returns the raid title rendered in lang. For LangKO or any
// parsing failure, the original koName is returned unchanged so callers
// always get a non-empty string.
func Translate(koName string, lang Lang) string {
	if lang == LangKO {
		return koName
	}
	m := titleRe.FindStringSubmatch(koName)
	if m == nil {
		return koName
	}
	raidType, season, location, boss, armor, difficulty := m[1], m[2], m[3], m[4], m[5], m[6]

	t, ok := pick(raidTypes, raidType, lang)
	if !ok {
		return koName
	}
	l, _ := pick(locations, location, lang)
	b, _ := pick(bosses, boss, lang)
	if l == "" {
		l = location
	}
	if b == "" {
		b = boss
	}

	base := fmt.Sprintf("%s S%s %s %s", t, season, l, b)
	if armor == "" {
		return base
	}
	a, _ := pick(armors, armor, lang)
	d, _ := pick(difficulties, difficulty, lang)
	if a == "" {
		a = armor
	}
	if d == "" {
		d = difficulty
	}
	return fmt.Sprintf("%s (%s, %s)", base, a, d)
}

func pick(m map[string]map[Lang]string, key string, lang Lang) (string, bool) {
	entry, ok := m[key]
	if !ok {
		return "", false
	}
	v, ok := entry[lang]
	return v, ok
}
