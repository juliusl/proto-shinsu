package control

import (
	"errors"
	"fmt"
	"net/http"
)

type Node struct {
	transport *NodeTransport
}

func (n *Node) AddAPI(root, term string, streamDesc *StreamDescriptor) {
	n.transport.api[fmt.Sprintf("%s/%s", root, term)] = streamDesc
}

func (n *Node) GetClient() *http.Client {
	return &http.Client{Transport: n.transport}
}

type NodeTransport struct {
	api map[string]*StreamDescriptor
	http.Transport
}

func CreateNode() (*Node, error) {
	n := &NodeTransport{api: make(map[string]*StreamDescriptor)}
	n.RegisterProtocol("api", n)
	n.RegisterProtocol("https", n)
	n.RegisterProtocol("http", n)
	n.RegisterProtocol("cache", n)
	n.RegisterProtocol("file", n)
	return &Node{transport: n}, nil
}

func (n *NodeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.URL.Scheme {
	case "https", "http":
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
