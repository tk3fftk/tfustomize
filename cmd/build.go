/*
Copyright © 2023 tk3fftk
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/cobra"
	"github.com/tk3fftk/tfustomize/api"
)

var regexpFormatNewLines = regexp.MustCompile(`\n{2,}`)
var print bool
var outputDir string
var outputFile string

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build [dir]",
	Short: "Build a tfustomization target from a directory.",
	Long: `The 'build' command constructs a tfustomization target from a specified directory. 
It checks for a 'tfustomization.hcl' file in the directory, loads the configuration.
The command concatenates files specified in the resources and patches blocks, merges them.`,
	Args: cobra.MaximumNArgs(1),
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
		if len(conf.Resources.Paths) == 0 {
			err := fmt.Errorf("tfustomization.hcl must have a resources block")
			return err
		}

		slog.Debug("tfustomization.hcl is loaded", "path", tfustomizationPath, "conf", conf)

		parser := api.NewHCLParser()

		basePaths, err := parser.CollectHCLFilePaths(filepath.Dir(tfustomizationPath), conf.Resources.Paths)
		if err != nil {
			return err
		}
		baseHCLFile, err := parser.ConcatFiles(basePaths)
		if err != nil {
			return err
		}

		overlayPaths, err := parser.CollectHCLFilePaths(filepath.Dir(tfustomizationPath), conf.Patches.Paths)
		if err != nil {
			return err
		}
		overlayHCLFile, err := parser.ConcatFiles(overlayPaths)
		if err != nil {
			return err
		}

		_, err = parser.MergeFileBlocks(baseHCLFile, overlayHCLFile)
		if err != nil {
			return err
		}

		result := regexpFormatNewLines.ReplaceAllString(string(hclwrite.Format(baseHCLFile.Bytes())), "\n")

		if print {
			fmt.Printf("%s", result)
		} else {
			outputDirPath := filepath.Join(baseConfDir, outputDir)
			if _, err := os.Stat(outputDirPath); os.IsNotExist(err) {
				err := os.Mkdir(outputDirPath, os.ModePerm)
				if err != nil {
					return err
				}
			}

			outputFilePath := filepath.Join(outputDirPath, outputFile)
			err := os.WriteFile(outputFilePath, []byte(result), 0666)
			if err != nil {
				return err
			}
		}

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
	buildCmd.Flags().BoolVarP(&print, "print", "p", false, "Print the result to the console instead of writing to a file")
	buildCmd.Flags().StringVarP(&outputDir, "out", "o", "generated", "Output directory")
	buildCmd.Flags().StringVarP(&outputFile, "outfile", "f", "main.tf", "Output filename")
}
