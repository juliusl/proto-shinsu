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
	host      string
	root      string
	namespace string
	term      string
	reference string
	sync.RWMutex
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

func (s *Address) HTTPSRoot() (*url.URL, error) {
	s.RLock()
	defer s.RUnlock()

	u, err := url.Parse(fmt.Sprintf("https://%s/%s/%s/%s/%s", s.host, s.root, s.namespace, s.term, s.reference))
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Address) HTTPRoot() (*url.URL, error) {
	s.RLock()
	defer s.RUnlock()

	u, err := url.Parse(fmt.Sprintf("http://%s/%s/%s/%s/%s", s.host, s.root, s.namespace, s.term, s.reference))
	if err != nil {
		return nil, err
	}

	return u, nil
}

const api_format = "api://%s/%s" // api://<root>/<term>
func (s *Address) APIRoot() (*url.URL, error) {
	s.RLock()
	defer s.RUnlock()

	return url.Parse(fmt.Sprintf(api_format, s.root, s.term))
}

const cache_format = "cache://%s/%s/%s/%s"

func (s *Address) CacheRoot(hash string) (*url.URL, error) {
	s.RLock()
	defer s.RUnlock()

	return url.Parse(fmt.Sprintf(cache_format, s.host, s.namespace, s.reference, hash))
}

const file_format = "file://%s/%s/%s"

func (s *Address) FileRoot() (*url.URL, error) {
	s.RLock()
	defer s.RUnlock()

	return url.Parse(fmt.Sprintf(file_format, s.host, s.namespace, s.reference))
}

func (s *Address) SecureRequest(ctx context.Context, body io.Reader) (*http.Request, error) {
	method := http.MethodGet

	if body != nil {
		method = http.MethodPost
	}

	loc, err := s.HTTPSRoot()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, loc.String(), body)
	if err != nil {
		return nil, err
	}

	err = Authorize(ctx, req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (s *Address) InsecureRequest(ctx context.Context, body io.Reader) (*http.Request, error) {
	method := http.MethodGet
	if body != nil {
		method = http.MethodPost
	}

	loc, err := s.HTTPRoot()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, loc.String(), body)
	if err != nil {
		return nil, err
	}

	return req, nil
}
