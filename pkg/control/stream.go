package control

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/juliusl/shinsu/pkg/channel"
)

// CreateStreamDescriptor is a function that returns an instance of a State struct
func CreateStreamDescriptor(client *http.Client, resolve func() (*Address, *url.URL, error), expectedHash []byte, hash func([]byte) ([]byte, error)) (*StreamDescriptor, error) {
	add, loc, err := resolve()
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

	sdesc := &StreamDescriptor{
		accepts: s.mediatype,
		state:   s,
		address: add,
	}

	return sdesc, nil
}

type StreamDescriptor struct {
	accepts string
	address *Address
	state   *State
	channel.TransientDescriptor
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
