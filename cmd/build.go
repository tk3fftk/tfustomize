/*
Copyright © 2023 tk3fftk
*/
package cmd

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/cobra"
	"github.com/tk3fftk/tfustomize/api"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		parser := api.NewHCLParser()

		base, err := parser.ReadHCLFile("base.hcl")
		if err != nil {
			panic(err)
		}
		overlay, err := parser.ReadHCLFile("overlay.hcl")
		if err != nil {
			panic(err)
		}
		output, err := parser.PatchFileAttributes(base, overlay)
		if err != nil {
			panic(err)
		}
		output, err = parser.MergeFileBlocks(base, overlay)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s", hclwrite.Format(output.Bytes()))
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
