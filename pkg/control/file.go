package control

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/juliusl/shinsu/pkg/channel"
)

func NewFileDescriptor(ctx context.Context, mediatype, path string, stat func(string) (os.FileInfo, error), open func(name string) (*os.File, error)) (*FileDescriptor, error) {
	finfo, err := stat(path)
	if err != nil {
		return nil, err
	}

	desc := &FileDescriptor{}
	desc.path = path
	desc.mediaType = mediatype
	desc.size = finfo.Size()
	desc.modtime = finfo.ModTime()
	desc.open = open
	desc.stat = stat

	ch := channel.CreateStableDescriptor(
		ctx,
		int(desc.size),
		desc.Open,
		desc.Resume)

	ch, err = ch.AddHealthChecks(func() error {
		finfo, err := stat(path)
		if err != nil {
			return err
		}

		if finfo.IsDir() {
			return errors.New("path is a directory")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	desc.channel = ch

	return desc, nil
}

type FileDescriptor struct {
	mediaType string
	path      string
	size      int64
	modtime   time.Time
	open      func(name string) (*os.File, error)
	stat      func(string) (fs.FileInfo, error)
	channel   *channel.StableDescriptor
	http.Handler
}

func (f *FileDescriptor) Open() (io.Reader, error) {
	of, err := f.open(f.path)
	if err != nil {
		return nil, err
	}

	return of, nil
}

func (f *FileDescriptor) Resume() (io.Reader, error) {
	finfo, err := f.stat(f.path)
	if err != nil {
		return nil, err
	}

	if finfo.Size() != f.size {
		return nil, errors.New("file has changed")
	}

	if f.modtime.After(finfo.ModTime()) {
		return nil, errors.New("file has changed")
	}

	return f.Open()
}

func (f *FileDescriptor) Channel() *channel.StableDescriptor {
	return f.channel
}

func (f *FileDescriptor) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	accept := req.Header.Get("Accept")
	if accept == "" {
		writer.WriteHeader(404)
		return
	}

	if !strings.Contains(accept, f.mediaType) {
		writer.WriteHeader(404)
		return
	}

	readcloser, err := f.channel.Open()
	if err != nil {
		writer.WriteHeader(500)
		return
	}

	defer readcloser.Close()

	written, err := io.Copy(writer, readcloser)
	if err != nil {
		writer.WriteHeader(500)
		return
	}

	headers := writer.Header()
	headers.Set("Content-Length", fmt.Sprint(written))
	headers.Set("Content-Type", req.Header.Get("Content-Type"))

	writer.WriteHeader(200)
}
