package schaledb

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/storage"
	"ba-torment-data-process/internal/types"

	"github.com/BeaverHouse/go-common/errorhandle"
	"github.com/BeaverHouse/go-common/logger"
)

// SchaleDB zh students.json ships empty SearchTags for every student, so
// Chinese community nicknames never reach searchKeywords the way kr/jp tags
// do. zhAliasesRaw is a hand-curated studentID -> nicknames map that fills
// that gap. Skin shorthand that decomposes by rule (水X = swimsuit X) is left
// to the LLM prompt; this file holds only meaning-based nicknames (社长, 魔王…).
//
//go:embed zh_aliases.json
var zhAliasesRaw []byte

func loadZhAliases(log logger.Logger) map[string][]string {
	var m map[string][]string
	if err := json.Unmarshal(zhAliasesRaw, &m); err != nil {
		log.Warn("Failed to parse zh aliases", logger.Field{Key: "error", Value: err})
		return map[string][]string{}
	}
	return m
}

// Localization data structure
type LocalizationRawData struct {
	BuffName map[string]string `json:"BuffName"`
}

// LocaleNamed carries only the localized display name of a nested object.
type LocaleNamed struct {
	Name string `json:"Name"`
}

// LocaleSkill carries the localized skill name plus selectable-EX sub names.
type LocaleSkill struct {
	Name        string        `json:"Name"`
	ExtraSkills []LocaleNamed `json:"ExtraSkills"`
}

type JapaneseStudentInfo struct {
	Name       string                 `json:"Name"`
	FamilyName string                 `json:"FamilyName"`
	SearchTags []string               `json:"SearchTags"`
	Skills     map[string]LocaleSkill `json:"Skills"`
	Weapon     LocaleNamed            `json:"Weapon"`
	Gear       LocaleNamed            `json:"Gear"`
}

// loadStudentNames returns map[studentID] = {Name, SearchTags} for any
// SchaleDB locale (kr, jp, en, zh). The shape matches JapaneseStudentInfo.
func loadStudentNames(log logger.Logger, lang string) (map[string]JapaneseStudentInfo, error) {
	url := fmt.Sprintf("%s/data/%s/students.json", constants.SchaleDBURL, lang)
	byteValue, err := storage.GetDataFromURL(log, url)
	if err != nil {
		return nil, constants.ErrDataFetch(fmt.Sprintf("students (%s)", lang), err)
	}
	var data map[string]JapaneseStudentInfo
	if err := json.Unmarshal(byteValue, &data); err != nil {
		return nil, constants.ErrDataDecode(fmt.Sprintf("students (%s)", lang), err)
	}
	return data, nil
}

// Reads /data/<lang>/localization.min.json file and returns BuffName map
func loadLocalization(log logger.Logger, lang string) (map[string]string, error) {
	url := fmt.Sprintf("%s/data/%s/localization.min.json", constants.SchaleDBURL, lang)
	byteValue, err := storage.GetDataFromURL(log, url)
	if err != nil {
		return nil, constants.ErrDataFetch(fmt.Sprintf("localization (%s)", lang), err)
	}

	var locData LocalizationRawData
	if err := json.Unmarshal(byteValue, &locData); err != nil {
		return nil, constants.ErrDataDecode(fmt.Sprintf("localization (%s)", lang), err)
	}

	return locData.BuffName, nil
}

// Reads /data/jp/students.json file and returns Japanese student info map.
// This is used to get student names in Japanese.
func loadJapaneseStudentInfo(log logger.Logger) (map[string]JapaneseStudentInfo, error) {
	byteValue, err := storage.GetDataFromURL(log, constants.SchaleDBURL+"/data/jp/students.json")
	if err != nil {
		return nil, constants.ErrDataFetch("japanese student info", err)
	}

	var studentData map[string]JapaneseStudentInfo
	if err := json.Unmarshal(byteValue, &studentData); err != nil {
		return nil, constants.ErrDataDecode("japanese student info", err)
	}

	return studentData, nil
}

// replaceTaggedName resolves a single <x:TAG> occurrence by trying each
// prefix against the localization BuffName map. Returns the original match
// when nothing resolves, so an unknown tag is visible rather than silently
// dropped.
func replaceTaggedName(desc, marker string, prefixes []string, buffNames map[string]string) string {
	re := regexp.MustCompile(`<` + marker + `:([^>]+)>`)
	return re.ReplaceAllStringFunc(desc, func(match string) string {
		tag := strings.TrimPrefix(strings.TrimSuffix(match, ">"), "<"+marker+":")
		for _, prefix := range prefixes {
			if value, exists := buffNames[prefix+tag]; exists {
				return value
			}
		}
		// A form tag such as <s:FormChange='잠수'> names the form inline and
		// has no localization row; the quoted name is the text.
		if _, quoted, found := strings.Cut(tag, "='"); found {
			return strings.TrimSuffix(quoted, "'")
		}
		return match
	})
}

