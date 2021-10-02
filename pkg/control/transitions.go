package control

import (
	"context"
	"io"
	"net/url"

	"github.com/juliusl/shinsu/pkg/channel"
)

func WriteCache(node *Node, mediatype string) (channel.TransitionFunction, error) {
	client, err := node.GetClient()
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, url *url.URL, reader io.Reader) (*channel.StableDescriptor, error) {
		url.Scheme = "cache"

		resp, err := client.Post(url.String(), mediatype, reader)
		if err != nil {
			return nil, err
		}

		loc, err := resp.Location()
		if err != nil {
			return nil, err
		}

		return node.transport.Source(ctx, loc)
	}, nil
}
