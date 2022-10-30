package manifest

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"

	"github.com/MovieStoreGuy/versionist/pkg/goproxy"
)

const (
	defaultVersion = "v0.0.0"

	resolvePackage = "package"
	resolveRegex   = "regexp:"
)

var (
	ErrInvalidMatch = errors.New("invalid match")
)

type (
	// Manifest describes the list of projects that
	// should be matched and resolve to configured version
	Manifest struct {
		goproxy goproxy.Client `yaml:"-"`

		GoVersion string     `yaml:"go_version"`
		Projects  []*Project `yaml:"projects"`
	}

	ManifestOption func(m *Manifest)

	// Project defines a a package with a version with a set of
	// identifiers
	Project struct {
		// Name is the complete name for the repo being referenced
		Package string `yaml:"package"`
		// Version references the version to pin any matched projects to.
		Version string `yaml:"version"`
		// Match defines a set of expressions that are used to see
		// if a project matches this definition.
		Match []Matcher `yaml:"match"`
	}

	projectYAML struct {
		Package string   `yaml:"package"`
		Version string   `yaml:"version"`
		Match   []string `yaml:"match"`
	}
)

var (
	_ yaml.Unmarshaler = (*Project)(nil)
)

func WithGoProxyClient(c goproxy.Client) ManifestOption {
	return func(m *Manifest) {
		m.goproxy = c
	}
}

// ReadManifest will load a yaml manifest from disk and
// have it ready to be consumed, any issues trying to decode or read
// will be returned as an error.
func ReadManifest(ctx context.Context, pathname string, opts ...ManifestOption) (*Manifest, error) {
	f, err := os.Open(pathname)
	if err != nil {
		return nil, err
	}
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)

	manifest := &Manifest{
		goproxy: goproxy.NewClient(),
	}

	for _, opt := range opts {
		opt(manifest)
	}

	err = multierr.Combine(
		dec.Decode(manifest),
		f.Close(),
		manifest.resolveVersions(ctx),
	)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

func (m *Manifest) CheckProject(name string) (version string, matched bool) {
	for _, p := range m.Projects {
		if p.Check(name) {
			return p.Version, true
		}
	}
	return defaultVersion, false
}

func (m *Manifest) resolveVersions(ctx context.Context) error {
	packages := make([]string, 0, len(m.Projects))
	for _, p := range m.Projects {
		switch p.Version {
		case "latest":
			packages = append(packages, p.Package)
		}
	}
	mappings, err := m.goproxy.GetLatest(ctx, packages...)
	if err != nil {
		return err
	}
	for _, p := range m.Projects {
		if v, ok := mappings[p.Package]; ok {
			p.Version = v
		}
	}
	return nil
}

func (p *Project) Check(name string) bool {
	for _, m := range p.Match {
		if m.MatchString(name) {
			return true
		}
	}
	return false
}

func (p *Project) UnmarshalYAML(node *yaml.Node) error {
	val := projectYAML{}
	if err := node.Decode(&val); err != nil {
		return err
	}

	p.Package, p.Version = val.Package, val.Version
	p.Match = append(p.Match, matchString(p.Package))
	for _, v := range val.Match {
		switch {
		case strings.HasPrefix(v, resolveRegex):
			reg, err := regexp.Compile(strings.TrimPrefix(v, resolveRegex))
			if err != nil {
				return err
			}
			p.Match = append(p.Match, reg)
		default:
			return fmt.Errorf("unknown match %s: %w", v, ErrInvalidMatch)
		}
	}

	return nil
}
