package filter

const (
	ConfigKeyCacheExpiration  = "cache-expiration"
	ConfigKeyCacheDisabled    = "cache-disabled"
	ConfigKeyCacheSize        = "cache-size"
	ConfigKeyDisabled         = "disabled"
	UpstreamConfigKeyProtocol = "upstream-protocol"
	UpstreamConfigKeyHost     = "upstream-host"
	UpstreamConfigKeyUri      = "upstream-uri"
	UpstreamConfigKeyMethod   = "upstream-method"
	JwtConfigKeySubject       = "jwt-subject-key"
	JwtConfigKeyIssuer        = "jwt-issuer-key"
	JwtConfigKeyLookupToken   = "jwt-lookup-token"
)

const (
	KeyScopedValueJwtClaims = "jwt-claims"
)

const (
	DefaultValueCacheExpiration = 30
	DefaultValueCacheSize       = 1_0000
)
