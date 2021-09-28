package control

import (
	"errors"
	"fmt"
	"net/http"
)

func NewTransport() *NodeTransport {
	n := &NodeTransport{api: make(map[string]*StreamDescriptor)}
	n.RegisterProtocol("api", n)
	n.RegisterProtocol("https", n)
	n.RegisterProtocol("http", n)
	n.RegisterProtocol("cache", n)
	n.RegisterProtocol("file", n)
	return n
}

type NodeTransport struct {
	api         map[string]*StreamDescriptor
	defaultHTTP *http.Client
	http.Transport
}

func (n *NodeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.URL.Scheme {
	case "https", "http":
		return n.defaultHTTP.Do(req)
	case "api":
		key := req.URL.Opaque
		api, ok := n.api[key]
		if !ok {
			return nil, errors.New("unknown api")
		}

		cache := req.URL
		cache.Scheme = "cache"
		api.SetLocation(cache)
		trans, err := api.ShouldTransition(req.Context())
		if err != nil {
			return nil, err
		}

		offset, expected, progress, err := trans.Position()
		if err != nil {
			return nil, err
		}

		position := fmt.Sprintf("%d, %d, %v", offset, expected, progress)

		return &http.Response{
			Status:     position,
			Header:     req.Header,
			StatusCode: http.StatusOK,
		}, nil
	case "cache", "file":
	}
	return nil, nil
}
