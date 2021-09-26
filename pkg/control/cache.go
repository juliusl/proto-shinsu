package control

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/juliusl/shinsu/pkg/channel"
)

func FromCache(resp *http.Response) (*channel.StableDescriptor, error) {
	if resp.StatusCode != http.StatusAccepted {
		return nil, errors.New("cache rejected")
	}

	_, err := strconv.ParseUint(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return nil, errors.New("missing content length")
	}

	location, err := resp.Location()
	if err != nil {
		return nil, err
	}

	return cache{}.Entry(resp.Request.Context(), location.Path)
}

func NewCache(dir string) *http.Client {
	return &http.Client{
		Transport: http.NewFileTransport(http.Dir(dir)),
	}
}

type cache struct {
	ttl         time.Duration
	entries     map[cacheEntry]time.Time
	cached      map[cacheEntry]func() (*os.File, error)
	fstransport *http.Transport
}

func (c cache) Entry(ctx context.Context, path string) (*channel.StableDescriptor, error) {
	fd, err := NewFileDescriptor(
		ctx,
		"cached",
		path,
		c.Stat,
		c.Open)
	if err != nil {
		return nil, err
	}

	return fd.Channel(), nil
}

func (c cache) Stat(name string) (os.FileInfo, error) {
	return c.getEntry(name)
}

func (c cache) Open(name string) (*os.File, error) {
	entry, err := c.getEntry(name)
	if err != nil {
		return nil, err
	}

	get, ok := c.cached[entry]
	if !ok {
		return nil, errors.New("entry does not exist")
	}

	return get()
}

func (c *cache) getEntry(name string) (cacheEntry, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("cache:///%s", name), nil)
	if err != nil {
		return cacheEntry{}, err
	}

	resp, err := c.fstransport.RoundTrip(req)
	if err != nil {
		return cacheEntry{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return cacheEntry{}, errors.New("file does not exist")
	}

	key := cacheEntry{
		name: name,
		size: resp.ContentLength,
	}

	added, ok := c.entries[key]
	if !ok {
		return cacheEntry{}, errors.New("not cached")
	}

	if time.Since(added) > c.ttl {
		delete(c.entries, key)
		delete(c.cached, cacheEntry{name: name, size: resp.ContentLength, added: added})
		return cacheEntry{}, errors.New("entry is invalid")
	}

	return cacheEntry{name: name, size: resp.ContentLength, added: added}, nil
}

type cacheEntry struct {
	name  string
	size  int64
	added time.Time
}

var _ os.FileInfo = (*cacheEntry)(nil)

// base name of the file
func (c cacheEntry) Name() string {
	return c.name
}

// length in bytes for regular files; system-dependent for others
func (c cacheEntry) Size() int64 {
	return c.size
}

const CacheMode = os.ModeExclusive | os.ModeTemporary

func (c cacheEntry) Mode() os.FileMode {
	return CacheMode
}

// modification time
func (c cacheEntry) ModTime() time.Time {
	return c.added
}

// abbreviation for Mode().IsDir()
func (c cacheEntry) IsDir() bool {
	return false
}

// underlying data source (can return nil)
func (c cacheEntry) Sys() interface{} {
	return nil
}
