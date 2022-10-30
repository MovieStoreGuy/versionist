package goproxy

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestDefaultProxy(t *testing.T) {
	t.Parallel()

	proxy := NewClient(
		WithGoProxyLogger(zaptest.NewLogger(t)),
	)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	t.Log("Go Proxy value:", os.Getenv("GOPROXY"))

	mappings, err := proxy.GetLatest(ctx, "golang.org/x/mod")
	require.NoError(t, err, "Must to error when checking latest sdk")
	require.Len(t, mappings, 1, "Must have only one entry")
}
