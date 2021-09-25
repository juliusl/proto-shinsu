package control

import (
	"net/http"
)

type httpsResponseDescriptor struct {
}

func (httpsResponseDescriptor) Header() http.Header {
	return nil
}

func (httpsResponseDescriptor) Write(p []byte) (int, error) {
	return 0, nil
}

func (httpsResponseDescriptor) WriteHeader(statusCode int) {
}
