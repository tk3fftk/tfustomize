package api

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

var tfUniqueBlockTypes = []string{
	"data",
	"module",
	"output",
	"provider",
	"resource",
	"terraform",
	"variable",
}

var tfNoLableBlockTypes = []string{
	"moved",
	"import",
	"removed",
}

type HCLParser struct {
}

func NewHCLParser() *HCLParser {
	return &HCLParser{}
}

func (p HCLParser) ReadHCLFile(filename string) (*hclwrite.File, error) {
	output := hclwrite.NewEmptyFile()

	src, err := os.ReadFile(filename)
	if err != nil {
		return output, err
	}

	file, diags := hclwrite.ParseConfig(src, filename, hcl.InitialPos)
	if diags.HasErrors() {
		return output, fmt.Errorf(diags.Error())
	}

	return file, nil
}

// CollectHCLFilePaths returns a list of .if files in the given paths.
// If a path is a directory, it returns all .if files in the directory.
// If a path is a file, it returns the file if it has a .tf extension.
// The baseDir parameter is used as the root directory when constructing the full path of each file.
func (p HCLParser) CollectHCLFilePaths(baseDir string, paths []string) ([]string, error) {
	var collectedPaths []string

	for _, path := range paths {
		fullPath := filepath.Join(baseDir, path)
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			return nil, err
		}

		if fileInfo.IsDir() {
			fileInfos, err := os.ReadDir(fullPath)
			if err != nil {
				return nil, err
			}
			for _, fileInfo := range fileInfos {
				if filepath.Ext(fileInfo.Name()) == ".tf" {
					collectedPaths = append(collectedPaths, filepath.Join(fullPath, fileInfo.Name()))
				}
			}
		} else {
			if filepath.Ext(fileInfo.Name()) == ".tf" {
				collectedPaths = append(collectedPaths, fullPath)
			} else {
				slog.Warn("Only .tf file extension is supported, so ignore the file", "filename", fileInfo.Name())
			}
		}
	}

	return collectedPaths, nil
}

// ConcatFiles concatenates the contents of the given .tf files.
func (p HCLParser) ConcatFiles(paths []string) (*hclwrite.File, error) {
	outputFile := hclwrite.NewEmptyFile()

	for _, path := range paths {
		file, err := p.ReadHCLFile(path)
		if err != nil {
			return nil, err
		}
		for _, block := range file.Body().Blocks() {
			outputFile.Body().AppendBlock(block)
		}
	}

	return outputFile, nil
}

func setBodyAttribute(target *hclwrite.Body, name string, attr *hclwrite.Attribute) *hclwrite.Body {
	tokens := attr.Expr().BuildTokens(nil)
	// Do not want to treat as reference, traversal and cty.Value(literal) sogi use SetAttribute"Raw"
	target.SetAttributeRaw(name, tokens)

	return target
}

func (p HCLParser) MergeFileBlocks(base *hclwrite.File, overlay *hclwrite.File) (*hclwrite.File, error) {
	mergeBlocks(base.Body(), overlay.Body())
	return base, nil
}

