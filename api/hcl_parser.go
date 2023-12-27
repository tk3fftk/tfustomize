package api

import (
	"fmt"
	"os"
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

func (p HCLParser) PatchFileAttributes(base *hclwrite.File, overlay *hclwrite.File) (*hclwrite.File, error) {
	patchBodyAttributes(base.Body(), overlay.Body())
	base.Body().AppendNewline()
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

	baseProviderBlocks := map[string]*hclwrite.Block{}
	baseResourceBlocks := map[string]*hclwrite.Block{}
	baseDataBlocks := map[string]*hclwrite.Block{}

	baseLocals := map[string]*hclwrite.Attribute{}
	overlayLocals := map[string]*hclwrite.Attribute{}

	for _, baseBlock := range baseBlocks {
		joinedLabel := strings.Join(baseBlock.Labels(), "_")
		switch baseBlock.Type() {
		case "provider":
			baseProviderBlocks[joinedLabel] = baseBlock
		case "resource":
			baseResourceBlocks[joinedLabel] = baseBlock
		case "data":
			baseDataBlocks[joinedLabel] = baseBlock
		case "locals":
			for name, attribute := range baseBlock.Body().Attributes() {
				baseLocals[name] = attribute
			}
		default:
			_ = fmt.Errorf("warn: type %v has come", baseBlock.Type())
		}
		base.RemoveBlock(baseBlock)
	}

	// baseにあるblockをoverlayで上書きして一時保管、なければbodyに直接追加
	for _, overlayBlock := range overlayBlocks {
		joinedLabel := strings.Join(overlayBlock.Labels(), "_")
		switch overlayBlock.Type() {
		case "provider":
			if baseProviderBlock, ok := baseProviderBlocks[joinedLabel]; ok {
				mergedBlock, err := mergeBlock(baseProviderBlock, overlayBlock)
				if err != nil {
					return nil, err
				}
				baseProviderBlocks[joinedLabel] = mergedBlock
			} else {
				base.AppendBlock(overlayBlock)
			}
		case "resource":
			if baseResourceBlock, ok := baseResourceBlocks[joinedLabel]; ok {
				mergedBlock, err := mergeBlock(baseResourceBlock, overlayBlock)
				if err != nil {
					return nil, err
				}
				baseResourceBlocks[joinedLabel] = mergedBlock
			} else {
				base.AppendBlock(overlayBlock)
			}
		case "data":
			if baseDataBlock, ok := baseDataBlocks[joinedLabel]; ok {
				mergedBlock, err := mergeBlock(baseDataBlock, overlayBlock)
				if err != nil {
					return nil, err
				}
				baseDataBlocks[joinedLabel] = mergedBlock
			} else {
				base.AppendBlock(overlayBlock)
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

	for _, baseProviderBlock := range baseProviderBlocks {
		base.AppendBlock(baseProviderBlock)
		base.AppendNewline()
	}
	for _, baseResourceBlock := range baseResourceBlocks {
		base.AppendBlock(baseResourceBlock)
		base.AppendNewline()
	}
	for _, baseDataBlock := range baseDataBlocks {
		base.AppendBlock(baseDataBlock)
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
