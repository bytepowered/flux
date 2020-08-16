package webx

import (
	"encoding/base64"
	"github.com/bytepowered/flux"
	"net/http"
	"strconv"
	"strings"
)

const (
	basic                 = "basic"
	defaultRealm          = "Restricted"
	HeaderAuthorization   = "Authorization"
	HeaderWWWAuthenticate = "WWW-Authenticate"
)

type BasicAuthConfig struct {
	// Skipper 用于跳过某些请求
	Skipper func(ctx WebContext) bool
	// Validator 用于检查请求BasicAuth密钥的函数
	Validator func(string, string, WebContext) (bool, error)
	// Default value "Restricted".
	Realm string
}

// NewBasicAuthMiddleware 返回BaseAuth中间件。
func NewBasicAuthMiddleware(validator func(string, string, WebContext) (bool, error)) WebMiddleware {
	return NewBasicAuthMiddlewareWith(BasicAuthConfig{
		Skipper:   func(WebContext) bool { return false },
		Validator: validator,
		Realm:     defaultRealm,
	})
}

// NewBasicAuthMiddleware 返回BaseAuth中间件
func NewBasicAuthMiddlewareWith(config BasicAuthConfig) WebMiddleware {
	// 参考Echo.BasicAut的实现。
	// Defaults
	if config.Validator == nil {
		panic("webex: basic-auth middleware requires a validator function")
	}
	if config.Realm == "" {
		config.Realm = defaultRealm
	}
	return func(next WebRouteHandler) WebRouteHandler {
		return func(webc WebContext) error {
			// Skip
			if config.Skipper != nil && config.Skipper(webc) {
				return next(webc)
			}
			auth := webc.RequestHeader().Get(HeaderAuthorization)
			l := len(basic)
			if len(auth) > l+1 && strings.ToLower(auth[:l]) == basic {
				b, err := base64.StdEncoding.DecodeString(auth[l+1:])
				if err != nil {
					return err
				}
				cred := string(b)
				for i := 0; i < len(cred); i++ {
					if cred[i] == ':' {
						// Verify credentials
						valid, err := config.Validator(cred[:i], cred[i+1:], webc)
						if err != nil {
							return err
						} else if valid {
							return next(webc)
						}
						break
					}
				}
			}

			realm := defaultRealm
			if config.Realm != defaultRealm {
				realm = strconv.Quote(config.Realm)
			}
			// Need to return `401` for browsers to pop-up login box.
			webc.ResponseHeader().Set(HeaderWWWAuthenticate, basic+" realm="+realm)
			return &flux.StateError{
				StatusCode: http.StatusUnauthorized,
				ErrorCode:  "BASIC_AUTH:UNAUTHORIZED",
				Message:    "BASIC_AUTH:UNAUTHORIZED",
			}
		}
	}
}
