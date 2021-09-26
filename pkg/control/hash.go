package control

import (
	"encoding/binary"
	"errors"
	"hash/crc64"
)

func HashCRC64(content []byte) ([]byte, error) {
	if len(content) <= 0 {
		return nil, errors.New("crc64: there is no content to checksum")
	}

	p := make([]byte, 8)
	checksum := crc64.Checksum(content, crc64.MakeTable(crc64.ECMA))
	binary.BigEndian.PutUint64(p, checksum)
	return p, nil
}