// replaceBuffTags resolves the buff/debuff/special-stack tags SchaleDB leaves
// in skill descriptions into their localized proper nouns.
func replaceBuffTags(desc string, buffNames map[string]string) string {
	// <s:TAG> marks per-character special stacks (e.g. 카요코(새해)의 부적).
	// Without this the raw tag reaches the LLM and the proper noun is lost.
	desc = replaceTaggedName(desc, "s", []string{"Special_", "Buff_", "CC_"}, buffNames)
	desc = replaceTaggedName(desc, "b", []string{"Buff_", "Special_", "CC_"}, buffNames)
	desc = replaceTaggedName(desc, "d", []string{"Debuff_", "Special_", "CC_"}, buffNames)

	return desc
}

// convertStringToStringSlice converts string to []string
func convertStringToStringSlice(field any) any {
	if str, ok := field.(string); ok {
		return []string{str}
	}
	return field
}

// isIntArrayOfArrays checks if value is of type [][]int
func isIntArrayOfArrays(value any) bool {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice {
		return false
	}

	for i := 0; i < v.Len(); i++ {
		innerValue := v.Index(i).Interface()
		innerV := reflect.ValueOf(innerValue)

		if innerV.Kind() != reflect.Slice {
			return false
		}

		for j := 0; j < innerV.Len(); j++ {
			element := innerV.Index(j).Interface()
			if _, ok := element.(float64); !ok {
				return false
			}
		}
	}
	return true
}

// processEffectField processes Effect fields
func processEffectField(effectMap map[string]any, fieldName string, processor func(any) any) {
	if field, exists := effectMap[fieldName]; exists {
		effectMap[fieldName] = processor(field)
	}
}

// processValueField processes Value field
func processValueField(effectMap map[string]any) {
	if valueField, exists := effectMap["Value"]; exists {
		if !isIntArrayOfArrays(valueField) {
			effectMap["AdditionalValue"] = valueField
			delete(effectMap, "Value")
		}
	}
}

// processEffectsSlice processes Effects array in a skill map
func processEffectsSlice(skillMap map[string]any, fieldName string) {
	if effects, ok := skillMap[fieldName]; ok {
		if effectsSlice, ok := effects.([]any); ok {
			for _, effect := range effectsSlice {
				if effectMap, ok := effect.(map[string]any); ok {
					// Convert Target string to []string if needed
					processEffectField(effectMap, "Target", convertStringToStringSlice)
					// Process Value field
					processValueField(effectMap)
				}
			}
		}
	}
}

// processSkillDesc processes Desc and Parameters in a skill map
func processSkillDesc(skillMap map[string]any, buffNames map[string]string) {
	if desc, descExists := skillMap["Desc"]; descExists {
		if parameters, paramExists := skillMap["Parameters"]; paramExists {
			if descStr, ok := desc.(string); ok {
				// Handle parameters as []any
				if paramSlice, ok := parameters.([]any); ok {
					// Replace <?1>, <?2>, etc. with last values from Parameters
					for i, param := range paramSlice {
						if innerSlice, ok := param.([]any); ok && len(innerSlice) > 0 {
							// Get the last value from the inner slice
							lastValue := innerSlice[len(innerSlice)-1]
							if lastValueStr, ok := lastValue.(string); ok {
								// Handle both regular and unicode encoded placeholders
								placeholder1 := fmt.Sprintf("<?%d>", i+1)
								placeholder2 := fmt.Sprintf("\\u003c?%d\\u003e", i+1)
								descStr = strings.ReplaceAll(descStr, placeholder1, lastValueStr)
								descStr = strings.ReplaceAll(descStr, placeholder2, lastValueStr)
							}
						}
					}

					// Replace buff/debuff tags
					descStr = replaceBuffTags(descStr, buffNames)
					// A skill whose text lives in its forms (ExtraSkills) has
					// an empty Desc; the level marker alone would read as text.
					if strings.TrimSpace(descStr) == "" {
						skillMap["Desc"] = ""
					} else {
						skillMap["Desc"] = descStr + " (Skill Lv.Max)"
					}
				}
			}
		}
	}
}

// skillSlots maps SchaleDB skill keys to their typed slots so locale names
// can be injected by key.
func skillSlots(s *types.StudentSkills) map[string]*types.StudentSkill {
	return map[string]*types.StudentSkill{
		"Normal":        s.Normal,
		"Ex":            s.Ex,
		"Public":        s.Public,
		"GearPublic":    s.GearPublic,
		"Passive":       s.Passive,
		"WeaponPassive": s.WeaponPassive,
		"ExtraPassive":  s.ExtraPassive,
	}
}

