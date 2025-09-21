package schaledb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ba-torment-data-process/internal/db/postgres"
	logic_download "ba-torment-data-process/internal/logic/download"
	logic_upload "ba-torment-data-process/internal/logic/upload"
	"ba-torment-data-process/internal/types"
)

const schaleDBURL = "https://schaledb.com/"

// Localization data structure
type LocalizationRawData struct {
	BuffName map[string]string `json:"BuffName"`
}

type JapaneseStudentInfo struct {
	Name       string   `json:"Name"`
	SearchTags []string `json:"SearchTags"`
}

// Reads /data/kr/localization.min.json file and returns BuffName map
func loadBuffLocalization() map[string]string {
	byteValue := logic_download.GetDataFromURL(schaleDBURL + "/data/kr/localization.min.json")

	var locData LocalizationRawData
	err := json.Unmarshal(byteValue, &locData)
	if err != nil {
		log.Fatalf("Failed to unmarshal localization: %v", err)
	}

	return locData.BuffName
}

// Reads /data/jp/students.json file and returns Japanese student info map.
// This is used to get student names in Japanese.
func loadJapaneseStudentInfo() map[string]JapaneseStudentInfo {
	byteValue := logic_download.GetDataFromURL(schaleDBURL + "/data/jp/students.json")

	var studentData map[string]JapaneseStudentInfo
	err := json.Unmarshal(byteValue, &studentData)
	if err != nil {
		log.Fatalf("Failed to unmarshal localization: %v", err)
	}

	return studentData
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

// UnusedProperties defines fields to remove from raw SchaleDB data
var UnusedProperties = []string{
	"Id",
	"IsReleased", // Release status
	"DefaultOrder",
	"PathName",
	"Icon",            // Icon
	"WeaponImg",       // Weapon image
	"Cover",           // Cover
	"Size",            // Size (type when appearing as enemy)
	"CollectionBG",    // Background
	"CharacterSSRNew", // Dialogue when joining
	"Illustrator",     // Illustrator
	"Designer",        // Designer
	"CharacterVoice",  // Voice actor
	"CharHeightImperial",
	"AmmoCost",              // Ammo cost
	"AmmoCount",             // Ammo count
	"MemoryLobby",           // Memorial unlock rank
	"MemoryLobbyBGM",        // Memorial BGM
	"SkillExMaterial",       // EX skill enhancement material
	"SkillExMaterialAmount", // EX skill enhancement material amount
	"SkillMaterial",         // Skill enhancement material
	"SkillMaterialAmount",   // Skill enhancement material amount
	"PotentialMaterial",     // Potential material
	"RegenCost",             // Cost recovery speed (fixed at 700)
	"CriticalDamageRate",    // Critical damage rate (fixed at 20000, 200%)
	"Birthday",              // Birthday (duplicate)
}

// ParseSchaleDBStudents parses SchaleDB data and returns processed student data
func ParseSchaleDBStudents(db *postgres.Queries) (map[string]*types.StudentData, error) {

	byteValue := logic_download.GetDataFromURL(schaleDBURL + "/data/kr/students.json")

	var rawData map[string]any
	err := json.Unmarshal(byteValue, &rawData)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	buffNames := loadBuffLocalization()
	japaneseStudentInfo := loadJapaneseStudentInfo()

	result := make(map[string]*types.StudentData)
	studentMap := make(map[string]string)

	for studentID, studentData := range rawData {
		dataMap, ok := studentData.(map[string]any)
		if !ok {
			log.Printf("Skipping invalid student data for ID %s", studentID)
			continue
		}

		// Remove unused properties
		for _, property := range UnusedProperties {
			delete(dataMap, property)
		}

		// Process Gear data
		if gear, ok := dataMap["Gear"]; ok {
			if gearMap, ok := gear.(map[string]any); ok {
				delete(gearMap, "TierUpMaterial")
				delete(gearMap, "TierUpMaterialAmount")
				delete(gearMap, "Desc")
				delete(gearMap, "Released")
			}
		}

		// Process Skills data
		if skills, ok := dataMap["Skills"]; ok {
			if skillsMap, ok := skills.(map[string]any); ok {
				delete(skillsMap, "Normal")

				for _, skill := range skillsMap {
					if skillMap, ok := skill.(map[string]any); ok {
						delete(skillMap, "Icon")
						delete(skillMap, "Name")
						delete(skillMap, "Duration")
						delete(skillMap, "Range")
						delete(skillMap, "Radius")

						// Process Effects
						if effects, ok := skillMap["Effects"]; ok {
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

						// Process Desc and Parameters
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
							// Delete Parameters after processing
							delete(skillMap, "Parameters")
						}
					}
				}
			}
		}

		// Convert processed map back to CompleteStudentData struct
		processedJSON, err := json.Marshal(dataMap)
		if err != nil {
			log.Printf("Failed to marshal student data for ID %s: %v", studentID, err)
			continue
		}

		var completeData types.StudentData
		err = json.Unmarshal(processedJSON, &completeData)
		if err != nil {
			log.Printf("Failed to unmarshal student data for ID %s: %v", studentID, err)
			continue
		}
		studentIDInt, err := strconv.Atoi(studentID)
		if err != nil {
			log.Printf("Failed to convert student ID %s to int: %v", studentID, err)
			continue
		}
		completeDataBytes, err := json.Marshal(completeData)
		if err != nil {
			log.Printf("Failed to marshal student data for ID %s: %v", studentID, err)
			continue
		}

		db.InsertStudentData(context.Background(), postgres.InsertStudentDataParams{
			StudentID:     int32(studentIDInt),
			NameKo:        completeData.Name,
			NameJa:        japaneseStudentInfo[studentID].Name,
			SearchKeyword: append(completeData.SearchTags, japaneseStudentInfo[studentID].SearchTags...),
			Detail:        completeDataBytes,
		})

		studentMap[studentID] = completeData.Name

		// Upload image + wait 3 second. Supabase S3 has performance issue when uploading too many files at once.
		err = uploadCharacterImage(studentIDInt, false)
		if err != nil {
			log.Printf("Failed to upload image for student %s: %v", studentID, err)
			return nil, err
		}
		log.Printf("Student %s (%s) processed", studentID, completeData.Name)

		time.Sleep(3 * time.Second)
	}

	studentMapBytes, err := json.Marshal(studentMap)
	if err != nil {
		log.Printf("Failed to marshal student map: %v", err)
		return nil, err
	}

	err = logic_upload.UploadFile("batorment/v3", "student-map.json", studentMapBytes, false)
	if err != nil {
		log.Printf("Failed to upload student map: %v", err)
		return nil, err
	}

	// Create student search map with Japanese names and search keywords
	studentSearchMap := make(map[string]map[string]interface{})
	for studentID, studentData := range rawData {
		if dataMap, ok := studentData.(map[string]any); ok {
			if nameKo, exists := dataMap["Name"]; exists {
				searchData := map[string]interface{}{
					"nameKo":         nameKo,
					"nameJa":         japaneseStudentInfo[studentID].Name,
					"searchKeywords": append(dataMap["SearchTags"].([]string), japaneseStudentInfo[studentID].SearchTags...),
				}
				studentSearchMap[studentID] = searchData
			}
		}
	}

	studentSearchMapBytes, err := json.Marshal(studentSearchMap)
	if err != nil {
		log.Printf("Failed to marshal student search map: %v", err)
		return nil, err
	}

	err = logic_upload.UploadFile("batorment/v3", "student-search-map.json", studentSearchMapBytes, false)
	if err != nil {
		log.Printf("Failed to upload student search map: %v", err)
		return nil, err
	}

	return result, nil
}

// Uploads the character image from SchaleDB via File Manager API.
func uploadCharacterImage(id int, dryRun bool) error {

	imgBytes := logic_download.GetDataFromURL(schaleDBURL + "images/student/icon/" + strconv.Itoa(id) + ".webp")

	path := "batorment/character"

	err := logic_upload.UploadFile(path, strconv.Itoa(id)+".webp", imgBytes, dryRun)
	if err != nil {
		return fmt.Errorf("failed to upload image to S3: %v", err)
	}

	return nil
}
