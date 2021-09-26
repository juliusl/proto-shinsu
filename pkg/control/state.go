package control

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

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

// State is an opaque type that can describe the state of a node in the control graph
type State struct {
	mediatype string
	offset    int
	size      int
	hash      []byte
}

func (s State) Start(expectedsize int, expectedHash []byte) *State {
	s.mediatype = TransientMediaType
	s.offset = 0
	s.size = expectedsize
	s.hash = expectedHash

	return &s
}

func (s *State) Commit(mediatype string, hash []byte) error {
	if s.mediatype == TransientMediaType {
		checksum, err := s.Compare(hash)
		if err != nil {
			return err
		}

		if checksum {
			s.mediatype = mediatype
			return nil
		}
	}

	return errors.New("state cannot be recommitted, start a new instance instead")
}

func (s State) IsStable() error {
	if s.mediatype == TransientMediaType {
		return errors.New("state is in transit")
	}

	if s.size != s.offset {
		return errors.New("state is incomplete")
	}

	if len(s.hash) == 0 {
		return errors.New("state must have a non-zero hash")
	}

	return nil
}

func (s *State) Compare(hash []byte) (bool, error) {
	checksum, err := binary.ReadUvarint(bytes.NewReader(s.hash))
	if err != nil {
		return false, err
	}

	other, err := binary.ReadUvarint(bytes.NewReader(hash))
	if err != nil {
		return false, err
	}

	return checksum == other, nil
}

func (s *State) Store(writer io.WriteCloser) error {
	return json.NewEncoder(writer).Encode(&encodedAddress{
		Mediatype: s.mediatype,
		Offset:    s.offset,
		Size:      s.size,
		Hash:      s.hash,
	})
}

func (s *State) Load(reader io.ReadCloser) error {
	decoded := &encodedAddress{}
	err := json.NewDecoder(reader).Decode(decoded)
	if err != nil {
		return err
	}

	err = s.Commit(decoded.Mediatype, decoded.Hash)
	if err != nil {
		return err
	}

	s.offset = decoded.Offset
	s.size = decoded.Size
	return nil
}

type encodedAddress struct {
	Mediatype string `json:"mediatype"`
	Offset    int    `json:"offset"`
	Size      int    `json:"size"`
	Hash      []byte `json:"hash"`
}

const TransientMediaType = "transient+unknown"
