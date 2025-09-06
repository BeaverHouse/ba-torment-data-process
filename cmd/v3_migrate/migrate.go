package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Old format structures
type OldPartyData struct {
	FinalRank   int              `json:"FINAL_RANK"`
	TormentRank int              `json:"TORMENT_RANK"`
	Score       int64            `json:"SCORE"`
	UserID      int              `json:"USER_ID"`
	Level       string           `json:"LEVEL"`
	PartyData   map[string][]int `json:"PARTY_DATA"`
}

type OldPartyFile struct {
	Filters       map[string][]int `json:"filters"`
	AssistFilters map[string][]int `json:"assist_filters"`
	MinPartys     int              `json:"min_partys"`
	MaxPartys     int              `json:"max_partys"`
	Parties       []OldPartyData   `json:"parties"`
}

type OldSummaryTorment struct {
	ClearCount    int              `json:"clear_count"`
	PartyCounts   map[string][]int `json:"party_counts"`
	Filters       map[string][]int `json:"filters"`
	AssistFilters map[string][]int `json:"assist_filters"`
	Top5Partys    []interface{}    `json:"top5_partys"`
}

type OldSummaryLunatic struct {
	ClearCount    int              `json:"clear_count"`
	PartyCounts   map[string][]int `json:"party_counts"`
	Filters       map[string][]int `json:"filters"`
	AssistFilters map[string][]int `json:"assist_filters"`
	Top5Partys    []interface{}    `json:"top5_partys"`
}

type OldSummaryFile struct {
	Torment OldSummaryTorment `json:"torment"`
	Lunatic OldSummaryLunatic `json:"lunatic"`
}

// New format structures
type NewPartyData struct {
	Rank      int     `json:"rank"`
	Score     int64   `json:"score"`
	PartyData [][]int `json:"partyData"`
}

type NewPartyFile struct {
	MinPartys int            `json:"minPartys"`
	MaxPartys int            `json:"maxPartys"`
	Parties   []NewPartyData `json:"parties"`
}

type NewSummaryTorment struct {
	ClearCount  int              `json:"clearCount"`
	PartyCounts map[string][]int `json:"partyCounts"`
	Top5Partys  []interface{}    `json:"top5Partys"`
}

type NewSummaryLunatic struct {
	ClearCount  int              `json:"clearCount"`
	PartyCounts map[string][]int `json:"partyCounts"`
	Top5Partys  []interface{}    `json:"top5Partys"`
}

type NewSummaryFile struct {
	Torment NewSummaryTorment `json:"torment"`
	Lunatic NewSummaryLunatic `json:"lunatic"`
}

type NewFilterFile struct {
	Filters       map[string]map[string]int `json:"filters"`
	AssistFilters map[string]map[string]int `json:"assistFilters"`
}

func convertArrayToMap(arr []int) map[string]int {
	result := make(map[string]int)

	// Array indices represent different star levels: [0*, 1*, 2*, 3*, 4*, 5*, 51, 52, 53, 54]
	starLevels := []string{"00", "10", "20", "30", "40", "50", "51", "52", "53", "54"}

	for i, count := range arr {
		if count > 0 {
			result[starLevels[i]] = count
		}
	}

	return result
}

func convertPartyData(oldParties []OldPartyData) ([]NewPartyData, int, int) {
	var newParties []NewPartyData
	minPartys := 99
	maxPartys := 0

	for _, party := range oldParties {
		newParty := NewPartyData{
			Rank:      party.FinalRank,
			Score:     party.Score,
			PartyData: make([][]int, 0),
		}

		// Convert party data from map to slice - process all parties without limit
		for i := 1; ; i++ {
			key := fmt.Sprintf("party_%d", i)
			if data, exists := party.PartyData[key]; exists {
				newParty.PartyData = append(newParty.PartyData, data)
			} else {
				break // No more parties found
			}
		}

		newParties = append(newParties, newParty)

		// Calculate min and max party counts
		partyCount := len(newParty.PartyData)
		if partyCount < minPartys {
			minPartys = partyCount
		}
		if partyCount > maxPartys {
			maxPartys = partyCount
		}
	}

	return newParties, minPartys, maxPartys
}

