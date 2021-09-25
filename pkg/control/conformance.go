package control

import "net/http"

var _ http.ResponseWriter = (*fileResponseWriter)(nil)
