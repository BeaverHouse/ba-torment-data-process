package gridimage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"sort"
	"strconv"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/logic/id"
	"ba-torment-data-process/internal/logic/storage"
	"ba-torment-data-process/internal/types"

	"github.com/BeaverHouse/go-common/errorhandle"
	"github.com/BeaverHouse/go-common/logger"
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
func GenerateGridImages(log logger.Logger, dryRun bool) error {
	students, err := fetchAllStudents(log)
	if err != nil {
		return err
	}

	totalCount := len(students)
	log.Info("Total students with usage data", logger.Field{Key: "count", Value: totalCount})

	// Calculate percentile cutoffs
	tier1StartIdx := totalCount * tier1Start / 100
	tier1EndIdx := totalCount * tier1End / 100
	tier2StartIdx := totalCount * tier2Start / 100
	tier2EndIdx := totalCount * tier2End / 100
	tier3StartIdx := totalCount * tier3Start / 100
	tier3EndIdx := totalCount * tier3End / 100

	log.Info("Percentile cutoffs",
		logger.Field{Key: "tier1", Value: fmt.Sprintf("%d-%d", tier1StartIdx, tier1EndIdx)},
		logger.Field{Key: "tier2", Value: fmt.Sprintf("%d-%d", tier2StartIdx, tier2EndIdx)},
		logger.Field{Key: "tier3", Value: fmt.Sprintf("%d-%d", tier3StartIdx, tier3EndIdx)})

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

		log.Info("Tier stats",
			logger.Field{Key: "tier", Value: tier.name + "%"},
			logger.Field{Key: "total", Value: len(tierStudents)},
			logger.Field{Key: "strikers", Value: len(strikers)},
			logger.Field{Key: "specials", Value: len(specials)})

		strikerFile := fmt.Sprintf("grid_striker_%s.jpg", tier.name)
		specialFile := fmt.Sprintf("grid_special_%s.jpg", tier.name)

		if err := generateAndUploadGrid(log, strikers, strikerFile, dryRun); err != nil {
			return err
		}
		if err := generateAndUploadGrid(log, specials, specialFile, dryRun); err != nil {
			return err
		}
	}

	log.Info("Successfully generated 6 grids")
	return nil
}

func fetchAllStudents(log logger.Logger) ([]StudentInfo, error) {
	url := supabaseBaseURL + "/batorment/v3/total-analysis.json"
	data, err := storage.GetDataFromURL(log, url)
	if err != nil {
		return nil, constants.ErrDataFetch("total-analysis", err)
	}

	var analysis types.TotalAnalysisOutput
	if err := json.Unmarshal(data, &analysis); err != nil {
		return nil, constants.ErrDataDecode("total-analysis", err)
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

func generateAndUploadGrid(log logger.Logger, students []StudentInfo, fileName string, dryRun bool) error {
	if len(students) == 0 {
		log.Warn("Skipping grid", logger.Field{Key: "file", Value: fileName}, logger.Field{Key: "reason", Value: "no students"})
		return nil
	}

	img, err := generateGrid(log, students)
	if err != nil {
		return constants.ErrGridGenerate(fileName, err)
	}

	if err := uploadGridImage(log, img, fileName, dryRun); err != nil {
		return constants.ErrUpload(fmt.Sprintf("grid %s", fileName), err)
	}

	log.Info("Generated grid", logger.Field{Key: "file", Value: fileName}, logger.Field{Key: "students", Value: len(students)})
	return nil
}

func generateGrid(log logger.Logger, students []StudentInfo) (image.Image, error) {
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
		portrait, err := fetchPortraitFromSupabase(log, student.ID)
		if err != nil {
			log.Warn("Failed to fetch portrait", logger.Field{Key: "studentID", Value: student.ID}, logger.Field{Key: "error", Value: err})
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

func fetchPortraitFromSupabase(log logger.Logger, studentID int) (image.Image, error) {
	url := supabaseBaseURL + "/batorment/character/" + strconv.Itoa(studentID) + ".webp"
	data, err := storage.GetDataFromURL(log, url)
	if err != nil {
		return nil, constants.ErrDataFetch(fmt.Sprintf("portrait %d", studentID), err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, constants.ErrDataDecode(fmt.Sprintf("portrait %d", studentID), err)
	}

	return img, nil
}

func uploadGridImage(log logger.Logger, img image.Image, fileName string, dryRun bool) error {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return errorhandle.ErrInternal(err)
	}
	return storage.UploadFile(log, "batorment/v3/grid", fileName, buf.Bytes(), dryRun)
}
