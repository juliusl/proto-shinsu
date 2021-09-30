package control

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/juliusl/shinsu/pkg/channel"
)

func NewTransport(name string) (*NodeTransport, error) {
	n := &NodeTransport{api: make(map[string]*StreamDescriptor)}
	n.RegisterProtocol("api", n)
	n.RegisterProtocol("https", n)
	n.RegisterProtocol("http", n)
	n.RegisterProtocol("cache", n)
	n.RegisterProtocol("file", n)

	uc, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	c, err := NewCache(uc, name, time.Hour*3)
	if err != nil {
		return nil, err
	}
	n.cache = c

	return n, nil
}

type NodeTransport struct {
	cache *Cache
	api   map[string]*StreamDescriptor
	https *http.Client
	http  *http.Client
	http.Transport
}

func (n *NodeTransport) Source(ctx context.Context, url *url.URL) (*channel.StableDescriptor, error) {
	return n.cache.Source(ctx, url)
}

func (n *NodeTransport) SetHTTPClient(h *http.Client) {
	n.http = h
}

func (n *NodeTransport) SetHTTPSClient(h *http.Client) {
	n.https = h
}

func (n *NodeTransport) AddAPI(address *Address, streamDesc *StreamDescriptor) error {
	r, err := address.APIRoot()
	if err != nil {
		return err
	}

	n.api[r.Path] = streamDesc

	return nil
}

func (n *NodeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.URL.Scheme {
	case "http":
		return n.http.Do(req)
	case "https":
		return n.https.Do(req)
	case "api":
		key := req.URL.Path
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
	case "cache":
		if req.Method == http.MethodPost {
			if req.URL.Host != n.cache.name {
				return nil, errors.New("incorrect host")
			}

			_, err := n.cache.Cache(req.Context(), req.URL.Path, req.Header.Get("Content-Type"), req.Body)
			if err != nil {
				return nil, err
			}

			resp := &http.Response{
				StatusCode: http.StatusAccepted,
				Header:     make(http.Header),
			}
			resp.Header.Set("Content-Length", fmt.Sprint(req.ContentLength))
			resp.Header.Set("Content-Type", req.Header.Get("Content-Type"))
			resp.Header.Set("Location", fmt.Sprintf("cache://%s%s", n.cache.name, req.URL.Path))

			return resp, nil
		}
		fallthrough
	case "file":
		if req.Method != http.MethodGet {
			return nil, errors.New("node files are readonly")
		}

		finfo, err := n.cache.Stat(req.URL.Path)
		if err != nil {
			return nil, err
		}

		entry, ok := finfo.Sys().(cacheEntry)
		if !ok {
			return nil, errors.New("unexpected type")
		}

		f, err := n.cache.Open(req.URL.Path)
		if err != nil {
			return nil, err
		}

		resp := &http.Response{
			Body:   f,
			Header: make(http.Header),
		}

		resp.Header.Set("Content-Type", entry.mediatype)
		resp.Header.Set("Content-Length", fmt.Sprint(entry.size))
		resp.StatusCode = http.StatusOK

		return resp, nil
	}

	return nil, errors.New("protocol not implemented")
}
