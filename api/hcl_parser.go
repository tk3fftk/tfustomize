package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

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

func (p HCLParser) PatchFileAttributes(base *hclwrite.File, overlay *hclwrite.File) (*hclwrite.File, error) {
	patchBodyAttributes(base.Body(), overlay.Body())
	return base, nil
}

func patchBodyAttributes(base *hclwrite.Body, overlay *hclwrite.Body) (*hclwrite.Body, error) {
	overlayAttributes := overlay.Attributes()

	// use overlay attributes if they exist
	for name, overlayAttribute := range overlayAttributes {
		// Parse the attribute's tokens into an expression
		// filename is used only for diagnostic messages. so it can be placeholder string.
		expr, diags := hclsyntax.ParseExpression(overlayAttribute.Expr().BuildTokens(nil).Bytes(), "overlays", hcl.InitialPos)
		if diags.HasErrors() {
			return nil, diags
		}

		// Evaluate the expression to get a cty.Value
		val, diags := expr.Value(nil)
		if diags.HasErrors() {
			return nil, diags
		}

		base.SetAttributeValue(name, val)
	}

	return base, nil
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
		case "provider", "resource", "data", "module", "terraform":
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
		switch blockType {
		case "provider", "resource", "data", "module", "terraform":
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

	for name, overlayLocalAttribute := range overlayLocals {
		baseLocals[name] = overlayLocalAttribute
	}
	resultedLocalBlock := hclwrite.NewBlock("locals", nil)
	for name, attribute := range baseLocals {
		resultedLocalBlock.Body().SetAttributeRaw(name, attribute.Expr().BuildTokens(nil))
	}
	base.AppendBlock(resultedLocalBlock)
	base.AppendNewline()

	for _, tmpBlock := range tmpBlocks {
		for _, block := range tmpBlock {
			base.AppendBlock(block)
		}
		base.AppendNewline()
	}

	return base, nil
}

func mergeBlock(baseBlock *hclwrite.Block, overlayBlock *hclwrite.Block) (*hclwrite.Block, error) {
	baseBlockBody := baseBlock.Body()
	overlayBlockBody := overlayBlock.Body()

	// どちらにも定義があるattributeをpatch
	patchBodyAttributes(baseBlockBody, overlayBlockBody)

	// overlay側にのみ定義があるattirbuteを追加
	// obtain and add attributes that are only defined in overlay
	overlayBodyAttributes := overlayBlockBody.Attributes()
	for name, overlayAttribute := range overlayBodyAttributes {
		if baseBlockBody.GetAttribute(name) == nil {
			baseBlockBody.SetAttributeRaw(name, overlayAttribute.Expr().BuildTokens(nil))
		}
	}

	// add blocks that are defined in overlay
	overlayBodyBlocks := overlayBlockBody.Blocks()
	for _, overlayBlock := range overlayBodyBlocks {
		baseBlockBody.AppendNewline()
		baseBlockBody.AppendBlock(overlayBlock)
	}

	return baseBlock, nil
}
