package control

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestCreateStreamDescriptor(t *testing.T) {
	transport := &http.Transport{}
	transport.RegisterProtocol("file", &testRoundTripper{})

	client := &http.Client{}
	client.Transport = transport

	sdesc, err := CreateStreamDescriptor(
		client,
		func() (*Address, *url.URL, error) {
			a := &Address{
				host:      "localhost",
				root:      "v2",
				namespace: "test",
				term:      "address",
				reference: "TestFromAddress",
			}

			fr, err := a.FileRoot()
			if err != nil {
				t.Error(err)
				t.Fail()
			}

			return a, fr, nil
		},
		[]byte{124, 183, 189, 253, 47, 166, 166, 154},
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
	later := State{}.Start(23, []byte{124, 183, 189, 253, 47, 166, 166, 154})
	err = later.Load(ioutil.NopCloser(bytes.NewReader(stored)))
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}

type testRoundTripper struct {
}

func (t *testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {

	resp := &http.Response{}
	resp.Body = ioutil.NopCloser(strings.NewReader("this is my cool content"))
	headers := &http.Header{}
	headers.Add("Content-Length", fmt.Sprint(len("this is my cool content")))
	headers.Add("Content-Type", "test.data+text")
	resp.Header = *headers
	resp.StatusCode = 200

	return resp, nil
}
