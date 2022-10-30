package manifest

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockGoproxy struct{}

func (mockGoproxy) GetLatest(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"uber.org/zap": "v1.90.0",
	}, nil
}

func TestLoadingManifest(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		scenario string
		path     string
		manifest *Manifest
		err      error
	}{
		{
			scenario: "simple manifest",
			path:     "testdata/complete.yml",
			manifest: &Manifest{
				GoVersion: "1.19",
				Projects: []*Project{
					{
						Package: "uber.org/zap",
						Version: "v1.90.0",
						Match: []Matcher{
							matchString("uber.org/zap"),
						},
					},
					{
						Package: "github.com/open-telemetry/opentelemetry-collector",
						Version: "latest",
						Match: []Matcher{
							matchString("github.com/open-telemetry/opentelemetry-collector"),
							regexp.MustCompile("^github.com/open-telemetry/opentelemetry-collector/(.*)$"),
						},
					},
				},
			},
			err: nil,
		},
	} {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			m, err := ReadManifest(context.Background(), tc.path,
				WithGoProxyClient(mockGoproxy{}),
			)
			assert.ErrorIs(t, err, tc.err, "Must match the expected error")
			assert.EqualValues(t, tc.manifest.GoVersion, m.GoVersion, "Must match the expected value")
			assert.EqualValues(t, tc.manifest.Projects, m.Projects, "Must match the expected value")
		})
	}
}

func TestCheckingProject(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{
		Projects: []*Project{
			{
				Package: "uber.org/zap",
				Version: "latest",
				Match: []Matcher{
					matchString("uber.org/zap"),
				},
			},
			{
				Package: "github.com/open-telemetry/opentelemetry-collector",
				Version: "latest",
				Match: []Matcher{
					matchString("github.com/open-telemetry/opentelemetry-collector"),
					regexp.MustCompile("^github.com/open-telemetry/opentelemetry-collector/(.*)$$"),
				},
			},
		},
	}

	for _, tc := range []struct {
		repo    string
		version string
		matched bool
	}{
		{
			repo:    "uber.org/zap/zaptest",
			version: defaultVersion,
			matched: false,
		},
		{
			repo:    "github.com/open-telemetry/opentelemetry-collector",
			version: "latest",
			matched: true,
		},
		{
			repo:    "github.com/open-telemetry/opentelemetry-collector/pdata/pmetric",
			version: "latest",
			matched: true,
		},
	} {
		tc := tc
		t.Run(tc.repo, func(t *testing.T) {
			t.Parallel()

			v, ok := manifest.CheckProject(tc.repo)
			assert.Equal(t, tc.version, v)
			assert.Equal(t, tc.matched, ok)
		})
	}

}
