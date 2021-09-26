package control

import (
	"context"
	"io/fs"
	"net/http"
	"os"
	"testing"
)

func TestFileDescriptor(t *testing.T) {
	ctx := context.Background()

	temp, err := os.Create(".temp")
	if err != nil {
		t.Error("could not create a temp test file")
		t.Fail()
	}

	temp.WriteString("test data")
	temp.Close()

	desc, err := NewFileDescriptor(ctx,
		"file.descriptor.test+txt",
		"test",
		func(p string) (fs.FileInfo, error) {
			return os.Stat(".temp")
		},
		func(p string) (*os.File, error) {
			return os.Open(".temp")
		})
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	mux := http.NewServeMux()
	mux.Handle("/test", desc)

	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	req.Header.Set("Accept", "file.descriptor.test+txt")

	h, pattern := mux.Handler(req)
	if pattern != "/test" {
		t.Error("unexpected pattern")
		t.Fail()
	}

	if h != desc {
		t.Error("unexpected file handler")
		t.Fail()
	}

	os.Remove(".temp")
}
