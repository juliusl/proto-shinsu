package control

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
)

// Address is an opaque type that can be formatted into a valid URI
type Address struct {
	protocol  string
	host      string
	root      string
	namespace string
	term      string
	reference string
	method    string
	sync.RWMutex
}

func (s *Address) SetProtocol(protocol string) (*Address, error) {
	s.Lock()
	defer s.Unlock()

	s.protocol = protocol
	return s, nil
}

func (s *Address) SetHost(host string) (*Address, error) {
	s.Lock()
	defer s.Unlock()

	s.host = host
	return s, nil
}

func (s *Address) SetRoot(root string) (*Address, error) {
	s.Lock()
	defer s.Unlock()

	s.root = root
	return s, nil
}

func (s *Address) SetNamespace(namespace string) (*Address, error) {
	s.Lock()
	defer s.Unlock()

	s.namespace = namespace
	return s, nil
}

func (s *Address) SetTerm(term string) (*Address, error) {
	s.Lock()
	defer s.Unlock()

	s.term = term
	return s, nil
}

func (s *Address) SetReference(reference string) (*Address, error) {
	s.Lock()
	defer s.Unlock()

	s.reference = reference
	return s, nil
}

func (s *Address) SetMethod(method string) (*Address, error) {
	s.Lock()
	defer s.Unlock()

	s.method = method
	return s, nil
}

func (s *Address) APILocation() (string, *url.URL, error) {
	s.RLock()
	defer s.RUnlock()

	u, err := url.Parse(fmt.Sprintf("%s://%s/%s/%s/%s/%s", s.protocol, s.host, s.root, s.namespace, s.term, s.reference))
	if err != nil {
		return "", nil, err
	}

	return s.method, u, nil
}

func (s *Address) Request(ctx context.Context, body io.Reader) (*http.Request, error) {
	method, loc, err := s.APILocation()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, loc.String(), body)
	if err != nil {
		return nil, err
	}

	return req, nil
}
