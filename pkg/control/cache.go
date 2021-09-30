package control

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
	"time"

	"github.com/juliusl/shinsu/pkg/channel"
)

func NewCache(rootcacheDir, name string, ttl time.Duration) (*Cache, error) {
	cacheDir := path.Join(rootcacheDir, name)

	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{}
	transport.RegisterProtocol("cache", http.NewFileTransport(http.Dir(cacheDir)))

	return &Cache{
		name:      name,
		cachedir:  cacheDir,
		ttl:       ttl,
		nameindex: make(map[string]uint64),
		revindex:  make(map[uint64]cacheEntry),
		entries:   make(map[cacheEntry]time.Time),
		cached:    make(map[cacheEntry]func() (*os.File, error)),
		fs:        transport,
	}, nil
}

type Cache struct {
	name      string
	cachedir  string
	ttl       time.Duration
	nameindex map[string]uint64
	revindex  map[uint64]cacheEntry
	entries   map[cacheEntry]time.Time
	cached    map[cacheEntry]func() (*os.File, error)
	fs        http.RoundTripper

	sync.RWMutex
}

func (c *Cache) Name() string {
	return c.name
}

func (c *Cache) Source(ctx context.Context, url *url.URL) (*channel.StableDescriptor, error) {
	if url.Host != c.name {
		return nil, errors.New("unknown reference")
	}

	h, ok := c.nameindex[url.Path]
	if !ok {
		return nil, errors.New("unknown reference")
	}

	key, ok := c.revindex[h]
	if !ok {
		return nil, errors.New("unknown reference")
	}

	fd, err := NewFileDescriptor(ctx, key.mediatype, key.name, c.Stat, c.Open)
	if err != nil {
		return nil, err
	}

	return fd.channel, nil
}

func (c *Cache) Cache(ctx context.Context, name, mediatype string, reader io.Reader) (*channel.StableDescriptor, error) {
	c.Lock()
	defer c.Unlock()

	tempCachePath := path.Join(c.cachedir, fmt.Sprintf("%s-temp", name))

	err := os.MkdirAll(path.Dir(tempCachePath), 0755)
	if err != nil {
		return nil, err
	}

	cachedFile, err := os.Create(tempCachePath)
	if err != nil {
		return nil, err
	}

	defer cachedFile.Close()

	hash := fnv.New64a()
	tee := io.TeeReader(reader, hash)

	copied, err := io.Copy(cachedFile, tee)
	if err != nil {
		return nil, err
	}

	key := cacheEntry{
		name:      name,
		mediatype: mediatype,
		size:      copied,
		added:     time.Now(),
	}

	h := hash.Sum64()
	cachedFilepath := c.path(h)
	if err != nil {
		return nil, err
	}

	err = os.Rename(tempCachePath, cachedFilepath)
	if err != nil {
		return nil, err
	}

	c.entries[key] = key.added
	c.cached[key] = func() (*os.File, error) {
		return os.Open(cachedFilepath)
	}
	c.revindex[h] = key
	c.nameindex[key.name] = h

	fd, err := NewFileDescriptor(ctx, key.mediatype, key.name, c.Stat, c.Open)
	if err != nil {
		return nil, err
	}

	return fd.channel, nil
}

func (c *Cache) Lookup(hash uint64) (*os.File, error) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	key, err := c.entry(hash)
	if err != nil {
		return nil, err
	}

	f, ok := c.cached[key]
	if !ok {
		return nil, errors.New("invalid cache entry")
	}

	return f()
}

func (c *Cache) Stat(name string) (os.FileInfo, error) {
	h, ok := c.nameindex[name]
	if !ok {
		return nil, errors.New("does not exist")
	}
	finfo, err := os.Stat(c.path(h))
	if err != nil {
		return nil, err
	}

	if finfo.IsDir() {
		return nil, errors.New("entry is a directory")
	}

	entry, ok := c.revindex[h]
	if !ok {
		return nil, errors.New("cache entry does not exist")
	}

	return entry, nil
}

func (c *Cache) Open(name string) (*os.File, error) {
	h, ok := c.nameindex[name]
	if !ok {
		return nil, errors.New("does not exist")
	}

	entry, err := c.entry(h)
	if err != nil {
		return nil, err
	}

	get, ok := c.cached[entry]
	if !ok {
		return nil, errors.New("entry does not exist")
	}

	return get()
}

func (c *Cache) path(hash uint64) string {
	return path.Join(c.cachedir, fmt.Sprint(hash))
}

func (c *Cache) entry(hash uint64) (cacheEntry, error) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	key, ok := c.revindex[hash]
	if !ok {
		return cacheEntry{}, errors.New("cached file does not exist")
	}

	added, ok := c.entries[key]
	if !ok {
		return cacheEntry{}, errors.New("invalid cache entry")
	}

	if time.Now().After(added.Add(c.ttl)) {
		return cacheEntry{}, errors.New("expired cache entry")
	}

	return key, nil
}

type cacheEntry struct {
	name      string
	mediatype string
	size      int64
	added     time.Time
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
	return c
}
