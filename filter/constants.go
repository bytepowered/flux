package filter

// 全局Filter排序
const (
	OrderFilterParameterParsing       = -100
	OrderFilterJwtVerification        = -90
	OrderFilterPermissionVerification = -80
)

const (
	keyConfigCacheExpiration      = "cache-expiration"
	keyConfigDisabled             = "disabled"
	keyConfigVerificationProtocol = "verification-protocol"
	keyConfigVerificationUri      = "verification-uri"
	keyConfigVerificationMethod   = "verification-method"
	keyConfigJwtSubjectKey        = "jwt-subject-key"
	keyConfigJwtIssuerKey         = "jwt-issuer-key"
)

const (
	KeyScopedValueJwtClaims = "jwt-claims"
)

const (
	defValueCacheExpiration = 30
)
