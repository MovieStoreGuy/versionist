package goproxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode"

	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/MovieStoreGuy/versionist/pkg/request"
)

type (
	Client interface {
		GetLatest(ctx context.Context, projects ...string) (mappings map[string]string, err error)
	}

	ClientOptionFunc func(proxy *goproxy)

	// Resolver defines an abstract for loading the go proxy URLs.
	Resolver interface {
		ResolveURLs() []url.URL
	}

	// ResolverFunc allows a function to be used
	// as the GoProxies interface value
	ResolverFunc func() []url.URL

	goproxy struct {
		net *http.Client
		log *zap.Logger

		reqfact request.Factory
		proxies Resolver
	}

	proxyInfo struct {
		Version string    `json:"version"`
		Time    time.Time `json:"time,omitempty"`
	}
)

func GoProxiesFromEnvironment() Resolver {
	reg := regexp.MustCompile("[,|]")
	return ResolverFunc(func() (urls []url.URL) {
		vals, set := os.LookupEnv("GOPROXY")
		if !set {
			vals = "https://proxy.golang.org,direct"
		}
		for _, v := range reg.Split(vals, -1) {
			switch v {
			case "off":
				return []url.URL{}
			case "direct":
				continue
			default:
				u, err := url.Parse(v)
				if err != nil {
					continue
				}
				if !strings.HasPrefix(u.Scheme, "http") || u.Host == "" {
					continue
				}
				urls = append(urls, *u)
			}
		}
		return urls
	})
}

func caseEncoder(text string) string {
	modified := make([]rune, 0, len(text))
	for _, r := range text {
		switch unicode.IsUpper(r) {
		case true:
			modified, r = append(modified, '!'), unicode.ToLower(r)
			fallthrough
		default:
			modified = append(modified, r)
		}
	}
	return string(modified)
}

func WithGoProxyLogger(log *zap.Logger) ClientOptionFunc {
	return func(proxy *goproxy) {
		proxy.log = log
	}
}

func WithGoProxyHTTPClient(c *http.Client) ClientOptionFunc {
	return func(proxy *goproxy) {
		proxy.net = c
	}
}

func WithRequestFactory(rf request.Factory) ClientOptionFunc {
	return func(proxy *goproxy) {
		proxy.reqfact = rf
	}
}

func WithGoProxyProxies(proxies Resolver) ClientOptionFunc {
	return func(proxy *goproxy) {
		proxy.proxies = proxies
	}
}

func NewClient(opts ...ClientOptionFunc) Client {
	proxy := &goproxy{
		net:     http.DefaultClient,
		log:     zap.NewNop(),
		reqfact: request.NewRequestFactory(),
		proxies: GoProxiesFromEnvironment(),
	}

	for _, opt := range opts {
		opt(proxy)
	}

	return proxy
}

func (gp *goproxy) GetLatest(ctx context.Context, projects ...string) (mappings map[string]string, errs error) {
	mappings = make(map[string]string, len(projects))
	for _, project := range projects {
		for _, u := range gp.proxies.ResolveURLs() {
			if _, ok := mappings[project]; ok {
				gp.log.Info("Already resolved project version", zap.String("project", project))
				continue
			}
			u.Path = path.Join(u.Path, caseEncoder(project), "@latest")
			req, err := gp.reqfact.NewRequest(ctx, http.MethodGet, u.String(), http.NoBody)
			if err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
			resp, err := gp.net.Do(req)
			if err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				gp.log.Error("Invalid status code", zap.Int("status-code", resp.StatusCode))
				errs = multierr.Append(errs, resp.Body.Close())
				continue
			}
			var info proxyInfo
			if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
				errs = multierr.Append(errs, err)
				errs = multierr.Append(errs, resp.Body.Close())
				continue
			}
			errs = multierr.Append(errs, resp.Body.Close())
			mappings[project] = info.Version
		}
	}
	return mappings, errs
}

func (fn ResolverFunc) ResolveURLs() []url.URL {
	return fn()
}