func migratePartyFile(oldPath, newPath string) error {
	// Read old file
	data, err := ioutil.ReadFile(oldPath)
	if err != nil {
		return fmt.Errorf("failed to read old party file: %v", err)
	}

	var oldFile OldPartyFile
	if err := json.Unmarshal(data, &oldFile); err != nil {
		return fmt.Errorf("failed to unmarshal old party file: %v", err)
	}

	// Convert to new format
	parties, minPartys, maxPartys := convertPartyData(oldFile.Parties)
	newFile := NewPartyFile{
		MinPartys: minPartys,
		MaxPartys: maxPartys,
		Parties:   parties,
	}

	// Create filter file
	filterFile := NewFilterFile{
		Filters:       make(map[string]map[string]int),
		AssistFilters: make(map[string]map[string]int),
	}

	for key, arr := range oldFile.Filters {
		filterFile.Filters[key] = convertArrayToMap(arr)
	}

	for key, arr := range oldFile.AssistFilters {
		filterFile.AssistFilters[key] = convertArrayToMap(arr)
	}

	// Write new party file
	newData, err := json.MarshalIndent(newFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal new party file: %v", err)
	}

	if err := ioutil.WriteFile(newPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to write new party file: %v", err)
	}

	// Write filter file
	filterPath := strings.Replace(newPath, "/party/", "/filter/", 1)
	filterData, err := json.MarshalIndent(filterFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal filter file: %v", err)
	}

	// Ensure filter directory exists
	filterDir := filepath.Dir(filterPath)
	if err := os.MkdirAll(filterDir, 0755); err != nil {
		return fmt.Errorf("failed to create filter directory: %v", err)
	}

	if err := ioutil.WriteFile(filterPath, filterData, 0644); err != nil {
		return fmt.Errorf("failed to write filter file: %v", err)
	}

	return nil
}

func migrateSummaryFile(oldPath, newPath string) error {
	// Read old file
	data, err := ioutil.ReadFile(oldPath)
	if err != nil {
		return fmt.Errorf("failed to read old summary file: %v", err)
	}

	var oldFile OldSummaryFile
	if err := json.Unmarshal(data, &oldFile); err != nil {
		return fmt.Errorf("failed to unmarshal old summary file: %v", err)
	}

	// Convert to new format
	newFile := NewSummaryFile{
		Torment: NewSummaryTorment{
			ClearCount:  oldFile.Torment.ClearCount,
			PartyCounts: oldFile.Torment.PartyCounts,
			Top5Partys:  oldFile.Torment.Top5Partys,
		},
		Lunatic: NewSummaryLunatic{
			ClearCount:  oldFile.Lunatic.ClearCount,
			PartyCounts: oldFile.Lunatic.PartyCounts,
			Top5Partys:  oldFile.Lunatic.Top5Partys,
		},
	}

	// Write new file
	newData, err := json.MarshalIndent(newFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal new summary file: %v", err)
	}

	if err := ioutil.WriteFile(newPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to write new summary file: %v", err)
	}

	return nil
}

func processFiles(oldDir, newDir string) error {
	// Ensure new directories exist
	dirs := []string{"party", "summary", "filter"}
	for _, dir := range dirs {
		fullPath := filepath.Join(newDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", fullPath, err)
		}
	}

	// Process party files
	partyDir := filepath.Join(oldDir, "party")
	if _, err := os.Stat(partyDir); err == nil {
		files, err := ioutil.ReadDir(partyDir)
		if err != nil {
			return fmt.Errorf("failed to read party directory: %v", err)
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") {
				oldPath := filepath.Join(partyDir, file.Name())
				newPath := filepath.Join(newDir, "party", file.Name())

				log.Printf("Migrating party file: %s -> %s", oldPath, newPath)
				if err := migratePartyFile(oldPath, newPath); err != nil {
					return fmt.Errorf("failed to migrate party file %s: %v", file.Name(), err)
				}
			}
		}
	}

	// Process summary files
	summaryDir := filepath.Join(oldDir, "summary")
	if _, err := os.Stat(summaryDir); err == nil {
		files, err := ioutil.ReadDir(summaryDir)
		if err != nil {
			return fmt.Errorf("failed to read summary directory: %v", err)
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") {
				oldPath := filepath.Join(summaryDir, file.Name())
				newPath := filepath.Join(newDir, "summary", file.Name())

				log.Printf("Migrating summary file: %s -> %s", oldPath, newPath)
				if err := migrateSummaryFile(oldPath, newPath); err != nil {
					return fmt.Errorf("failed to migrate summary file %s: %v", file.Name(), err)
				}
			}
		}
	}

	return nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run migrate.go <old_files_dir> <new_files_dir>")
		fmt.Println("Example: go run cmd/v3_migrate/migrate.go files/old files/v3")
		os.Exit(1)
	}

	oldDir := os.Args[1]
	newDir := os.Args[2]

	log.Printf("Starting migration from %s to %s", oldDir, newDir)

	if err := processFiles(oldDir, newDir); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully!")
}
