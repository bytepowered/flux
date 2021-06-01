package flux

const (
	ErrorCodeGatewayInternal    = "GATEWAY:INTERNAL"
	ErrorCodeGatewayTransporter = "GATEWAY:TRANSPORTER"
	ErrorCodeGatewayEndpoint    = "GATEWAY:ENDPOINT"
)

const (
	ErrorCodeRequestInvalid   = "GATEWAY:REQUEST:INVALID"
	ErrorCodeRequestNotFound  = "GATEWAY:REQUEST:NOT_FOUND"
	ErrorCodeRequestCircuited = "GATEWAY:REQUEST:CIRCUITED"
	ErrorCodeRequestCanceled  = "GATEWAY:REQUEST:CANCELED"
)

const (
	ErrorCodePermissionDenied = "GATEWAY:PERMISSION:ACCESS_DENIED"
)

const (
	ErrorCodeJwtMalformed = "GATEWAY:AUTHORIZATION:JWT:MALFORMED"
	ErrorCodeJwtExpired   = "GATEWAY:AUTHORIZATION:JWT:EXPIRED"
	ErrorCodeJwtNotFound  = "GATEWAY:AUTHORIZATION:JWT:NOTFOUND"
)

const (
	ErrorMessageProtocolUnknown = "GATEWAY:PROTOCOL:UNKNOWN"

	// Transport errors
	ErrorMessageTransportDubboInvokeFailed        = "TRANSPORT:DU:INVOKE/error"
	ErrorMessageTransportDubboAssembleFailed      = "TRANSPORT:DU:ASSEMBLE/error"
	ErrorMessageTransportDubboClientCanceled      = "TRANSPORT:DU:CANCELED/client"
	ErrorMessageTransportDubboDecodeInvalidHeader = "TRANSPORT:DU:DECODE:INVALID/headers"
	ErrorMessageTransportDubboDecodeInvalidStatus = "TRANSPORT:DU:DECODE:INVALID/status"
	ErrorMessageTransportHttpInvokeFailed         = "TRANSPORT:HT:INVOKE/error"
	ErrorMessageTransportHttpAssembleFailed       = "TRANSPORT:HT:ASSEMBLE/error"
	ErrorMessageTransportCodecError               = "TRANSPORT:CODEC/error"

	ErrorMessagePermissionAccessDenied    = "PERMISSION:ACCESS_DENIED"
	ErrorMessagePermissionServiceNotFound = "PERMISSION:SERVICE:NOT_FOUND"
	ErrorMessagePermissionVerifyError     = "PERMISSION:VERIFY:ERROR"

	ErrorMessageWebServerRequestNotFound = "SERVER:REQUEST:NOT_FOUND"

	ErrorMessageRequestPrepare = "REQUEST:BODY:PREPARE"
)
