package request

import (
	"context"
	"io"
	"net/http"

	"github.com/MovieStoreGuy/versionist/pkg/netrc"
)

type (
	Factory interface {
		NewRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error)
	}

	FactoryFunc func(req *http.Request) *http.Request

	factory []FactoryFunc
)

var (
	_ Factory = (*factory)(nil)
)

func NewRequestFactory(opts ...FactoryFunc) Factory {
	return append(factory{}, opts...)
}

func WithNetrcAuthentication(machines netrc.Machines) FactoryFunc {
	return func(req *http.Request) *http.Request {
		if detail, ok := machines.GetMachineDetails(req.Host); ok {
			req.SetBasicAuth(detail.Login, detail.Password)
		}
		return req
	}
}
func WithHeaders(headers http.Header) FactoryFunc {
	return func(req *http.Request) *http.Request {
		for k, v := range headers {
			req.Header[k] = v
		}
		return req
	}
}

func (f factory) NewRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	for _, fn := range f {
		req = fn(req)
	}
	return req, nil
}
