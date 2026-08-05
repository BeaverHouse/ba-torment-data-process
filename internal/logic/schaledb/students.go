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
					skillMap["Desc"] = descStr + " (Skill Lv.Max)"
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

// injectLocaleNames keeps the Japanese dataset as the structural source while
// taking every display name from its exact locale dataset.
func injectLocaleNames(data *types.StudentData, ko, ja, en, zh JapaneseStudentInfo) {
	data.Name = ko.Name
	for key, skill := range skillSlots(&data.Skills) {
		if skill == nil {
			continue
		}
		skill.Name = ko.Skills[key].Name
		skill.NameJa = ja.Skills[key].Name
		skill.NameEn = en.Skills[key].Name
		skill.NameZh = zh.Skills[key].Name
		for i := range skill.ExtraSkills {
			skill.ExtraSkills[i].Name = ko.Skills[key].ExtraSkills[i].Name
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
	data.Weapon.Name = ko.Weapon.Name
	data.Weapon.NameJa = ja.Weapon.Name
	data.Weapon.NameEn = en.Weapon.Name
	data.Weapon.NameZh = zh.Weapon.Name
	data.Gear.Name = ko.Gear.Name
	data.Gear.NameJa = ja.Gear.Name
	data.Gear.NameEn = en.Gear.Name
	data.Gear.NameZh = zh.Gear.Name
}

func validateStudentLocales(roster map[string]any, locales map[string]map[string]JapaneseStudentInfo) error {
	for lang, students := range locales {
		if len(students) != len(roster) {
			return constants.ErrStudentLocalesInvalid(fmt.Sprintf("student locale %s has %d entries, want %d", lang, len(students), len(roster)))
		}
		for studentID := range roster {
			student, exists := students[studentID]
			if !exists || strings.TrimSpace(student.Name) == "" {
				return constants.ErrStudentLocalesInvalid(fmt.Sprintf("student locale %s is missing name for %s", lang, studentID))
			}

			japanese := locales["jp"][studentID]
			for key, japaneseSkill := range japanese.Skills {
				localizedSkill := student.Skills[key]
				if japaneseSkill.Name != "" && strings.TrimSpace(localizedSkill.Name) == "" {
					return constants.ErrStudentLocalesInvalid(fmt.Sprintf("student locale %s is missing skill name for %s/%s", lang, studentID, key))
				}
				if len(localizedSkill.ExtraSkills) != len(japaneseSkill.ExtraSkills) {
					return constants.ErrStudentLocalesInvalid(fmt.Sprintf("student locale %s has mismatched extra skills for %s/%s", lang, studentID, key))
				}
				for i, extraSkill := range japaneseSkill.ExtraSkills {
					if extraSkill.Name != "" && strings.TrimSpace(localizedSkill.ExtraSkills[i].Name) == "" {
						return constants.ErrStudentLocalesInvalid(fmt.Sprintf("student locale %s is missing extra skill name for %s/%s/%d", lang, studentID, key, i))
					}
				}
			}
			if japanese.Weapon.Name != "" && strings.TrimSpace(student.Weapon.Name) == "" {
				return constants.ErrStudentLocalesInvalid(fmt.Sprintf("student locale %s is missing weapon name for %s", lang, studentID))
			}
			if japanese.Gear.Name != "" && strings.TrimSpace(student.Gear.Name) == "" {
				return constants.ErrStudentLocalesInvalid(fmt.Sprintf("student locale %s is missing gear name for %s", lang, studentID))
			}
		}
	}
	return nil
}

// ParseSchaleDBStudents parses SchaleDB data and returns processed student data
func ParseSchaleDBStudents(log logger.Logger, db *postgres.Queries) (map[string]*types.StudentData, error) {
	japaneseBytes, err := storage.GetDataFromURL(log, constants.SchaleDBURL+"/data/jp/students.json")
	if err != nil {
		return nil, constants.ErrDataFetch("japanese students", err)
	}
	var japaneseRawData map[string]any
	if err := json.Unmarshal(japaneseBytes, &japaneseRawData); err != nil {
		return nil, constants.ErrDataDecode("japanese students", err)
	}

	rawData := japaneseRawData

	buffNames, err := loadLocalization(log, "kr")
	if err != nil {
		return nil, err
	}
	koreanStudentInfo, err := loadStudentNames(log, "kr")
	if err != nil {
		return nil, err
	}
	japaneseStudentInfo, err := loadStudentNames(log, "jp")
	if err != nil {
		return nil, err
	}
	englishStudentInfo, err := loadStudentNames(log, "en")
	if err != nil {
		return nil, err
	}
	chineseStudentInfo, err := loadStudentNames(log, "zh")
	if err != nil {
		return nil, err
	}
	if err := validateStudentLocales(rawData, map[string]map[string]JapaneseStudentInfo{
		"kr": koreanStudentInfo,
		"jp": japaneseStudentInfo,
		"en": englishStudentInfo,
		"zh": chineseStudentInfo,
	}); err != nil {
		return nil, err
	}

	zhAliases := loadZhAliases(log)

	result := make(map[string]*types.StudentData)
	studentMap := make(map[string]string)

	for studentID, studentData := range rawData {
		dataMap, ok := studentData.(map[string]any)
		if !ok {
			log.Warn("Skipping invalid student data", logger.Field{Key: "studentID", Value: studentID})
			continue
		}

		// Process Skills data
		if skills, ok := dataMap["Skills"]; ok {
			if skillsMap, ok := skills.(map[string]any); ok {
				for _, skill := range skillsMap {
					if skillMap, ok := skill.(map[string]any); ok {
						// Process Effects
						processEffectsSlice(skillMap, "Effects")

						// Process Desc and Parameters
						processSkillDesc(skillMap, buffNames)

						// Process ExtraSkills (for selectable EX skills)
						if extraSkills, ok := skillMap["ExtraSkills"]; ok {
							if extraSkillsSlice, ok := extraSkills.([]any); ok {
								for _, extraSkill := range extraSkillsSlice {
									if extraSkillMap, ok := extraSkill.(map[string]any); ok {
										// Process Effects in ExtraSkill
										processEffectsSlice(extraSkillMap, "Effects")
										// Process Desc and Parameters in ExtraSkill
										processSkillDesc(extraSkillMap, buffNames)
									}
								}
							}
						}
					}
				}
			}
		}

		// Convert processed map back to CompleteStudentData struct
		processedJSON, err := json.Marshal(dataMap)
		if err != nil {
			log.Warn("Failed to marshal student data", logger.Field{Key: "studentID", Value: studentID}, logger.Field{Key: "error", Value: err})
			continue
		}

		var completeData types.StudentData
		err = json.Unmarshal(processedJSON, &completeData)
		if err != nil {
			log.Warn("Failed to unmarshal student data", logger.Field{Key: "studentID", Value: studentID}, logger.Field{Key: "error", Value: err})
			continue
		}
		injectLocaleNames(&completeData, koreanStudentInfo[studentID], japaneseStudentInfo[studentID], englishStudentInfo[studentID], chineseStudentInfo[studentID])
		studentIDInt64, err := strconv.ParseInt(studentID, 10, 32)
		if err != nil {
			log.Warn("Failed to convert student ID to int", logger.Field{Key: "studentID", Value: studentID}, logger.Field{Key: "error", Value: err})
			continue
		}
		completeDataBytes, err := json.Marshal(completeData)
		if err != nil {
			log.Warn("Failed to marshal student data", logger.Field{Key: "studentID", Value: studentID}, logger.Field{Key: "error", Value: err})
			continue
		}

		// Merge search tags across all four locales so LLM keyword search
		// resolves nicknames and abbreviations regardless of user language.
		mergedTags := append([]string{}, koreanStudentInfo[studentID].SearchTags...)
		mergedTags = append(mergedTags, japaneseStudentInfo[studentID].SearchTags...)
		mergedTags = append(mergedTags, englishStudentInfo[studentID].SearchTags...)
		mergedTags = append(mergedTags, chineseStudentInfo[studentID].SearchTags...)
		mergedTags = append(mergedTags, zhAliases[studentID]...)

		db.InsertStudentData(context.Background(), postgres.InsertStudentDataParams{
			StudentID:     int32(studentIDInt64),
			NameKo:        koreanStudentInfo[studentID].Name,
			NameJa:        japaneseStudentInfo[studentID].Name,
			NameEn:        englishStudentInfo[studentID].Name,
			NameZh:        chineseStudentInfo[studentID].Name,
			SearchKeyword: mergedTags,
			Detail:        completeDataBytes,
		})

		studentMap[studentID] = completeData.Name

		// Upload image + wait 3 second. Supabase S3 has performance issue when uploading too many files at once.
		err = uploadCharacterImage(log, int(studentIDInt64), false)
		if err != nil {
			log.Warn("Failed to upload image for student", logger.Field{Key: "studentID", Value: studentID}, logger.Field{Key: "error", Value: err})
			return nil, err
		}
		log.Info("Student processed", logger.Field{Key: "studentID", Value: studentID}, logger.Field{Key: "name", Value: completeData.Name})
	}

	if err := storage.MarshalAndUpload(log, studentMap, "batorment/v3", "student-map.json", false, ""); err != nil {
		log.Warn("Failed to upload student map", logger.Field{Key: "error", Value: err})
		return nil, err
	}

	// Create student search map with multi-language names and merged search keywords.
	studentSearchMap := make(map[string]map[string]any)
	for studentID, studentData := range rawData {
		_, ok := studentData.(map[string]any)
		if !ok {
			continue
		}
		nameKo := koreanStudentInfo[studentID].Name
		koTags := koreanStudentInfo[studentID].SearchTags
		searchKeywords := append([]string{}, koTags...)
		searchKeywords = append(searchKeywords, japaneseStudentInfo[studentID].SearchTags...)
		searchKeywords = append(searchKeywords, englishStudentInfo[studentID].SearchTags...)
		searchKeywords = append(searchKeywords, chineseStudentInfo[studentID].SearchTags...)
		searchKeywords = append(searchKeywords, zhAliases[studentID]...)
		studentSearchMap[studentID] = map[string]any{
			"nameKo":         nameKo,
			"nameJa":         japaneseStudentInfo[studentID].Name,
			"nameEn":         englishStudentInfo[studentID].Name,
			"nameZh":         chineseStudentInfo[studentID].Name,
			"searchKeywords": searchKeywords,
		}
	}

	if err := storage.MarshalAndUpload(log, studentSearchMap, "batorment/v3", "student-search-map.json", false, ""); err != nil {
		log.Warn("Failed to upload student search map", logger.Field{Key: "error", Value: err})
		return nil, err
	}

	return result, nil
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
