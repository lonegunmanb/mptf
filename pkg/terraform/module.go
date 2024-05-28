package terraform

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/afero"
	"path/filepath"
	"strings"
	"sync"
)

var Fs = afero.NewOsFs()
var wantedTypes = map[string]func(module *Module) *[]*RootBlock{
	"resource": func(m *Module) *[]*RootBlock {
		return &m.ResourceBlocks
	},
	"data": func(m *Module) *[]*RootBlock {
		return &m.DataBlocks
	},
	"module": func(m *Module) *[]*RootBlock {
		return &m.ModuleBlocks
	},
}

type Module struct {
	Dir            string
	AbsDir         string
	writeFiles     map[string]*hclwrite.File
	lock           *sync.Mutex
	ResourceBlocks []*RootBlock
	DataBlocks     []*RootBlock
	ModuleBlocks   []*RootBlock
	Key            string
	Source         string
	Version        string
}

func (m *Module) loadConfig(cfg, filename string) error {
	writeFile, diag := hclwrite.ParseConfig([]byte(cfg), filename, hcl.InitialPos)
	if diag.HasErrors() {
		return diag
	}
	readFile, diag := hclsyntax.ParseConfig([]byte(cfg), filename, hcl.InitialPos)
	if diag.HasErrors() {
		return diag
	}
	m.writeFiles[filename] = writeFile
	readBlocks := readFile.Body.(*hclsyntax.Body).Blocks
	writeBlocks := writeFile.Body().Blocks()
	for i, rb := range readBlocks {
		getter, want := wantedTypes[rb.Type]
		if !want {
			continue
		}
		hclBlock := NewBlock(m, rb, writeBlocks[i])
		blocks := getter(m)
		*blocks = append(*blocks, hclBlock)
	}
	return nil
}

type TerraformModuleRef struct {
	Key     string `json:"Key"`
	Source  string `json:"Source"`
	Dir     string `json:"Dir"`
	AbsDir  string
	Version string `json:"Version"`
}

func LoadModule(mr TerraformModuleRef) (*Module, error) {
	files, err := afero.ReadDir(Fs, mr.AbsDir)
	if err != nil {
		return nil, err
	}
	m := &Module{
		Dir:        mr.Dir,
		AbsDir:     mr.AbsDir,
		writeFiles: make(map[string]*hclwrite.File),
		lock:       &sync.Mutex{},
		Key:        mr.Key,
		Source:     mr.Source,
		Version:    mr.Version,
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !strings.HasSuffix(f.Name(), ".tf") {
			continue
		}
		n := filepath.Join(mr.AbsDir, f.Name())
		content, err := afero.ReadFile(Fs, n)
		if err != nil {
			return nil, err
		}
		if err = m.loadConfig(string(content), f.Name()); err != nil {
			return nil, err
		}
	}

	return m, err
}

func (m *Module) SaveToDisk() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	for fn, wf := range m.writeFiles {
		content := wf.Bytes()
		err := afero.WriteFile(Fs, filepath.Join(m.Dir, fn), hclwrite.Format(content), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Module) AddBlock(fileName string, block *hclwrite.Block) {
	func() {
		m.lock.Lock()
		defer m.lock.Unlock()
		if _, ok := m.writeFiles[fileName]; !ok {
			m.writeFiles[fileName] = hclwrite.NewFile()
		}
	}()
	writeFile := m.writeFiles[fileName]
	lock.Lock(fileName)
	defer lock.Unlock(fileName)
	tokens := writeFile.Body().BuildTokens(nil)
	if tokens[len(tokens)-1].Type != hclsyntax.TokenNewline {
		writeFile.Body().AppendNewline()
	}
	writeFile.Body().AppendBlock(block)
	writeFile.Body().AppendNewline()
}
