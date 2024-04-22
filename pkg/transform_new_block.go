package pkg

import (
	"fmt"
	"github.com/Azure/golden"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

var _ golden.ApplyBlock = &NewBlockTransform{}
var _ golden.CustomDecodeBase = &NewBlockTransform{}
var _ mptfBlock = &NewBlockTransform{}

type NewBlockTransform struct {
	*golden.BaseBlock
	*BaseTransform
	NewBlockType  string   `hcl:"new_block_type"`
	FileName      string   `hcl:"filename" validate:"endswith=.tf"`
	Labels        []string `hcl:"labels,optional"`
	newWriteBlock *hclwrite.Block
}

func (n *NewBlockTransform) isReservedField(name string) bool {
	reserved := map[string]struct{}{
		"new_block_type": {},
		"for_each":       {},
		"asraw":          {},
		"labels":         {},
		"filename":       {},
	}
	_, ok := reserved[name]
	return ok
}

func (n *NewBlockTransform) Decode(block *golden.HclBlock, context *hcl.EvalContext) error {
	var err error
	n.NewBlockType, err = getRequiredStringAttribute("new_block_type", block, context)
	if err != nil {
		return err
	}
	n.FileName, err = getRequiredStringAttribute("filename", block, context)
	if err != nil {
		return err
	}
	n.Labels = nil
	labelsAttr, ok := block.Attributes()["labels"]
	if ok {
		labelsValue, err := labelsAttr.Value(context)
		if err != nil {
			return fmt.Errorf("error while evaluating labels: %+v", err)
		}
		for i := 0; i < labelsValue.LengthInt(); i++ {
			n.Labels = append(n.Labels, labelsValue.Index(cty.NumberIntVal(int64(i))).AsString())
		}
	}
	n.newWriteBlock = hclwrite.NewBlock(n.NewBlockType, n.Labels)
	for _, b := range block.NestedBlocks() {
		if b.Type == "asraw" {
			if err := decodeAsRawBlock(n.newWriteBlock, b); err != nil {
				return err
			}
			continue
		}
	}
	_ = decodeAsStringBlock(n, n.newWriteBlock, block, 0, context)
	return nil
}

func (n *NewBlockTransform) Type() string {
	return "new_block"
}

func (n *NewBlockTransform) Apply() error {
	n.Config().(*MetaProgrammingTFConfig).AddBlock(n.FileName, n.newWriteBlock)
	return nil
}

func (n *NewBlockTransform) NewWriteBlock() *hclwrite.Block {
	return n.newWriteBlock
}

func getRequiredStringAttribute(name string, block *golden.HclBlock, context *hcl.EvalContext) (string, error) {
	targetBlockAddress, ok := block.Attributes()[name]
	if !ok {
		return "", fmt.Errorf("`%s` is required", name)
	}
	v, err := targetBlockAddress.Value(context)
	if err != nil {
		return "", err
	}
	if v.Type() != cty.String {
		return "", fmt.Errorf("`%s` must be a string", name)
	}
	return v.AsString(), nil
}
