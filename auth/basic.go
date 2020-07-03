package auth

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"
	"strings"
	"time"

	"github.com/9seconds/httransform/v2/layers"
	"github.com/PumpkinSeed/errors"
	"zvelo.io/ttlru"
)

const (
	basicAuthCacheFor            = time.Hour
	basicAuthCacheSizeMultiplier = 2
)

var (
	ErrBasicAuthMalformed = errors.Wrap(errors.New("malformed header"), ErrAuth)
	ErrBasicAuthScheme    = errors.Wrap(errors.New("incorrect scheme"), ErrAuth)
	ErrBasicAuthPayload   = errors.Wrap(errors.New("incorrect payload"), ErrAuth)
	ErrBasicAuthDelimiter = errors.Wrap(errors.New("incorrect delimiter"), ErrAuth)
	ErrBasicAuthNoUser    = errors.Wrap(errors.New("no such user"), ErrAuth)
)

type basicAuthResult struct {
	reply interface{}
	err   error
}

type basicAuthUserInfo struct {
	user     []byte
	password []byte
}

func (u *basicAuthUserInfo) OK(user, password []byte) bool {
	userNum := subtle.ConstantTimeCompare(u.user, user)
	passNum := subtle.ConstantTimeCompare(u.password, password)

	return userNum+passNum == 2
}

type basicAuth struct {
	cache ttlru.Cache
	infos []basicAuthUserInfo
}

func (b *basicAuth) Auth(ctx *layers.LayerContext) (bool, interface{}, error) {
	header := ctx.RequestHeaders.Get("proxy-authorization")

	if header == nil {
		return false, nil, nil
	}

	if item, ok := b.cache.Get(header.Value); ok {
		reply := item.(*basicAuthResult)
		return true, reply.reply, reply.err
	}

	resp := b.doAuth(header.Value)
	b.cache.Set(header.Value, &resp)

	return true, resp.reply, resp.err
}

func (b *basicAuth) doAuth(text string) basicAuthResult {
	pos := strings.IndexByte(text, ' ')
	if pos < 0 {
		return basicAuthResult{
			err: ErrBasicAuthMalformed,
		}
	}

	if !strings.EqualFold(text[:pos], "Basic") {
		return basicAuthResult{
			err: ErrBasicAuthScheme,
		}
	}

	for pos < len(text) && (text[pos] == ' ' || text[pos] == '\t') {
		pos++
	}

	decoded, err := base64.StdEncoding.DecodeString(text[pos:])
	if err != nil {
		return basicAuthResult{
			err: errors.Wrap(err, ErrBasicAuthPayload),
		}
	}

	pos = bytes.IndexByte(decoded, ':')
	if pos < 0 {
		return basicAuthResult{
			err: ErrBasicAuthDelimiter,
		}
	}

	found := false
	for idx := range b.infos {
		found = b.infos[idx].OK(decoded[:pos], decoded[pos+1:]) || found
	}

	if found {
		return basicAuthResult{
			reply: string(decoded[:pos]),
		}
	}

	return basicAuthResult{
		err: ErrBasicAuthNoUser,
	}
}

func NewBasicAuth(userPasswords map[string]string) Auth {
	userInfos := make([]basicAuthUserInfo, len(userPasswords))
	idx := 0

	for k, v := range userPasswords {
		userInfos[idx].user = []byte(k)
		userInfos[idx].password = []byte(v)
		idx++
	}

	return &basicAuth{
		cache: ttlru.New(basicAuthCacheSizeMultiplier*len(userPasswords),
			ttlru.WithTTL(basicAuthCacheFor)),
		infos: userInfos,
	}
}
