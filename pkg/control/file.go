package control

import (
	"net/http"

	"github.com/juliusl/shinsu/pkg/channel"
)

type FileDescriptor struct {
	channel.StableDescriptor
}

type fileResponseWriter struct {
}

func (fileResponseWriter) Header() http.Header {
	return nil
}

func (fileResponseWriter) Write(p []byte) (int, error) {
	return 0, nil
}

func (fileResponseWriter) WriteHeader(statusCode int) {
}
