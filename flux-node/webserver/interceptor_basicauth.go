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
package webserver

import (
	"encoding/base64"
	flux2 "github.com/bytepowered/flux/flux-node"
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
	Skipper flux2.WebSkipper
	// Validator 用于检查请求BasicAuth密钥的函数
	Validator func(string, string, flux2.WebExchange) (bool, error)
	// Default value "Restricted".
	Realm string
}

// NewBasicAuthMiddleware 返回BaseAuth中间件。
func NewBasicAuthMiddleware(validator func(string, string, flux2.WebExchange) (bool, error)) flux2.WebInterceptor {
	return NewBasicAuthMiddlewareWith(BasicAuthConfig{
		Skipper:   func(flux2.WebExchange) bool { return false },
		Validator: validator,
		Realm:     defaultRealm,
	})
}

// NewBasicAuthMiddleware 返回BaseAuth中间件
func NewBasicAuthMiddlewareWith(config BasicAuthConfig) flux2.WebInterceptor {
	// 参考Echo.BasicAut的实现。
	// Defaults
	if config.Validator == nil {
		panic("webex: basic-auth webmidware requires a validator function")
	}
	if config.Realm == "" {
		config.Realm = defaultRealm
	}
	return func(next flux2.WebHandler) flux2.WebHandler {
		return func(webex flux2.WebExchange) error {
			// Skip
			if config.Skipper != nil && config.Skipper(webex) {
				return next(webex)
			}
			auth := webex.HeaderVar(HeaderAuthorization)
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
						valid, err := config.Validator(cred[:i], cred[i+1:], webex)
						if err != nil {
							return err
						} else if valid {
							return next(webex)
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
			webex.SetResponseHeader(HeaderWWWAuthenticate, basic+" realm="+realm)
			return &flux2.ServeError{
				StatusCode: http.StatusUnauthorized,
				ErrorCode:  "UNAUTHORIZED",
				Message:    "BASIC_AUTH:UNAUTHORIZED",
			}
		}
	}
}
