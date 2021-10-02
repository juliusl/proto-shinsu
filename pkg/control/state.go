package control

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

// State is an opaque type that can describe the state of a node in the control graph
type State struct {
	mediatype string
	offset    int64
	size      int64
	hash      []byte
}

func (s State) Start(expectedsize int64, expectedHash []byte) *State {
	s.mediatype = TransientMediaType
	s.offset = 0
	s.size = expectedsize
	s.hash = expectedHash

	return &s
}

func (s *State) Commit(mediatype string, hash []byte) error {
	if s.mediatype == TransientMediaType {
		checksum, err := s.CompareHash(hash)
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

func (s *State) Compare(other *State) (bool, error) {
	if s == other {
		return true, nil
	}

	if s.mediatype != other.mediatype {
		return false, nil
	}

	if s.offset != other.offset {
		return false, nil
	}

	if s.size != other.size {
		return false, nil
	}

	return s.CompareHash(other.hash)
}

func (s *State) CompareHash(hash []byte) (bool, error) {
	checksum, err := binary.ReadUvarint(bytes.NewReader(s.hash))
	if err != nil {
		return false, err
	}

	otherchecksum, err := binary.ReadUvarint(bytes.NewReader(hash))
	if err != nil {
		return false, err
	}

	return checksum == otherchecksum, nil
}

func (s *State) Store(writer io.WriteCloser) error {
	defer writer.Close()

	return json.NewEncoder(writer).Encode(&encodedAddress{
		Mediatype: s.mediatype,
		Offset:    s.offset,
		Size:      s.size,
		Hash:      s.hash,
	})
}

func (s *State) Load(reader io.ReadCloser) error {
	defer reader.Close()

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
	Offset    int64  `json:"offset"`
	Size      int64  `json:"size"`
	Hash      []byte `json:"hash"`
}

const TransientMediaType = "transient+unknown"
