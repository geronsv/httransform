package auth

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"github.com/valyala/fasthttp"
)

type basicAuth struct {
	user         string
	expectedAuth []byte
}

func (b *basicAuth) Authenticate(ctx *fasthttp.RequestCtx) (string, error) {
	authError := ErrFailedAuth

	ctx.Request.Header.VisitAll(func(key, value []byte) {
		if bytes.EqualFold(key, []byte("Proxy-Authorization")) {
			authError = b.doAuth(value)
		}
	})

	if authError == nil {
        return b.user, nil
	}

    return "", authError
}

func (b *basicAuth) doAuth(header []byte) error {
	pos := bytes.IndexByte(header, ' ')
	if pos < 0 {
		return ErrMalformedHeaderValue
	}

	if !bytes.EqualFold(header[:pos], []byte("Basic")) {
		return fmt.Errorf("unsupported auth schema %s", string(header[:pos]))
	}

	for pos < len(header) && (header[pos] == ' ' || header[pos] == '\t') {
		pos++
	}

	if subtle.ConstantTimeCompare(header[pos:], b.expectedAuth) != 1 {
		return ErrFailedAuth
	}

	return nil
}

func NewBasicAuth(user, password string) Interface {
	userpassword := []byte(user + ":" + password)
	encoded := base64.StdEncoding.EncodeToString(userpassword)

	return &basicAuth{
		user:         user,
		expectedAuth: []byte(encoded),
	}
}
