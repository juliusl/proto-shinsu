package control

import (
	"testing"
)

func TestFromStart(t *testing.T) {
	s := State{}.Start(100, []byte{0, 1, 2, 3, 4, 5, 6})

	err := s.Commit("test.complete+empty", []byte{0, 1, 2, 3, 4, 5, 6})
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}
