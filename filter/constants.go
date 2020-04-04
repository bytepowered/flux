package filter

// 全局Filter排序
const (
	OrderFilterJwtVerification        = -90
	OrderFilterPermissionVerification = -80
)

const (
	keyConfigCacheExpiration      = "cache-expiration"
	keyConfigDisabled             = "disabled"
	keyConfigVerificationProtocol = "verification-protocol"
	keyConfigVerificationHost     = "verification-host"
	keyConfigVerificationUri      = "verification-uri"
	keyConfigVerificationMethod   = "verification-method"
	keyConfigJwtSubjectKey        = "jwt-subject-key"
	keyConfigJwtIssuerKey         = "jwt-issuer-key"
	keyConfigJwtLookupKey         = "jwt-lookup-key"
)

const (
	KeyScopedValueJwtClaims = "jwt-claims"
)

const (
	defValueCacheExpiration = 30
)
