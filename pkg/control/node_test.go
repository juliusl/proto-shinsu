package control

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/juliusl/shinsu/pkg/channel"
)

func TestNode(t *testing.T) {
	tr, err := NewTransport("localhost", time.Hour*3)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	a := EmptyAddress()
	a, err = a.FromString("v2:///#blobs/uploads")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	api, err := a.APIRoot()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if api.String() != "api://empty://?label=host/v2/blobs/uploads" {
		t.Error("Unexpected api root", api.String())
		t.Fail()
	}

	n, err := CreateNode(a, tr)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if n.address.host != "localhost" {
		t.Error("Unexpected host")
		t.Fail()
	}

	u, err := n.address.HTTPSRoot()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if u.String() != "https://localhost/v2/empty://?label=namespace/blobs/uploads/empty://?label=reference" {
		t.Error("Unexpected https root", u.String())
		t.Fail()
	}

	u, err = url.Parse("ref://v1@localhost/library/ubuntu")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	n.SetCookies(u, nil)

	u, err = n.address.HTTPRoot()
	if err != nil {
		t.Error()
		t.Fail()
	}

	if u.String() != "http://localhost/v2/library/ubuntu/blobs/uploads/v1" {
		t.Error("unexpected http root", u.String())
		t.Fail()
	}
}

func TestNodeAPI(t *testing.T) {
	tr, err := NewTransport("test-transport", time.Hour*3)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	add := &Address{}
	add, err = add.SetHost("test-transport")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	add, err = add.SetRoot("v2")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	add, err = add.SetTerm("blobs/uploads")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	n, err := CreateNode(add, tr)
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

		c, err := n.GetClient()
		if err != nil {
			return nil, err
		}

		url.Scheme = "cache"
		resp, err := c.Post(url.String(), "test+test", reader)
		if err != nil {
			return nil, err
		}

		loc, err := resp.Location()
		if err != nil {
			return nil, err
		}

		return tr.Source(ctx, loc)
	})
	sd.SetExpected(int64(len("test string hello")))
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
	sd.SetSource(fd.Source())

	au, err := add.API(http.MethodPost, http.MethodPatch, http.MethodPut)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	tr.AddAPI(add, sd)

	c, err := n.GetClient()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	resp, err := c.Get(au.String())
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if resp.StatusCode != http.StatusOK {
		t.Error("expected status OK", resp.StatusCode, resp.Status)
		t.Fail()
	}

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

	reload, err := CreateStreamDescriptor(
		n,
		func() (*Address, *url.URL, error) {
			u, err := add.CacheRoot()
			if err != nil {
				return nil, nil, err
			}

			return add, u, nil
		},
		[]byte("test"),
		func(b []byte) ([]byte, error) { return []byte("test"), nil })

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if reload.state == nil {
		t.Error("expected a state")
		t.Fail()
	}

	err = tr.AddAPI(add, reload)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	resp, err = c.Get(au.String())
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if resp.StatusCode != 200 {
		t.Error("unexpected status code")
		t.Fail()
	}
}
