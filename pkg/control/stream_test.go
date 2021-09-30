package control

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestCreateStreamDescriptor(t *testing.T) {
	a := &Address{
		root:      "v2",
		namespace: "test",
		term:      "address",
		reference: "TestFromAddress",
	}
	tr, err := NewTransport("localhost")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	n, err := CreateNode(a, tr)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	cl, err := n.GetClient()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	u, err := a.CacheRoot()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	resp, err := cl.Post(u.String(), "test+test", strings.NewReader("test content"))
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if resp.StatusCode != http.StatusAccepted {
		t.Error("expected the cache to accept the content")
		t.Fail()
	}

	h, err := HashCRC64([]byte("test content"))
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	sdesc, err := CreateStreamDescriptor(
		n,
		func() (*Address, *url.URL, error) {
			fr, err := a.CacheRoot()
			if err != nil {
				return nil, nil, err
			}

			return a, fr, nil
		},
		h,
		HashCRC64)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	state := sdesc.state
	if state == nil {
		t.Error("expected a state")
		t.Fail()
	}

	pr, pw := io.Pipe()
	go state.Store(pw)

	stored := make([]byte, 2000)
	read, err := pr.Read(stored)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if read <= 0 {
		t.Error("expected bytes to have been written")
	}

	if state == nil {
		t.Error("expected state to not be nil")
		t.Fail()
	}

	// Simulate reading this data later
	later := State{}.Start(int64(len("test content")), h)
	err = later.Load(ioutil.NopCloser(bytes.NewReader(stored)))
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}
