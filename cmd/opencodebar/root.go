package opencodebar

import (
	"fmt"
	"os"

	"github.com/JValdivia23/quota-cli/pkg/opencodebar"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "opencodebar",
	Short: "A CLI tool to fetch and display HPC quota usage.",
	Long: `Quota CLI is a fast, cross-platform tool developed in Go 
to measure and report on user quotas in HPC environments like Derecho and Casper.

Running 'quota' without arguments will fetch and display the current quota usage.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior: Fetch and display quota
		fmt.Println("Fetching current quota usage...")

		// TODO: Extract project from config or arguments. Hardcoding for now.
		mockProject := "UCB-123"
		data, err := opencodebar.FetchQuota(mockProject)
		if err != nil {
			fmt.Printf("Error fetching quota: %v\n", err)
			return
		}

		fmt.Print(opencodebar.FormatQuota(data))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.quota-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".quota-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".quota-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
