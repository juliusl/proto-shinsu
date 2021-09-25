package control

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/juliusl/shinsu/pkg/channel"
)

type (
	StreamDescriptor struct {
		state channel.TransientDescriptor
	}

	StringAddress struct {
		protocol  string
		host      string
		root      string
		namespace string
		term      string
		reference string
		method    string
	}
)

func (s StringAddress) Location() (*url.URL, error) {
	return url.Parse(fmt.Sprintf("%s://%s/%s/%s/%s/%s", s.protocol, s.host, s.root, s.namespace, s.term, s.reference))
}

func (s StringAddress) Request(ctx context.Context, body io.Reader) (*http.Request, error) {
	loc, err := s.Location()
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, s.method, loc.String(), body)
}

func (s StringAddress) HashLen64() ([]byte, error) {
	return nil, nil
}
