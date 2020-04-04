package filter

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"github.com/bytepowered/lakego"
	"github.com/dgrijalva/jwt-go"
	"strings"
	"time"
)

const (
	FilterIdJWTVerification = "JwtVerificationFilter"
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
	subjectKey           string
	issuerKey            string
	verificationProtocol string
	verificationMethod   string
	verificationUri      string
}

// Jwt Filter，负责解码和验证Http请求的JWT令牌数据。
// 支持从Dubbo接口获取Secret。
type JwtVerificationFilter struct {
	disabled     bool
	config       JwtConfig
	certKeyCache lakego.Cache
}

func (j *JwtVerificationFilter) Init(config flux.Config) error {
	j.disabled = config.BooleanOrDefault(keyConfigDisabled, false)
	if j.disabled {
		logger.Infof("JWT filter is DISABLED !!")
		return nil
	}
	j.config = JwtConfig{
		issuerKey:            config.StringOrDefault(keyConfigJwtIssuerKey, "iss"),
		subjectKey:           config.StringOrDefault(keyConfigJwtSubjectKey, "sub"),
		verificationProtocol: config.String(keyConfigVerificationProtocol),
		verificationUri:      config.String(keyConfigVerificationUri),
		verificationMethod:   config.String(keyConfigVerificationMethod),
	}
	logger.Infof("JWT filter initializing, config: %+v", j.config)
	// Key缓存大小
	cacheExpiration := config.Int64OrDefault(keyConfigCacheExpiration, defValueCacheExpiration)
	j.certKeyCache = lakego.NewSimple(lakego.WithExpiration(time.Minute * time.Duration(cacheExpiration)))
	return nil
}

func (*JwtVerificationFilter) TypeId() string {
	return FilterIdJWTVerification
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
		// TODO 支持从Header/Query/Params读取JWT数据
		tokenString := ctx.RequestReader().Header(keyHeaderAuthorization)
		if "" == tokenString {
			return ErrorAuthorizationHeaderRequired
		}
		if claims, err := j.decodeVerified(tokenString, ctx); nil != err {
			return err
		} else {
			// Claims to scoped
			ctx.SetScopedValue(KeyScopedValueJwtClaims, claims)
			// JWT Storage
			ctx.SetAttrValue(flux.XJwtToken, tokenString)
			ctx.SetAttrValue(flux.XJwtIssuer, claims[j.config.issuerKey])
			ctx.SetAttrValue(flux.XJwtSubject, claims[j.config.subjectKey])
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
		issuer := pkg.ToString(claims[j.config.issuerKey])
		subject := pkg.ToString(claims[j.config.subjectKey])
		subjectCacheKey := fmt.Sprintf("%s.%s", issuer, subject)
		return j.certKeyCache.GetOrLoad(subjectCacheKey, func(_ lakego.Key) (lakego.Value, error) {
			switch strings.ToUpper(j.config.verificationProtocol) {
			case flux.ProtocolDubbo:
				return j.loadJwtCertKey(flux.ProtocolDubbo, issuer, subject, claims)
			case flux.ProtocolHttp:
				return j.loadJwtCertKey(flux.ProtocolHttp, issuer, subject, claims)
			default:
				return nil, fmt.Errorf("unknown verification protocol: %s", j.config.verificationProtocol)
			}
		})
	}
}

func (j *JwtVerificationFilter) loadJwtCertKey(proto string, issuer, subject string, claims jwt.MapClaims) (interface{}, error) {
	exchange, _ := ext.GetExchange(proto)
	if ret, err := exchange.Invoke(&flux.Endpoint{
		UpstreamMethod: j.config.verificationMethod,
		UpstreamUri:    j.config.verificationUri,
		Arguments: []flux.Argument{
			{TypeClass: pkg.JavaLangStringClassName, ArgName: "issuer", ArgValue: flux.NewWrapValue(issuer)},
			{TypeClass: pkg.JavaLangStringClassName, ArgName: "subject", ArgValue: flux.NewWrapValue(subject)},
			{TypeClass: pkg.JavaUtilMapClassName, ArgName: "claims", ArgValue: flux.NewWrapValue(claims)},
		},
	}, nil); nil != err {
		return false, err
	} else {
		return strings.Contains(pkg.ToString(ret), "success"), nil
	}
}
