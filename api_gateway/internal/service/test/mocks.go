// api_gateway/internal/service/test/mocks.go
package test

import (
	"io"
	"net/http"

	"github.com/stretchr/testify/mock"
)

// CommonMockHTTPRoundTripper - общий мок для http.RoundTripper
type CommonMockHTTPRoundTripper struct {
	mock.Mock
}

func (m *CommonMockHTTPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// CommonMockReadCloser - общий мок для io.ReadCloser
type CommonMockReadCloser struct {
	io.Reader
}

func (m *CommonMockReadCloser) Close() error {
	return nil
}
