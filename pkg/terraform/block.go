package terraform

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"sort"
	"strings"
)

type Block struct {
	*hclsyntax.Block
	WriteBlock   *hclwrite.Block
	Count        *Attribute
	ForEach      *Attribute
	Attributes   map[string]*Attribute
	NestedBlocks NestedBlocks
	Type         string
	Labels       []string
	Address      string
}

func NewBlock(rb *hclsyntax.Block, wb *hclwrite.Block) *Block {
	b := &Block{
		Type:       rb.Type,
		Labels:     rb.Labels,
		Address:    strings.Join(append([]string{rb.Type}, rb.Labels...), "."),
		Block:      rb,
		WriteBlock: wb,
	}
	if countAttr, ok := rb.Body.Attributes["count"]; ok {
		b.Count = NewAttribute("count", countAttr, wb.Body().GetAttribute("count"))
	}
	if forEachAttr, ok := rb.Body.Attributes["for_each"]; ok {
		b.ForEach = NewAttribute("for_each", forEachAttr, wb.Body().GetAttribute("for_each"))
	}
	b.Attributes = attributes(rb.Body, wb.Body())
	b.NestedBlocks = nestedBlocks(rb.Body, wb.Body())
	return b
}

func (b *Block) EvalContext() cty.Value {
	v := map[string]cty.Value{}
	for n, a := range b.Attributes {
		v[n] = cty.StringVal(a.String())
	}
	if b.Count != nil {
		v["count"] = cty.StringVal(b.Count.String())
	}
	if b.ForEach != nil {
		v["for_each"] = cty.StringVal(b.ForEach.String())
	}
	for k, values := range b.NestedBlocks.Values() {
		v[k] = values
	}
	return cty.ObjectVal(v)
}

func attributes(rb *hclsyntax.Body, wb *hclwrite.Body) map[string]*Attribute {
	attributes := rb.Attributes
	r := make(map[string]*Attribute, len(attributes))
	for name, attribute := range attributes {
		r[name] = NewAttribute(name, attribute, wb.GetAttribute(name))
	}
	return r
}

func nestedBlocks(rb *hclsyntax.Body, wb *hclwrite.Body) NestedBlocks {
	blocks := rb.Blocks
	r := make(map[string][]*NestedBlock)
	for i, block := range blocks {
		nb := NewNestedBlock(block, wb.Blocks()[i])
		r[nb.Type] = append(r[nb.Type], nb)
	}
	for _, v := range r {
		sort.Slice(v, func(i, j int) bool {
			return v[i].Block.Range().Start.Line < v[j].Block.Range().Start.Line
		})
	}
	return r
}