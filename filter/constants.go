package filter

// 全局Filter排序
const (
	OrderFilterJwtVerification        = -90
	OrderFilterPermissionVerification = -80
)

const (
	keyConfigCacheExpiration  = "cache-expiration"
	keyConfigDisabled         = "disabled"
	keyConfigUpstreamProtocol = "upstream-protocol"
	keyConfigUpstreamHost     = "upstream-host"
	keyConfigUpstreamUri      = "upstream-uri"
	keyConfigUpstreamMethod   = "upstream-method"
	keyConfigJwtSubjectKey    = "jwt-subject-key"
	keyConfigJwtIssuerKey     = "jwt-issuer-key"
)

const (
	KeyScopedValueJwtClaims = "jwt-claims"
)

const (
	defValueCacheExpiration = 30
)
