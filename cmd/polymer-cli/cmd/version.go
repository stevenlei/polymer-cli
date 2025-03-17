package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information
var (
	Version = "0.1.0"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("polymer-cli v%s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
