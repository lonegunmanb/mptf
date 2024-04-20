package terraform

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"sort"
	"strings"
	"sync"
)

var _ Block = new(RootBlock)
var _ Locakable = new(RootBlock)

var RootBlockReflectionInformation = func(v map[string]cty.Value, b *RootBlock) {
	v["mptf"] = cty.ObjectVal(map[string]cty.Value{
		"block_address": cty.StringVal(b.Address),
		"range": cty.ObjectVal(map[string]cty.Value{
			"file_name":    cty.StringVal(b.Range().Filename),
			"start_line":   cty.NumberIntVal(int64(b.Range().Start.Line)),
			"start_column": cty.NumberIntVal(int64(b.Range().Start.Column)),
			"end_line":     cty.NumberIntVal(int64(b.Range().End.Line)),
			"end_column":   cty.NumberIntVal(int64(b.Range().End.Column)),
		}),
	})
}

type RootBlock struct {
	lock *sync.Mutex
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

func (b *RootBlock) Lock() {
	b.lock.Lock()
}

func (b *RootBlock) Unlock() {
	b.lock.Unlock()
}

func (b *RootBlock) WriteBody() *hclwrite.Body {
	return b.WriteBlock.Body()
}

func (b *RootBlock) GetAttributes() map[string]*Attribute {
	return b.Attributes
}

func (b *RootBlock) GetNestedBlocks() NestedBlocks {
	return b.NestedBlocks
}

func NewBlock(rb *hclsyntax.Block, wb *hclwrite.Block) *RootBlock {
	b := &RootBlock{
		lock:       &sync.Mutex{},
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

func (b *RootBlock) EvalContext() cty.Value {
	v := map[string]cty.Value{}
	RootBlockReflectionInformation(v, b)
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