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

type JapaneseStudentInfo struct {
	Name       string   `json:"Name"`
	SearchTags []string `json:"SearchTags"`
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

// replaceBuffTags replaces <b:> and <d:> tags in skill descriptions with Korean names
func replaceBuffTags(desc string, buffNames map[string]string) string {
	// Handle <b:SOMETHING> tags
	bRegex := regexp.MustCompile(`<b:([^>]+)>`)
	desc = bRegex.ReplaceAllStringFunc(desc, func(match string) string {
		// Extract SOMETHING part
		tag := strings.TrimPrefix(strings.TrimSuffix(match, ">"), "<b:")

		// Find Buff_SOMETHING or Special_SOMETHING
		for _, prefix := range []string{"Buff_", "Special_", "CC_"} {
			if value, exists := buffNames[prefix+tag]; exists {
				return value
			}
		}

		// Return original if not found
		return match
	})

	// Handle <d:SOMEOTHERTHING> tags
	dRegex := regexp.MustCompile(`<d:([^>]+)>`)
	desc = dRegex.ReplaceAllStringFunc(desc, func(match string) string {
		// Extract SOMEOTHERTHING part
		tag := strings.TrimPrefix(strings.TrimSuffix(match, ">"), "<d:")

		// Find Debuff_SOMEOTHERTHING or Special_SOMEOTHERTHING
		for _, prefix := range []string{"Debuff_", "Special_", "CC_"} {
			if value, exists := buffNames[prefix+tag]; exists {
				return value
			}
		}

		// Return original if not found
		return match
	})

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

// ParseSchaleDBStudents parses SchaleDB data and returns processed student data
func ParseSchaleDBStudents(log logger.Logger, db *postgres.Queries) (map[string]*types.StudentData, error) {

	byteValue, err := storage.GetDataFromURL(log, constants.SchaleDBURL+"/data/kr/students.json")
	if err != nil {
		return nil, constants.ErrDataFetch("students", err)
	}

	var rawData map[string]any
	if err := json.Unmarshal(byteValue, &rawData); err != nil {
		return nil, constants.ErrDataDecode("students", err)
	}

	buffNames, err := loadLocalization(log, "kr")
	if err != nil {
		return nil, err
	}
	japaneseStudentInfo, err := loadJapaneseStudentInfo(log)
	if err != nil {
		return nil, err
	}
	// FE multi-lang: nameEn / nameZh + extra search tags. Non-fatal if upstream
	// 404s — we just skip that locale and the FE falls back to ko.
	englishStudentInfo, err := loadStudentNames(log, "en")
	if err != nil {
		log.Warn("Failed to load English student names", logger.Field{Key: "error", Value: err})
		englishStudentInfo = map[string]JapaneseStudentInfo{}
	}
	chineseStudentInfo, err := loadStudentNames(log, "zh")
	if err != nil {
		log.Warn("Failed to load Chinese student names", logger.Field{Key: "error", Value: err})
		chineseStudentInfo = map[string]JapaneseStudentInfo{}
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
		mergedTags := append([]string{}, completeData.SearchTags...)
		mergedTags = append(mergedTags, japaneseStudentInfo[studentID].SearchTags...)
		mergedTags = append(mergedTags, englishStudentInfo[studentID].SearchTags...)
		mergedTags = append(mergedTags, chineseStudentInfo[studentID].SearchTags...)
		mergedTags = append(mergedTags, zhAliases[studentID]...)

		db.InsertStudentData(context.Background(), postgres.InsertStudentDataParams{
			StudentID:     int32(studentIDInt64),
			NameKo:        completeData.Name,
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
		dataMap, ok := studentData.(map[string]any)
		if !ok {
			continue
		}
		nameKo, exists := dataMap["Name"]
		if !exists {
			continue
		}
		var koTags []string
		if tags, ok := dataMap["SearchTags"].([]interface{}); ok {
			for _, tag := range tags {
				if s, ok := tag.(string); ok {
					koTags = append(koTags, s)
				}
			}
		}
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
