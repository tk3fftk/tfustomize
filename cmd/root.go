/*
Copyright © 2023 tk3fftk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logLevel = new(slog.LevelVar) // Info by default
var debug bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tfustomize",
	Short: "Customization of Terraform HCL",
	Long:  ``,
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

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tfustomize.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	if debug {
		h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
		slog.SetDefault(slog.New(h))
		logLevel.Set(slog.LevelDebug)
	}
}
