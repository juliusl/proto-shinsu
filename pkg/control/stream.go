package control

import (
	"context"
	"io"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"

	"github.com/juliusl/shinsu/pkg/channel"
)

// CreateStreamDescriptor is a function that returns an instance of a State struct
func CreateStreamDescriptor(node *Node, resolve func() (*Address, *url.URL, error), expectedHash []byte, hash func([]byte) ([]byte, error)) (*StreamDescriptor, error) {
	add, loc, err := resolve()
	if err != nil {
		return nil, err
	}

	client, err := node.GetClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.Get(loc.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	size, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 32)
	if err != nil {
		return nil, err
	}

	s := State{}.Start(size, expectedHash)
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	h, err := hash(content)
	if err != nil {
		return nil, err
	}

	contentType := resp.Header.Get("Content-Type")
	err = s.Commit(contentType, h)
	if err != nil {
		return nil, err
	}

	s.offset = int64(len(content))

	sdesc := &StreamDescriptor{
		accepts: s.mediatype,
		state:   s,
		address: add,
	}

	sdesc.SetTransition(
		func(ctx context.Context, url *url.URL, reader io.Reader) (*channel.StableDescriptor, error) {
			url.Scheme = "cache"

			resp, err := client.Post(url.String(), s.mediatype, reader)
			if err != nil {
				return nil, err
			}

			loc, err := resp.Location()
			if err != nil {
				return nil, err
			}

			return node.transport.Source(ctx, loc)
		})
	sdesc.SetExpected(size)

	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)
	sdesc.cancel = cancelFunc
	src, err := node.transport.Source(ctx, loc)
	if err != nil {
		return nil, err
	}
	sdesc.SetSource(src)

	return sdesc, nil
}

type StreamDescriptor struct {
	accepts string
	address *Address
	state   *State
	cancel  context.CancelFunc
	channel.TransientDescriptor
}

func (s *StreamDescriptor) Cancel() {
	s.cancel()
}

// Update is a function that updates the source for the stream descriptor
func (s *StreamDescriptor) Update(ctx context.Context, contentType string, checksum []byte) (*StreamDescriptor, error) {
	_, expected, pro, err := s.Position()
	if err != nil {
		return nil, err
	}

	if pro == 0.0 &&
		strings.Contains(s.accepts, contentType) ||
		strings.Contains(s.accepts, "*/*") {
		// source is unset so update can proceed

		return nil, nil
	}

	if pro == 1.0 {
		state := State{}.Start(expected, checksum)

		stable, err := s.Source()
		if err != nil {
			return nil, err
		}

		err = state.Commit(contentType, checksum)
		if err != nil {
			return nil, err
		}

		_, err = s.SetSource(stable)
		if err != nil {
			return nil, err
		}

	}

	return nil, nil
}
