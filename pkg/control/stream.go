package control

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/juliusl/shinsu/pkg/channel"
)

// TODO this should return a stream descriptor
// FromAddress is a function that returns an instance of a State struct
func FromAddress(client *http.Client, address *Address, hash func([]byte) ([]byte, error)) (*State, error) {
	method, url, err := address.APILocation()
	if err != nil {
		return nil, err
	}

	if method != http.MethodGet {
		return nil, errors.New("address does not have required method GET")
	}

	resp, err := client.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	s := &State{}

	s.mediatype = resp.Header.Get("Content-Type")
	size, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 32)
	if err != nil {
		return nil, err
	}
	s.size = int(size)

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	s.offset = len(content)
	h, err := hash(content)
	if err != nil {
		return nil, err
	}
	s.hash = h

	return s, nil
}

type StreamDescriptor struct {
	accepts string
	address *Address
	state   *State
	channel.TransientDescriptor
}

func (s *StreamDescriptor) SetAccepts(accepts string) {
	s.accepts = accepts
}

func (s *StreamDescriptor) SetState(state *State) {
	s.state = state
}

func (s *StreamDescriptor) SetAddress(address *Address) {
	s.address = address
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
