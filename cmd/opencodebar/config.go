package opencodebar

import (
	"fmt"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long: `Manage Quota CLI configuration settings such as 
API keys, target environments, and default user configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Managing configurations... (Not Yet Implemented)")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
