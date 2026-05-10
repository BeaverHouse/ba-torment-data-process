package schaledb

import (
	"context"
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
	"ba-torment-data-process/internal/ui"

	"github.com/BeaverHouse/go-common/logger"
)

// Localization data structure
type LocalizationRawData struct {
	BuffName map[string]string `json:"BuffName"`
}

type JapaneseStudentInfo struct {
	Name       string   `json:"Name"`
	SearchTags []string `json:"SearchTags"`
}

// Reads /data/<lang>/localization.min.json file and returns BuffName map
func loadLocalization(lang string) (map[string]string, error) {
	url := fmt.Sprintf("%s/data/%s/localization.min.json", constants.SchaleDBURL, lang)
	byteValue, err := storage.GetDataFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch localization (%s): %w", lang, err)
	}

	var locData LocalizationRawData
	if err := json.Unmarshal(byteValue, &locData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal localization (%s): %w", lang, err)
	}

	return locData.BuffName, nil
}

// Reads /data/jp/students.json file and returns Japanese student info map.
// This is used to get student names in Japanese.
func loadJapaneseStudentInfo() (map[string]JapaneseStudentInfo, error) {
	byteValue, err := storage.GetDataFromURL(constants.SchaleDBURL + "/data/jp/students.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch japanese student info: %w", err)
	}

	var studentData map[string]JapaneseStudentInfo
	if err := json.Unmarshal(byteValue, &studentData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal japanese student info: %w", err)
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
func ParseSchaleDBStudents(db *postgres.Queries) (map[string]*types.StudentData, error) {

	byteValue, err := storage.GetDataFromURL(constants.SchaleDBURL + "/data/kr/students.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch students: %w", err)
	}

	var rawData map[string]any
	if err := json.Unmarshal(byteValue, &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal students: %w", err)
	}

	buffNames, err := loadLocalization("kr")
	if err != nil {
		return nil, err
	}
	japaneseStudentInfo, err := loadJapaneseStudentInfo()
	if err != nil {
		return nil, err
	}

	result := make(map[string]*types.StudentData)
	studentMap := make(map[string]string)

	for studentID, studentData := range rawData {
		dataMap, ok := studentData.(map[string]any)
		if !ok {
			ui.Log.Warn("Skipping invalid student data", logger.F("studentID", studentID))
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
			ui.Log.Warn("Failed to marshal student data", logger.F("studentID", studentID), logger.F("error", err))
			continue
		}

		var completeData types.StudentData
		err = json.Unmarshal(processedJSON, &completeData)
		if err != nil {
			ui.Log.Warn("Failed to unmarshal student data", logger.F("studentID", studentID), logger.F("error", err))
			continue
		}
		studentIDInt64, err := strconv.ParseInt(studentID, 10, 32)
		if err != nil {
			ui.Log.Warn("Failed to convert student ID to int", logger.F("studentID", studentID), logger.F("error", err))
			continue
		}
		completeDataBytes, err := json.Marshal(completeData)
		if err != nil {
			ui.Log.Warn("Failed to marshal student data", logger.F("studentID", studentID), logger.F("error", err))
			continue
		}

		db.InsertStudentData(context.Background(), postgres.InsertStudentDataParams{
			StudentID:     int32(studentIDInt64),
			NameKo:        completeData.Name,
			NameJa:        japaneseStudentInfo[studentID].Name,
			SearchKeyword: append(completeData.SearchTags, japaneseStudentInfo[studentID].SearchTags...),
			Detail:        completeDataBytes,
		})

		studentMap[studentID] = completeData.Name

		// Upload image + wait 3 second. Supabase S3 has performance issue when uploading too many files at once.
		err = uploadCharacterImage(int(studentIDInt64), false)
		if err != nil {
			ui.Log.Warn("Failed to upload image for student", logger.F("studentID", studentID), logger.F("error", err))
			return nil, err
		}
		ui.Log.Info("Student processed", logger.F("studentID", studentID), logger.F("name", completeData.Name))
	}

	if err := storage.MarshalAndUpload(studentMap, "batorment/v3", "student-map.json", false, ""); err != nil {
		ui.Log.Warn("Failed to upload student map", logger.F("error", err))
		return nil, err
	}

	// Create student search map with Japanese names and search keywords
	studentSearchMap := make(map[string]map[string]any)
	for studentID, studentData := range rawData {
		if dataMap, ok := studentData.(map[string]any); ok {
			if nameKo, exists := dataMap["Name"]; exists {
				// Convert []interface{} to []string for SearchTags
				var searchTags []string
				if tags, ok := dataMap["SearchTags"].([]interface{}); ok {
					for _, tag := range tags {
						if tagStr, ok := tag.(string); ok {
							searchTags = append(searchTags, tagStr)
						}
					}
				}

				searchData := map[string]any{
					"nameKo":         nameKo,
					"nameJa":         japaneseStudentInfo[studentID].Name,
					"searchKeywords": append(searchTags, japaneseStudentInfo[studentID].SearchTags...),
				}
				studentSearchMap[studentID] = searchData
			}
		}
	}

	if err := storage.MarshalAndUpload(studentSearchMap, "batorment/v3", "student-search-map.json", false, ""); err != nil {
		ui.Log.Warn("Failed to upload student search map", logger.F("error", err))
		return nil, err
	}

	return result, nil
}

// Uploads the character image from SchaleDB via File Manager API.
func uploadCharacterImage(id int, dryRun bool) error {
	imgBytes, err := storage.GetDataFromURL(constants.SchaleDBURL + "images/student/icon/" + strconv.Itoa(id) + ".webp")
	if err != nil {
		return fmt.Errorf("failed to fetch character image %d: %w", id, err)
	}

	path := "batorment/character"

	if err := storage.UploadFile(path, strconv.Itoa(id)+".webp", imgBytes, dryRun); err != nil {
		return fmt.Errorf("failed to upload image to S3: %w", err)
	}

	return nil
}
