package filter

import (
	"fmt"
	"github.com/bytepowered/cache"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/support"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/cast"
	"strings"
	"time"
)

const (
	TypeIdJWTVerification = "JwtVerificationFilter"
)

const (
	keyHeaderAuthorization = "Authorization"
)

var (
	ErrorAuthorizationHeaderRequired = &flux.StateError{
		StatusCode: flux.StatusUnauthorized,
		ErrorCode:  flux.ErrorCodeRequestInvalid,
		Message:    "JWT:REQUIRES_TOKEN",
	}
	ErrorIllegalToken = &flux.StateError{
		StatusCode: flux.StatusUnauthorized,
		ErrorCode:  flux.ErrorCodeRequestInvalid,
		Message:    "JWT:ILLEGAL_TOKEN",
	}
)

func JwtVerificationFilterFactory() interface{} {
	return new(JwtVerificationFilter)
}

type JwtConfig struct {
	lookupToken    string
	subjectKey     string
	issuerKey      string
	upstreamProto  string
	upstreamMethod string
	upstreamUri    string
}

// Jwt Filter，负责解码和验证Http请求的JWT令牌数据。
// 支持从Dubbo接口获取Secret。
type JwtVerificationFilter struct {
	disabled     bool
	config       JwtConfig
	certKeyCache cache.Cache
}

func (j *JwtVerificationFilter) Init(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		ConfigKeyCacheExpiration: defValueCacheExpiration,
		ConfigKeyCacheSize:       defValueCacheSize,
		ConfigKeyDisabled:        false,
		JwtConfigKeyLookupToken:  keyHeaderAuthorization,
		JwtConfigKeyIssuer:       "iss",
		JwtConfigKeySubject:      "sub",
	})
	j.disabled = config.GetBool(ConfigKeyDisabled)
	if j.disabled {
		logger.Info("JWT filter is DISABLED")
		return nil
	}
	logger.Info("JWT filter initializing")
	j.config = JwtConfig{
		lookupToken:    config.GetString(JwtConfigKeyLookupToken),
		issuerKey:      config.GetString(JwtConfigKeyIssuer),
		subjectKey:     config.GetString(JwtConfigKeySubject),
		upstreamProto:  config.GetString(UpstreamConfigKeyProtocol),
		upstreamUri:    config.GetString(UpstreamConfigKeyUri),
		upstreamMethod: config.GetString(UpstreamConfigKeyMethod),
	}
	// Key缓存大小
	expiration := time.Minute * time.Duration(config.GetInt64(ConfigKeyCacheExpiration))
	size := config.GetInt(ConfigKeyCacheSize)
	j.certKeyCache = cache.New(size).Expiration(expiration).LRU().Build()
	return nil
}

func (*JwtVerificationFilter) TypeId() string {
	return TypeIdJWTVerification
}

func (*JwtVerificationFilter) Order() int {
	return OrderFilterJwtVerification
}

func (j *JwtVerificationFilter) DoFilter(next flux.FilterHandler) flux.FilterHandler {
	if j.disabled {
		return next
	}
	return func(ctx flux.Context) *flux.StateError {
		if false == ctx.Endpoint().Authorize {
			return next(ctx)
		}
		tokenString := cast.ToString(support.ScopeLookupContextValue(j.config.lookupToken, ctx))
		if "" == tokenString {
			return ErrorAuthorizationHeaderRequired
		}
		if claims, err := j.decodeVerified(tokenString, ctx); nil != err {
			return err
		} else {
			// Claims to scoped
			ctx.SetValue(KeyScopedValueJwtClaims, claims)
			// JWT Storage
			ctx.SetAttribute(flux.XJwtToken, tokenString)
			ctx.SetAttribute(flux.XJwtIssuer, claims[j.config.issuerKey])
			ctx.SetAttribute(flux.XJwtSubject, claims[j.config.subjectKey])
			return next(ctx)
		}
	}
}

func (j *JwtVerificationFilter) decodeVerified(tokenValue string, ctx flux.Context) (jwt.MapClaims, *flux.StateError) {
	token, err := jwt.Parse(tokenValue, j.jwtCertKeyFactory(ctx))
	if nil != err {
		return nil, &flux.StateError{
			StatusCode: flux.StatusUnauthorized,
			Message:    "JWT:PARSING",
			Internal:   err,
		}
	}
	if !token.Valid {
		return nil, ErrorIllegalToken
	}
	return token.Claims.(jwt.MapClaims), nil
}

func (j *JwtVerificationFilter) jwtCertKeyFactory(_ flux.Context) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("unexpected claims : %s", token.Claims)
		}
		// 获取用户标识，从缓存中加载用户的JWT密钥
		issuer := cast.ToString(claims[j.config.issuerKey])
		subject := cast.ToString(claims[j.config.subjectKey])
		subjectCacheKey := fmt.Sprintf("%s.%s", issuer, subject)
		return j.certKeyCache.GetOrLoad(subjectCacheKey, func(key interface{}) (interface{}, *time.Duration, error) {
			switch strings.ToUpper(j.config.upstreamProto) {
			case flux.ProtoDubbo:
				return j.loadJwtCertKey(flux.ProtoDubbo, issuer, subject, claims)
			case flux.ProtoHttp:
				return j.loadJwtCertKey(flux.ProtoHttp, issuer, subject, claims)
			default:
				return nil, cache.NoExpiration, fmt.Errorf("unknown verification protocol: %s", j.config.upstreamProto)
			}
		})
	}
}

func (j *JwtVerificationFilter) loadJwtCertKey(proto string, issuer, subject string, claims jwt.MapClaims) (interface{}, *time.Duration, error) {
	backend, _ := ext.GetBackend(proto)
	if ret, err := backend.Invoke(&flux.Endpoint{
		UpstreamMethod: j.config.upstreamMethod,
		UpstreamUri:    j.config.upstreamUri,
		Arguments: []flux.Argument{
			ext.NewStringArgument("issuer", issuer),
			ext.NewStringArgument("subject", subject),
			ext.NewHashMapArgument("claims", claims),
		},
	}, nil); nil != err {
		return false, cache.NoExpiration, err
	} else {
		return strings.Contains(cast.ToString(ret), "success"), cache.NoExpiration, nil
	}
}
