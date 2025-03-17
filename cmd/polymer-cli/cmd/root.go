package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var apiKey string
var apiURL string
var debug bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "polymer-cli",
	Short: "A CLI tool for interacting with Polymer Prove API",
	Long: `polymer-cli is a command line tool to interact with the Polymer Prove API.
Learn more about the Prove API at https://docs.polymerlabs.org`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	// Disable the completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.polymer-cli.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Polymer API key")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "https://proof.testnet.polymer.zone", "Polymer API URL")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")

	// Bind flags to viper
	viper.BindPFlag("api-key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("api-url", rootCmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".polymer-cli" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".polymer-cli")
	}

	// Read environment variables with prefix POLYMER_
	viper.SetEnvPrefix("POLYMER")
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
