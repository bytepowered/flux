package flux

const (
	ErrorCodeGatewayInternal  = "GATEWAY:INTERNAL"
	ErrorCodeGatewayBackend   = "GATEWAY:BACKEND"
	ErrorCodeGatewayEndpoint  = "GATEWAY:ENDPOINT"
	ErrorCodeGatewayCircuited = "GATEWAY:CIRCUITED"
	ErrorCodeRequestInvalid   = "REQUEST:INVALID"
	ErrorCodeRequestNotFound  = "REQUEST:NOT_FOUND"
	ErrorCodePermissionDenied = "PERMISSION:ACCESS_DENIED"
)

const (
	ErrorMessageBackendDecodeResponse  = "BACKEND:DECODE_RESPONSE"
	ErrorMessageBackendDecoderNotFound = "BACKEND:DECODER:NOT_FOUND"

	ErrorMessageDubboInvokeFailed        = "BACKEND:DU:INVOKE"
	ErrorMessageDubboAssembleFailed      = "BACKEND:DU:ASSEMBLE"
	ErrorMessageDubboDecodeInvalidHeader = "BACKEND:DU:DECODE:INVALID_HEADERS"
	ErrorMessageDubboDecodeInvalidStatus = "BACKEND:DU:DECODE:INVALID_STATUS"

	ErrorMessageHttpInvokeFailed   = "BACKEND:HT:INVOKE"
	ErrorMessageHttpAssembleFailed = "BACKEND:HT:ASSEMBLE"

	ErrorMessageHystrixCircuited = "HYSTRIX:CIRCUITED"

	ErrorMessagePermissionAccessDenied    = "PERMISSION:ACCESS_DENIED"
	ErrorMessagePermissionServiceNotFound = "PERMISSION:SERVICE:NOT_FOUND"
	ErrorMessagePermissionVerifyError     = "PERMISSION:VERIFY:ERROR"

	ErrorMessageEndpointVersionNotFound  = "ENDPOINT:VERSION:NOT_FOUND"
	ErrorMessageWebServerResponseMarshal = "SERVER:RESPONSE:MARSHAL"

	ErrorMessageRequestPrepare = "REQUEST:BODY:PREPARE"
	ErrorMessageRequestParsing = "REQUEST:BODY:PARSING"
)
