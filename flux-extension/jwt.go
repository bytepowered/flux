package extension

import (
	"fmt"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/common"
	fluxpkg "github.com/bytepowered/flux/flux-pkg"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"net/http"
	"strings"
)

const (
	TypeIdJWTFilter = "jwt_filter"
)

const (
	FeatureJWT = "feature:jwt"
)

var (
	ErrNoTokenInRequest = request.ErrNoTokenInRequest
)

var _ flux.Filter = new(JWTFilter)

type JWTConfig struct {
	// 默认查找Token的函数
	TokenExtractor func(ctx *flux.Context) (string, error)
	// 加载签名验证密钥的函数
	SecretLoader func(ctx *flux.Context) ([]byte, error)
}

func NewJWTFilter(config JWTConfig) *JWTFilter {
	return &JWTFilter{
		Config: config,
	}
}

type JWTFilter struct {
	Config JWTConfig
}

func (f *JWTFilter) FilterId() string {
	return TypeIdJWTFilter
}

func (f *JWTFilter) Init(_ *flux.Configuration) error {
	if f.Config.TokenExtractor == nil {
		f.Config.TokenExtractor = func(ctx *flux.Context) (string, error) {
			return ExtractTokenOAuth2(ctx)
		}
	}
	fluxpkg.AssertNotNil(f.Config.SecretLoader, "<secret-loader> must not nil")
	return nil
}

func (f *JWTFilter) DoFilter(next flux.FilterInvoker) flux.FilterInvoker {
	return func(ctx *flux.Context) *flux.ServeError {
		// Endpoint指定不需要授权
		if !ctx.Endpoint().Authorize() {
			return next(ctx)
		}
		tokenStr, err := f.Config.TokenExtractor(ctx)
		// 启用JWT特性，但没有传Token参数
		if ctx.Endpoint().HasAttr(FeatureJWT) && (tokenStr == "" || err == ErrNoTokenInRequest) {
			return &flux.ServeError{
				StatusCode: http.StatusUnauthorized,
				ErrorCode:  flux.ErrorCodeJwtExpired,
				Message:    "SERVER:JWT: token not found",
			}
		}
		// 解析和校验
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return f.Config.SecretLoader(ctx)
		})
		if token != nil && token.Valid {
			return next(ctx)
		}
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return &flux.ServeError{
					StatusCode: http.StatusBadRequest,
					ErrorCode:  flux.ErrorCodeJwtMalformed,
					Message:    "SERVER:JWT: token malformed",
				}
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				// Token is either expired or not active yet
				return &flux.ServeError{
					StatusCode: http.StatusUnauthorized,
					ErrorCode:  flux.ErrorCodeJwtRequires,
					Message:    "SERVER:JWT:token is expired/not active",
				}
			} else {
				return &flux.ServeError{
					StatusCode: http.StatusBadRequest,
					ErrorCode:  flux.ErrorCodeJwtMalformed,
					Message:    "SERVER:JWT: Couldn't handle",
					CauseError: err,
				}
			}
		} else {
			return &flux.ServeError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  flux.ErrorCodeJwtMalformed,
				Message:    "SERVER:JWT: Couldn't handle",
				CauseError: err,
			}
		}
	}
}

// ExtractTokenOAuth2 按OAuth2请求，从Header:Authorization和form:access_token中抓取Token
func ExtractTokenOAuth2(ctx *flux.Context) (string, error) {
	return request.OAuth2Extractor.ExtractToken(ctx.Request())
}

// ExtractTokenByFeature 根据Endpoint属性配置抓取Token的值
func ExtractTokenByFeature(ctx *flux.Context) (string, error) {
	expr := ctx.Endpoint().GetAttr(FeatureJWT).GetString()
	if "" == expr {
		return "", fmt.Errorf("<%s> not found in endpoint.attrs", FeatureJWT)
	}
	scope, key, ok := fluxpkg.LookupParseExpr(expr)
	if !ok {
		return "", fmt.Errorf("<%s> value is not a valid expr: %s", FeatureJWT, expr)
	}
	return TokenStripBearerPrefix(common.LookupWebValue(ctx, scope, key))
}

// Strips 'Bearer ' prefix from bearer token string
func TokenStripBearerPrefix(token string) (string, error) {
	// Should be a bearer token
	if len(token) > 6 && strings.ToUpper(token[0:7]) == "BEARER " {
		return token[7:], nil
	}
	return token, nil
}
