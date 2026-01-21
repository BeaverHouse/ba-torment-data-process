package gridimage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"sort"
	"strconv"

	"ba-torment-data-process/internal/logic/id"
	"ba-torment-data-process/internal/logic/storage"
	"ba-torment-data-process/internal/types"

	"github.com/fogleman/gg"
	_ "golang.org/x/image/webp"
)

const (
	cellWidth       = 120
	cellHeight      = 150
	cols            = 10
	fontSize        = 18
	padding         = 4
	supabaseBaseURL = "https://twauaebyyujvvvusbrwe.supabase.co/storage/v1/object/public/pb7h4uvn2b6m0lyu7i6r3j8ac"
)

// Percentile-based tiers
const (
	tier1Start = 10 // 10-25%
	tier1End   = 25
	tier2Start = 25 // 25-40%
	tier2End   = 40
	tier3Start = 40 // 40-55%
	tier3End   = 55
)

type StudentInfo struct {
	ID         int
	TotalUsage int
	IsStriker  bool
}

// GenerateGridImages generates images based on percentile tiers:
// - 10-25%, 25-40%, 40-55%: grids for Striker/Special (6 grids)
// Total: 6 images
func GenerateGridImages(dryRun bool) error {
	students, err := fetchAllStudents()
	if err != nil {
		return fmt.Errorf("failed to fetch students: %w", err)
	}

	totalCount := len(students)
	log.Printf("Total students with usage data: %d", totalCount)

	// Calculate percentile cutoffs
	tier1StartIdx := totalCount * tier1Start / 100
	tier1EndIdx := totalCount * tier1End / 100
	tier2StartIdx := totalCount * tier2Start / 100
	tier2EndIdx := totalCount * tier2End / 100
	tier3StartIdx := totalCount * tier3Start / 100
	tier3EndIdx := totalCount * tier3End / 100

	log.Printf("Percentile cutoffs: Tier1=%d-%d, Tier2=%d-%d, Tier3=%d-%d",
		tier1StartIdx, tier1EndIdx, tier2StartIdx, tier2EndIdx, tier3StartIdx, tier3EndIdx)

	// Generate tier grids
	tiers := []struct {
		name  string
		start int
		end   int
	}{
		{"10-25", tier1StartIdx, tier1EndIdx},
		{"25-40", tier2StartIdx, tier2EndIdx},
		{"40-55", tier3StartIdx, tier3EndIdx},
	}

	for _, tier := range tiers {
		tierStudents := students[tier.start:tier.end]
		strikers, specials := splitBySquadType(tierStudents)

		log.Printf("Tier %s%%: Total=%d, Strikers=%d, Specials=%d",
			tier.name, len(tierStudents), len(strikers), len(specials))

		strikerFile := fmt.Sprintf("grid_striker_%s.jpg", tier.name)
		specialFile := fmt.Sprintf("grid_special_%s.jpg", tier.name)

		if err := generateAndUploadGrid(strikers, strikerFile, dryRun); err != nil {
			return err
		}
		if err := generateAndUploadGrid(specials, specialFile, dryRun); err != nil {
			return err
		}
	}

	log.Printf("Successfully generated 6 grids")
	return nil
}

func fetchAllStudents() ([]StudentInfo, error) {
	url := supabaseBaseURL + "/batorment/v3/total-analysis.json"
	data := storage.GetDataFromURL(url)

	var analysis types.TotalAnalysisOutput
	if err := json.Unmarshal(data, &analysis); err != nil {
		return nil, fmt.Errorf("failed to unmarshal total-analysis: %w", err)
	}

	var students []StudentInfo
	for _, char := range analysis.CharacterAnalyses {
		students = append(students, StudentInfo{
			ID:         char.StudentID,
			TotalUsage: char.TotalUsage,
			IsStriker:  id.IsStriker(char.StudentID),
		})
	}

	// Sort by totalUsage descending
	sort.Slice(students, func(i, j int) bool {
		return students[i].TotalUsage > students[j].TotalUsage
	})

	return students, nil
}

func splitBySquadType(students []StudentInfo) ([]StudentInfo, []StudentInfo) {
	var strikers, specials []StudentInfo
	for _, s := range students {
		if s.IsStriker {
			strikers = append(strikers, s)
		} else {
			specials = append(specials, s)
		}
	}
	return strikers, specials
}

func generateAndUploadGrid(students []StudentInfo, fileName string, dryRun bool) error {
	if len(students) == 0 {
		log.Printf("Skipping %s: no students", fileName)
		return nil
	}

	img, err := generateGrid(students)
	if err != nil {
		return fmt.Errorf("failed to generate grid %s: %w", fileName, err)
	}

	if err := uploadGridImage(img, fileName, dryRun); err != nil {
		return fmt.Errorf("failed to upload grid %s: %w", fileName, err)
	}

	log.Printf("Generated: %s (%d students)", fileName, len(students))
	return nil
}

func generateGrid(students []StudentInfo) (image.Image, error) {
	rows := (len(students) + cols - 1) / cols
	width := cols * cellWidth
	height := rows * cellHeight

	dc := gg.NewContext(width, height)
	dc.SetRGB(1, 1, 1) // white background
	dc.Clear()

	// Draw all cells (background + border)
	for row := range rows {
		for col := range cols {
			x := float64(col * cellWidth)
			y := float64(row * cellHeight)

			// Checkerboard background
			if (row+col)%2 == 0 {
				dc.SetRGB(0.95, 0.95, 0.95) // light gray
			} else {
				dc.SetRGB(1, 1, 1) // white
			}
			dc.DrawRectangle(x, y, cellWidth, cellHeight)
			dc.Fill()

			// Border
			dc.SetRGB(0.8, 0.8, 0.8) // gray border
			dc.SetLineWidth(1)
			dc.DrawRectangle(x, y, cellWidth, cellHeight)
			dc.Stroke()
		}
	}

	// Draw students
	for i, student := range students {
		col := i % cols
		row := i / cols
		x := float64(col * cellWidth)
		y := float64(row * cellHeight)

		// Draw portrait from Supabase
		portrait, err := fetchPortraitFromSupabase(student.ID)
		if err != nil {
			log.Printf("Failed to fetch portrait for %d: %v", student.ID, err)
			continue
		}
		dc.DrawImage(portrait, int(x)+padding, int(y)+padding+fontSize+6)

		// Draw student ID
		dc.SetRGB(0, 0, 0)
		if err := dc.LoadFontFace("/System/Library/Fonts/Helvetica.ttc", fontSize); err != nil {
			dc.LoadFontFace("", fontSize) // fallback
		}
		idText := strconv.Itoa(student.ID)
		dc.DrawStringAnchored(idText, x+float64(cellWidth)/2, y+fontSize, 0.5, 0.5)
	}

	return dc.Image(), nil
}

func fetchPortraitFromSupabase(studentID int) (image.Image, error) {
	url := supabaseBaseURL + "/batorment/character/" + strconv.Itoa(studentID) + ".webp"
	data := storage.GetDataFromURL(url)

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

func uploadGridImage(img image.Image, fileName string, dryRun bool) error {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}
	return storage.UploadFile("batorment/v3/grid", fileName, buf.Bytes(), dryRun)
}
