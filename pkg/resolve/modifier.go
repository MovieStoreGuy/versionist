package resolve

import (
	"io"
	"os"

	"go.uber.org/zap"
	"golang.org/x/mod/modfile"

	"github.com/MovieStoreGuy/versionist/pkg/internal/filewalk"
	"github.com/MovieStoreGuy/versionist/pkg/manifest"
)

const (
	ModFilename = "*go.mod"
	ModComment  = "Modified by versionist"
)

type Modifier struct {
	bom  *manifest.Manifest
	root string
	log  *zap.Logger
}

type ModifierOption func(m *Modifier)

func WithLogger(log *zap.Logger) ModifierOption {
	return func(m *Modifier) {
		m.log = log
	}
}

func NewModifier(root string, bom *manifest.Manifest, opts ...ModifierOption) Modifier {
	m := Modifier{root: root, bom: bom, log: zap.NewNop()}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

func (m *Modifier) Update() error {
	walked, err := filewalk.NewWalkedFS(m.root, ModFilename)
	if err != nil {
		return err
	}
	return walked.Range(func(name string, f *os.File) error {
		m.log.Info("Reading go mod file", zap.String("path", name))

		content, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		mod, err := modfile.Parse(name, content, nil)
		if err != nil {
			return err
		}

		modified := false
		if mod.Go.Version != m.bom.GoVersion {
			mod.Go.Version = m.bom.GoVersion
			modified = true
		}

		for _, req := range mod.Require {
			if req.Indirect {
				continue
			}
			if ver, update := m.bom.CheckProject(req.Mod.Path); update {
				modified = true
				req.Mod.Version = ver
			}
		}

		if !modified {
			m.log.Info("No modifications", zap.String("path", name))
			return nil
		}
		mod.AddComment(ModComment)

		data, err := mod.Format()
		if err != nil {
			return err
		}

		m.log.Info("Rewritting go mod file", zap.String("path", name))
		if err := f.Truncate(0); err != nil {
			return err
		}
		_, err = f.Write(data)
		return err
	})
}
