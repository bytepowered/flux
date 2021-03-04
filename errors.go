package flux

import "net/http"

const (
	ErrorCodeGatewayInternal  = "GATEWAY:INTERNAL"
	ErrorCodeGatewayBackend   = "GATEWAY:BACKEND"
	ErrorCodeGatewayEndpoint  = "GATEWAY:ENDPOINT"
	ErrorCodeGatewayCircuited = "GATEWAY:CIRCUITED"
	ErrorCodeGatewayCanceled  = "GATEWAY:CANCELED"
	ErrorCodeRequestInvalid   = "REQUEST:INVALID"
	ErrorCodeRequestNotFound  = "REQUEST:NOT_FOUND"
	ErrorCodePermissionDenied = "PERMISSION:ACCESS_DENIED"
)

const (
	ErrorMessageBackendDecodeResponse = "BACKEND:DECODE_RESPONSE"

	ErrorMessageDubboInvokeFailed        = "BACKEND:DU:INVOKE"
	ErrorMessageDubboAssembleFailed      = "BACKEND:DU:ASSEMBLE"
	ErrorMessageDubboDecodeInvalidHeader = "BACKEND:DU:DECODE:INVALID_HEADERS"
	ErrorMessageDubboDecodeInvalidStatus = "BACKEND:DU:DECODE:INVALID_STATUS"

	ErrorMessageHttpInvokeFailed   = "BACKEND:HT:INVOKE"
	ErrorMessageHttpAssembleFailed = "BACKEND:HT:ASSEMBLE"

	ErrorMessagePermissionAccessDenied    = "PERMISSION:ACCESS_DENIED"
	ErrorMessagePermissionServiceNotFound = "PERMISSION:SERVICE:NOT_FOUND"
	ErrorMessagePermissionVerifyError     = "PERMISSION:VERIFY:ERROR"

	ErrorMessageWebServerRequestNotFound = "SERVER:REQUEST:NOT_FOUND"

	ErrorMessageRequestPrepare = "REQUEST:BODY:PREPARE"
)

var (
	ErrRouteNotFound = &ServeError{
		StatusCode: http.StatusNotFound,
		ErrorCode:  ErrorCodeRequestNotFound,
		Message:    ErrorMessageWebServerRequestNotFound,
	}
)
