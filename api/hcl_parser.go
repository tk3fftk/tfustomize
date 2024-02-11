package api

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

var tfBlockTypes = []string{"data", "module", "output", "provider", "resource", "terraform", "variable"}

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

func (p HCLParser) ConcatFile(baseDir string, pathes []string) (*hclwrite.File, error) {
	outputFile := hclwrite.NewEmptyFile()

	for _, path := range pathes {
		file, err := p.ReadHCLFile(filepath.Join(baseDir, path))
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
		// TF的にはラベルは2つまでの利用に見えるが、HCLの仕様上はlabelを3つ以上設定できるので念のためラベルを結合した文字列をキーにしておく
		joinedLabel := strings.Join(baseBlock.Labels(), "_")
		blockType := baseBlock.Type()
		switch blockType {
		case "data", "module", "output", "provider", "resource", "terraform", "variable":
			if tmpBlocks[blockType] == nil {
				tmpBlocks[blockType] = map[string]*hclwrite.Block{}
			}

			tmpBlocks[blockType][joinedLabel] = baseBlock
		case "locals":
			for name, attribute := range baseBlock.Body().Attributes() {
				baseLocals[name] = attribute
			}
		default:
			_ = fmt.Errorf("warn: type %v has come", baseBlock.Type())
		}
		base.RemoveBlock(baseBlock)
	}

	// baseにあるblockをoverlayで上書きして一時保管、baseになければoverlayの値をbodyに直接追加
	for _, overlayBlock := range overlayBlocks {
		joinedLabel := strings.Join(overlayBlock.Labels(), "_")
		blockType := overlayBlock.Type()
		slog.Debug("processing overlay blocks", "blockType", blockType, "joinedLabel", joinedLabel)

		switch blockType {
		case "data", "module", "output", "provider", "resource", "terraform", "variable":
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
		case "locals":
			for name, attribute := range overlayBlock.Body().Attributes() {
				overlayLocals[name] = attribute
			}
		default:
			_ = fmt.Errorf("warn: type %v has come", overlayBlock.Type())
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

	for _, blockType := range tfBlockTypes {
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

	for name, baseBlockBodyArrtibute := range baseBlockBody.Attributes() {
		tmpAttributes[name] = baseBlockBodyArrtibute
	}
	for name, overlayBlockBodyArrtibute := range overlayBlockBody.Attributes() {
		tmpAttributes[name] = overlayBlockBodyArrtibute
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
