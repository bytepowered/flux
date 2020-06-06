package filter

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/lakego"
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
	ErrorAuthorizationHeaderRequired = &flux.InvokeError{
		StatusCode: flux.StatusUnauthorized,
		Message:    "JWT:REQUIRES_TOKEN",
	}
	ErrorIllegalToken = &flux.InvokeError{
		StatusCode: flux.StatusUnauthorized,
		Message:    "JWT:ILLEGAL_TOKEN",
	}
)

func JwtVerificationFilterFactory() interface{} {
	return new(JwtVerificationFilter)
}

type JwtConfig struct {
	lookupToken string
	subjectKey  string
	issuerKey   string
	upProto     string
	upMethod    string
	upUri       string
}

// Jwt Filter，负责解码和验证Http请求的JWT令牌数据。
// 支持从Dubbo接口获取Secret。
type JwtVerificationFilter struct {
	disabled     bool
	config       JwtConfig
	certKeyCache lakego.Cache
}

func (j *JwtVerificationFilter) Init(config flux.Configuration) error {
	j.disabled = config.GetBoolDefault(keyConfigDisabled, false)
	if j.disabled {
		logger.Infof("JWT filter is DISABLED !!")
		return nil
	}
	logger.Infof("JWT filter initializing")
	j.config = JwtConfig{
		lookupToken: config.GetStringDefault(keyConfigJwtLookupToken, keyHeaderAuthorization),
		issuerKey:   config.GetStringDefault(keyConfigJwtIssuerKey, "iss"),
		subjectKey:  config.GetStringDefault(keyConfigJwtSubjectKey, "sub"),
		upProto:     config.GetString(keyConfigUpstreamProtocol),
		upUri:       config.GetString(keyConfigUpstreamUri),
		upMethod:    config.GetString(keyConfigUpstreamMethod),
	}
	// Key缓存大小
	cacheExpiration := config.GetInt64Default(keyConfigCacheExpiration, defValueCacheExpiration)
	j.certKeyCache = lakego.NewSimple(lakego.WithExpiration(time.Minute * time.Duration(cacheExpiration)))
	return nil
}

func (*JwtVerificationFilter) TypeId() string {
	return TypeIdJWTVerification
}

func (*JwtVerificationFilter) Order() int {
	return OrderFilterJwtVerification
}

func (j *JwtVerificationFilter) Invoke(next flux.FilterInvoker) flux.FilterInvoker {
	if j.disabled {
		return next
	}
	return func(ctx flux.Context) *flux.InvokeError {
		if false == ctx.Endpoint().Authorize {
			return next(ctx)
		}
		tokenString := cast.ToString(LookupValue(j.config.lookupToken, ctx))
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

func (j *JwtVerificationFilter) decodeVerified(tokenValue string, ctx flux.Context) (jwt.MapClaims, *flux.InvokeError) {
	token, err := jwt.Parse(tokenValue, j.jwtCertKeyFactory(ctx))
	if nil != err {
		return nil, &flux.InvokeError{
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
		return j.certKeyCache.GetOrLoad(subjectCacheKey, func(_ lakego.Key) (lakego.Value, error) {
			switch strings.ToUpper(j.config.upProto) {
			case flux.ProtoDubbo:
				return j.loadJwtCertKey(flux.ProtoDubbo, issuer, subject, claims)
			case flux.ProtoHttp:
				return j.loadJwtCertKey(flux.ProtoHttp, issuer, subject, claims)
			default:
				return nil, fmt.Errorf("unknown verification protocol: %s", j.config.upProto)
			}
		})
	}
}

func (j *JwtVerificationFilter) loadJwtCertKey(proto string, issuer, subject string, claims jwt.MapClaims) (interface{}, error) {
	exchange, _ := ext.GetExchange(proto)
	if ret, err := exchange.Invoke(&flux.Endpoint{
		UpstreamMethod: j.config.upMethod,
		UpstreamUri:    j.config.upUri,
		Arguments: []flux.Argument{
			ext.NewStringArgument("issuer", issuer),
			ext.NewStringArgument("subject", subject),
			ext.NewHashMapArgument("claims", claims),
		},
	}, nil); nil != err {
		return false, err
	} else {
		return strings.Contains(cast.ToString(ret), "success"), nil
	}
}
