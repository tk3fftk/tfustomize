/*
Copyright © 2023 tk3fftk
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/cobra"
	"github.com/tk3fftk/tfustomize/api"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a kustomization target from a directory.",
	Long:  `Build a kustomization target from a directory.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		basedir := filepath.Base("")
		if len(args) == 1 {
			basedir = filepath.Join(basedir, args[0])
		}

		// to test
		conf, _ := api.LoadConfig(filepath.Join(basedir, "tfustomization.hcl"))

		fmt.Printf("%+v\n", conf)

		parser := api.NewHCLParser()

		base, err := parser.ReadHCLFile("base.hcl")
		if err != nil {
			panic(err)
		}
		overlay, err := parser.ReadHCLFile("overlay.hcl")
		if err != nil {
			panic(err)
		}
		_, err = parser.PatchFileAttributes(base, overlay)
		if err != nil {
			panic(err)
		}
		_, err = parser.MergeFileBlocks(base, overlay)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s", hclwrite.Format(base.Bytes()))
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
