package cmd

import (
	"ba-torment-data-process/internal/logic/gridimage"

	"github.com/BeaverHouse/go-common/logger"
	"github.com/spf13/cobra"
)

var generateStudentGridImageCmd = &cobra.Command{
	Use:   "generate-student-grid-image",
	Short: "Generate student grid images",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		log, err := logger.NewLogger()
		if err != nil {
			return err
		}

		if err := gridimage.GenerateGridImages(log, dryRun); err != nil {
			return err
		}

		log.Info("Successfully generated all grid images")
		return nil
	},
}

func init() {
	generateStudentGridImageCmd.Flags().Bool("dry-run", false, "Save to local files/ directory instead of uploading")
	rootCmd.AddCommand(generateStudentGridImageCmd)
}