// injectLocaleNames fills NameJa/NameEn/NameZh on skills, weapon, and gear
// from the per-locale SchaleDB datasets. Missing locales leave fields empty —
// consumers fall back to the Korean Name.
func injectLocaleNames(data *types.StudentData, ja, en, zh JapaneseStudentInfo) {
	data.FamilyNameJa = ja.FamilyName
	data.FamilyNameEn = en.FamilyName
	data.FamilyNameZh = zh.FamilyName
	for key, skill := range skillSlots(&data.Skills) {
		if skill == nil {
			continue
		}
		skill.NameJa = ja.Skills[key].Name
		skill.NameEn = en.Skills[key].Name
		skill.NameZh = zh.Skills[key].Name
		for i := range skill.ExtraSkills {
			if i < len(ja.Skills[key].ExtraSkills) {
				skill.ExtraSkills[i].NameJa = ja.Skills[key].ExtraSkills[i].Name
			}
			if i < len(en.Skills[key].ExtraSkills) {
				skill.ExtraSkills[i].NameEn = en.Skills[key].ExtraSkills[i].Name
			}
			if i < len(zh.Skills[key].ExtraSkills) {
				skill.ExtraSkills[i].NameZh = zh.Skills[key].ExtraSkills[i].Name
			}
		}
	}
	data.Weapon.NameJa = ja.Weapon.Name
	data.Weapon.NameEn = en.Weapon.Name
	data.Weapon.NameZh = zh.Weapon.Name
	data.Gear.NameJa = ja.Gear.Name
	data.Gear.NameEn = en.Gear.Name
	data.Gear.NameZh = zh.Gear.Name
}

// The students' names and search tags in the languages beyond Korean.
type studentLocales struct {
	ja, en, zh map[string]JapaneseStudentInfo
	zhAliases  map[string][]string
}

// The Japanese names are required; English and Chinese are non-fatal if
// upstream 404s — that locale is skipped and the FE falls back to ko.
func loadStudentLocales(log logger.Logger) (studentLocales, error) {
	ja, err := loadJapaneseStudentInfo(log)
	if err != nil {
		return studentLocales{}, err
	}
	en, err := loadStudentNames(log, "en")
	if err != nil {
		log.Warn("Failed to load English student names", logger.Field{Key: "error", Value: err})
		en = map[string]JapaneseStudentInfo{}
	}
	zh, err := loadStudentNames(log, "zh")
	if err != nil {
		log.Warn("Failed to load Chinese student names", logger.Field{Key: "error", Value: err})
		zh = map[string]JapaneseStudentInfo{}
	}
	return studentLocales{ja: ja, en: en, zh: zh, zhAliases: loadZhAliases(log)}, nil
}

// tags merges the search tags of all four locales so LLM keyword search
// resolves nicknames and abbreviations regardless of user language.
func (l studentLocales) tags(studentID string, koTags []string) []string {
	tags := append([]string{}, koTags...)
	tags = append(tags, l.ja[studentID].SearchTags...)
	tags = append(tags, l.en[studentID].SearchTags...)
	tags = append(tags, l.zh[studentID].SearchTags...)
	return append(tags, l.zhAliases[studentID]...)
}

// Resolves the effects and descriptions of every skill, selectable EX skills included.
func processSkills(dataMap map[string]any, buffNames map[string]string) {
	skillsMap, _ := dataMap["Skills"].(map[string]any)
	for _, skill := range skillsMap {
		skillMap, ok := skill.(map[string]any)
		if !ok {
			continue
		}
		processEffectsSlice(skillMap, "Effects")
		processSkillDesc(skillMap, buffNames)

		extraSkills, _ := skillMap["ExtraSkills"].([]any)
		for _, extraSkill := range extraSkills {
			if extraSkillMap, ok := extraSkill.(map[string]any); ok {
				processEffectsSlice(extraSkillMap, "Effects")
				processSkillDesc(extraSkillMap, buffNames)
			}
		}
	}
}

// The row stored for one student: its processed SchaleDB data with every
// locale's names. A student whose data cannot be converted is skipped.
func studentRow(studentID string, dataMap map[string]any, locales studentLocales) (postgres.InsertStudentDataParams, error) {
	resource := "student " + studentID
	processedJSON, err := json.Marshal(dataMap)
	if err != nil {
		return postgres.InsertStudentDataParams{}, constants.ErrDataEncode(resource, err)
	}
	var completeData types.StudentData
	if err := json.Unmarshal(processedJSON, &completeData); err != nil {
		return postgres.InsertStudentDataParams{}, constants.ErrDataDecode(resource, err)
	}
	injectLocaleNames(&completeData, locales.ja[studentID], locales.en[studentID], locales.zh[studentID])
	id, err := strconv.ParseInt(studentID, 10, 32)
	if err != nil {
		return postgres.InsertStudentDataParams{}, constants.ErrDataDecode(resource, err)
	}
	detail, err := json.Marshal(completeData)
	if err != nil {
		return postgres.InsertStudentDataParams{}, constants.ErrDataEncode(resource, err)
	}
	return postgres.InsertStudentDataParams{
		StudentID:     int32(id),
		NameKo:        completeData.Name,
		NameJa:        locales.ja[studentID].Name,
		NameEn:        locales.en[studentID].Name,
		NameZh:        locales.zh[studentID].Name,
		SearchKeyword: locales.tags(studentID, completeData.SearchTags),
		Detail:        detail,
	}, nil
}