func mergeBlocks(base *hclwrite.Body, overlay *hclwrite.Body) (*hclwrite.Body, error) {
	baseBlocks := base.Blocks()
	overlayBlocks := overlay.Blocks()

	tmpBlocks := map[string]map[string]*hclwrite.Block{}

	baseLocals := map[string]*hclwrite.Attribute{}
	overlayLocals := map[string]*hclwrite.Attribute{}

	for _, baseBlock := range baseBlocks {
		// From the perspective of Terraform, it seems that only up to two labels can be used,
		// but since the HCL specification allows setting three or more labels.
		// We concatenate the labels into a string to use as a key.
		joinedLabel := strings.Join(baseBlock.Labels(), "_")
		blockType := baseBlock.Type()
		if slices.Contains(tfUniqueBlockTypes, blockType) {
			if tmpBlocks[blockType] == nil {
				tmpBlocks[blockType] = map[string]*hclwrite.Block{}
			}

			tmpBlocks[blockType][joinedLabel] = baseBlock
		} else if blockType == "locals" {
			for name, attribute := range baseBlock.Body().Attributes() {
				baseLocals[name] = attribute
			}
		} else if slices.Contains(tfNoLableBlockTypes, blockType) {
			if tmpBlocks[blockType] == nil {
				tmpBlocks[blockType] = map[string]*hclwrite.Block{}
			}

			joinedLabel = fmt.Sprintf("%s%d", joinedLabel, len(tmpBlocks[blockType]))
			tmpBlocks[blockType][joinedLabel] = baseBlock
		} else {
			_ = fmt.Errorf("warn: type %v has come. it's ignored.", blockType)
		}
		base.RemoveBlock(baseBlock)
	}

	for _, overlayBlock := range overlayBlocks {
		joinedLabel := strings.Join(overlayBlock.Labels(), "_")
		blockType := overlayBlock.Type()
		slog.Debug("processing overlay blocks", "blockType", blockType, "joinedLabel", joinedLabel)

		if slices.Contains(tfUniqueBlockTypes, blockType) {
			if tmpBlock, ok := tmpBlocks[blockType][joinedLabel]; ok {
				mergedBlock, err := mergeBlock(tmpBlock, overlayBlock)
				if err != nil {
					return nil, err
				}
				tmpBlocks[blockType][joinedLabel] = mergedBlock
			} else {
				base.AppendBlock(overlayBlock)
				base.AppendNewline()
			}
		} else if blockType == "locals" {
			for name, attribute := range overlayBlock.Body().Attributes() {
				overlayLocals[name] = attribute
			}
		} else if slices.Contains(tfNoLableBlockTypes, blockType) {
			// There is no lable to identify the block, so we just append it.
			base.AppendBlock(overlayBlock)
			base.AppendNewline()
		} else {
			_ = fmt.Errorf("warn: type %v has come", blockType)
		}
	}

	if len(baseLocals) != 0 {
		for name, overlayLocalAttribute := range overlayLocals {
			baseLocals[name] = overlayLocalAttribute
		}

		sortedNames := make([]string, 0, len(baseLocals))
		for name := range baseLocals {
			sortedNames = append(sortedNames, name)
		}
		sort.Strings(sortedNames)

		resultedLocalBlock := hclwrite.NewBlock("locals", nil)
		for _, name := range sortedNames {
			setBodyAttribute(resultedLocalBlock.Body(), name, baseLocals[name])
		}
		base.AppendBlock(resultedLocalBlock)
		base.AppendNewline()
	}

	for _, blockType := range append(tfUniqueBlockTypes, tfNoLableBlockTypes...) {
		slog.Debug("processing result blocks", "blockType", blockType)
		if tmpBlocks[blockType] == nil {
			slog.Debug("blockType is nil, so skipped", "blockType", blockType)
			continue
		}
		for joinedLabel, block := range tmpBlocks[blockType] {
			slog.Debug("processing result blocks", "joinedLabel", joinedLabel)
			base.AppendBlock(block)
		}
		base.AppendNewline()
	}

	return base, nil
}

func mergeBlock(baseBlock *hclwrite.Block, overlayBlock *hclwrite.Block) (*hclwrite.Block, error) {
	resultBlock := hclwrite.NewBlock(baseBlock.Type(), baseBlock.Labels())
	resultBlockBody := resultBlock.Body()
	baseBlockBody := baseBlock.Body()
	overlayBlockBody := overlayBlock.Body()

	tmpAttributes := map[string]*hclwrite.Attribute{}

	for name, baseBlockBodyAttribute := range baseBlockBody.Attributes() {
		tmpAttributes[name] = baseBlockBodyAttribute
	}
	for name, overlayBlockBodyAttribute := range overlayBlockBody.Attributes() {
		tmpAttributes[name] = overlayBlockBodyAttribute
	}

	sortedNames := make([]string, 0, len(tmpAttributes))
	for name := range tmpAttributes {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	for _, name := range sortedNames {
		slog.Debug("processing attribute", "name", name, "value", tmpAttributes[name])
		setBodyAttribute(resultBlockBody, name, tmpAttributes[name])
	}

	// TODO: User can choose patch or append block
	// append blocks that are defined in overlay
	for _, baseBlockBodyBlock := range baseBlockBody.Blocks() {
		resultBlockBody.AppendNewline()
		resultBlockBody.AppendBlock(baseBlockBodyBlock)
	}
	for _, overlayBlockBodyBlock := range overlayBlockBody.Blocks() {
		resultBlockBody.AppendNewline()
		resultBlockBody.AppendBlock(overlayBlockBodyBlock)
	}

	return resultBlock, nil
}
