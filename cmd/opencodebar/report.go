package opencodebar

import (
	"fmt"

	"github.com/spf13/cobra"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a detailed usage report",
	Long: `Fetches comprehensive quota data and generates
a detailed usage report for the user's projects and scratch spaces.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Generating detailed quota report... (Not Yet Implemented)")
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
