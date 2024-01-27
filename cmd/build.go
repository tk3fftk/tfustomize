/*
Copyright © 2023 tk3fftk
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/cobra"
	"github.com/tk3fftk/tfustomize/api"
)

var regexpFormatNewLines = regexp.MustCompile(`\n{2,}`)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a tfustomization target from a directory.",
	Long:  `Build a tfustomization target from a directory.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		baseConfDir := filepath.Base("")
		if len(args) == 1 {
			baseConfDir = filepath.Join(baseConfDir, args[0])
		}
		tfustomizationPath := filepath.Join(baseConfDir, "tfustomization.hcl")

		if _, err := os.Stat(tfustomizationPath); err != nil {
			return err
		}

		conf, _ := api.LoadConfig(tfustomizationPath)
		if len(conf.Resources.Pathes) == 0 {
			err := fmt.Errorf("tfustomization.hcl must have a resources block")
			return err
		}

		fmt.Printf("%+v\n", conf)

		parser := api.NewHCLParser()

		baseHCLFile, err := parser.ConcatFile(filepath.Dir(tfustomizationPath), conf.Resources.Pathes)
		if err != nil {
			return err
		}
		overlayHCLFile, err := parser.ConcatFile(filepath.Dir(tfustomizationPath), conf.Patches.Pathes)
		if err != nil {
			return err
		}

		_, err = parser.PatchFileAttributes(baseHCLFile, overlayHCLFile)
		if err != nil {
			return err
		}
		_, err = parser.MergeFileBlocks(baseHCLFile, overlayHCLFile)
		if err != nil {
			return err
		}

		result := string(hclwrite.Format(baseHCLFile.Bytes()))
		fmt.Printf("%s", regexpFormatNewLines.ReplaceAllString(result, "\n"))

		return nil
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
