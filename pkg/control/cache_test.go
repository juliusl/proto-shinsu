package control

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func testfiles() (*os.File, error) {
	f, err := os.Create("testfile")
	if err != nil {
		return nil, err
	}

	f.WriteString("test file hello world")
	f.Close()

	return f, nil
}

func TestCache(t *testing.T) {
	rootCacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	c, err := NewCache(rootCacheDir, "test", time.Hour*10)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	ctx := context.Background()
	f, err := testfiles()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	t.Cleanup(func() {
		os.Remove(f.Name())
	})

	defer f.Close()

	opened, err := os.Open("testfile")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	desc, err := c.Cache(ctx, "testfile", "test.media.type+test", opened)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	read, err := desc.Open()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	bytes, err := ioutil.ReadAll(read)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if string(bytes) != "test file hello world" {
		t.Error("expected a value")
		t.Fail()
	}

}
