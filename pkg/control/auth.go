package control

import (
	"context"
	"errors"
	"net/http"
)

type authKey = struct{}

var authrKey authKey = struct{}{}

type Authorizor = func(req *http.Request) error

func WithAuthorizor(ctx context.Context, authr Authorizor) context.Context {
	return context.WithValue(ctx, authrKey, authr)
}

func Authorize(ctx context.Context, req *http.Request) error {
	val := ctx.Value(authrKey)
	if val == nil {
		return errors.New("context has no authorizor")
	}

	authr, ok := val.(Authorizor)
	if !ok {
		return errors.New("invalid context")
	}

	err := authr(req)
	if err != nil {
		return err
	}

	return nil
}