// ParseSchaleDBStudents stores every SchaleDB student with its portrait and
// uploads the student maps.
func ParseSchaleDBStudents(log logger.Logger, db *postgres.Queries) error {
	byteValue, err := storage.GetDataFromURL(log, constants.SchaleDBURL+"/data/kr/students.json")
	if err != nil {
		return constants.ErrDataFetch("students", err)
	}

	var rawData map[string]any
	if err := json.Unmarshal(byteValue, &rawData); err != nil {
		return constants.ErrDataDecode("students", err)
	}

	buffNames, err := loadLocalization(log, "kr")
	if err != nil {
		return err
	}
	locales, err := loadStudentLocales(log)
	if err != nil {
		return err
	}

	studentMap := make(map[string]string)

	for studentID, studentData := range rawData {
		dataMap, ok := studentData.(map[string]any)
		if !ok {
			log.Warn("Skipping invalid student data", logger.Field{Key: "studentID", Value: studentID})
			continue
		}
		processSkills(dataMap, buffNames)

		row, err := studentRow(studentID, dataMap, locales)
		if err != nil {
			log.Warn("Skipping student", logger.Field{Key: "studentID", Value: studentID}, logger.Field{Key: "error", Value: err})
			continue
		}
		if err := db.InsertStudentData(context.Background(), row); err != nil {
			return errorhandle.ErrDBOperation("insert student", err)
		}

		studentMap[studentID] = row.NameKo

		// Upload image + wait 3 second. Supabase S3 has performance issue when uploading too many files at once.
		err = uploadCharacterImage(log, int(row.StudentID), false)
		if err != nil {
			log.Warn("Failed to upload image for student", logger.Field{Key: "studentID", Value: studentID}, logger.Field{Key: "error", Value: err})
			return err
		}
		log.Info("Student processed", logger.Field{Key: "studentID", Value: studentID}, logger.Field{Key: "name", Value: row.NameKo})
	}

	if err := storage.MarshalAndUpload(log, studentMap, "batorment/v3", "student-map.json", false, ""); err != nil {
		log.Warn("Failed to upload student map", logger.Field{Key: "error", Value: err})
		return err
	}

	if err := storage.MarshalAndUpload(log, searchMap(rawData, locales), "batorment/v3", "student-search-map.json", false, ""); err != nil {
		log.Warn("Failed to upload student search map", logger.Field{Key: "error", Value: err})
		return err
	}

	return nil
}

// The student search map: every student's names in four languages with the
// search keywords of all of them.
func searchMap(rawData map[string]any, locales studentLocales) map[string]map[string]any {
	studentSearchMap := make(map[string]map[string]any)
	for studentID, studentData := range rawData {
		dataMap, ok := studentData.(map[string]any)
		if !ok {
			continue
		}
		nameKo, exists := dataMap["Name"]
		if !exists {
			continue
		}
		var koTags []string
		tags, _ := dataMap["SearchTags"].([]any)
		for _, tag := range tags {
			if s, ok := tag.(string); ok {
				koTags = append(koTags, s)
			}
		}
		studentSearchMap[studentID] = map[string]any{
			"nameKo":         nameKo,
			"nameJa":         locales.ja[studentID].Name,
			"nameEn":         locales.en[studentID].Name,
			"nameZh":         locales.zh[studentID].Name,
			"searchKeywords": locales.tags(studentID, koTags),
		}
	}
	return studentSearchMap
}

// Uploads the character image from SchaleDB via File Manager API.
func uploadCharacterImage(log logger.Logger, id int, dryRun bool) error {
	imgBytes, err := storage.GetDataFromURL(log, constants.SchaleDBURL+"images/student/icon/"+strconv.Itoa(id)+".webp")
	if err != nil {
		return constants.ErrDataFetch(fmt.Sprintf("character image %d", id), err)
	}

	path := "batorment/character"

	if err := storage.UploadFile(log, path, strconv.Itoa(id)+".webp", imgBytes, dryRun); err != nil {
		return err
	}

	return nil
}
