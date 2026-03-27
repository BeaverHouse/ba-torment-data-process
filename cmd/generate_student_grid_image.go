package cmd

import (
	"fmt"

	"ba-torment-data-process/internal/logic/gridimage"
	"ba-torment-data-process/internal/ui"

	"github.com/spf13/cobra"
)

var generateStudentGridImageCmd = &cobra.Command{
	Use:   "generate-student-grid-image",
	Short: "Generate student grid images",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if err := gridimage.GenerateGridImages(dryRun); err != nil {
			return fmt.Errorf("failed to generate grid images: %w", err)
		}

		ui.Log.Info("Successfully generated all grid images")
		return nil
	},
}

func init() {
	generateStudentGridImageCmd.Flags().Bool("dry-run", false, "Save to local files/ directory instead of uploading")
	rootCmd.AddCommand(generateStudentGridImageCmd)
}
