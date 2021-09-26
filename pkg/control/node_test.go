package control

import (
	"context"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	"github.com/juliusl/shinsu/pkg/channel"
)

func TestNodeAPI(t *testing.T) {
	n, err := CreateNode()
	if err != nil {
		t.Error(err.Error())
	}

	sd := &StreamDescriptor{}
	sd.SetTransition(func(ctx context.Context, url *url.URL, reader io.Reader) (*channel.StableDescriptor, error) {
		q := url.Query()
		seq := q["method"]
		if seq[0] != "POST" {
			t.Error("Unexpected method")
			t.Fail()
		}

		if seq[1] != "PATCH" {
			t.Error("Unexpected method")
			t.Fail()
		}

		if seq[2] != "PUT" {
			t.Error("Unexpected method")
			t.Fail()
		}
		url.RawQuery = ""

		resp, err := n.GetClient().Post(url.String(), "cached.test+test", reader)
		if err != nil {
			return nil, err
		}

		return FromCache(resp)
	})
	sd.SetExpected(len("test string hello"))
	ctx := context.Background()
	temp, err := os.Create("test")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	temp.WriteString("test string hello")
	temp.Close()

	fd, err := NewFileDescriptor(ctx, "+test", "test", StatOS, os.Open)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	sd.SetSource(fd.Channel())

	n.AddAPI("v2", "blobs/uploads", sd)

	n.GetClient().Get("api:v2/blobs/uploads?method=POST&method=PATCH&method=PUT")

	cached, err := sd.Source()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	cr, err := cached.Open()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	d, err := ioutil.ReadAll(cr)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if string(d) != "test string hello" {
		t.Error("unexpected cache value")
		t.Fail()
	}
}
