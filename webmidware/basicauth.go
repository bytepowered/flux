/*
The MIT License (MIT)

Copyright (c) 2017 LabStack

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package webmidware

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
	Skipper flux.WebSkipper
	// Validator 用于检查请求BasicAuth密钥的函数
	Validator func(string, string, flux.WebContext) (bool, error)
	// Default value "Restricted".
	Realm string
}

// NewBasicAuthMiddleware 返回BaseAuth中间件。
func NewBasicAuthMiddleware(validator func(string, string, flux.WebContext) (bool, error)) flux.WebInterceptor {
	return NewBasicAuthMiddlewareWith(BasicAuthConfig{
		Skipper:   func(flux.WebContext) bool { return false },
		Validator: validator,
		Realm:     defaultRealm,
	})
}

// NewBasicAuthMiddleware 返回BaseAuth中间件
func NewBasicAuthMiddlewareWith(config BasicAuthConfig) flux.WebInterceptor {
	// 参考Echo.BasicAut的实现。
	// Defaults
	if config.Validator == nil {
		panic("webex: basic-auth webmidware requires a validator function")
	}
	if config.Realm == "" {
		config.Realm = defaultRealm
	}
	return func(next flux.WebHandler) flux.WebHandler {
		return func(webc flux.WebContext) error {
			// Skip
			if config.Skipper != nil && config.Skipper(webc) {
				return next(webc)
			}
			auth := webc.HeaderValue(HeaderAuthorization)
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
			webc.SetResponseHeader(HeaderWWWAuthenticate, basic+" realm="+realm)
			return &flux.ServeError{
				StatusCode: http.StatusUnauthorized,
				ErrorCode:  "UNAUTHORIZED",
				Message:    "BASIC_AUTH:UNAUTHORIZED",
			}
		}
	}
}
