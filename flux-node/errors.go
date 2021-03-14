package flux

import (
	"fmt"
	"github.com/spf13/cast"
	"net/http"
)

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
	ErrorMessageProtocolUnknown = "GATEWAY:PROTOCOL:UNKNOWN"

	ErrorMessageBackendDecodeResponse = "BACKEND:DECODE_RESPONSE"
	ErrorMessageBackendWriteResponse  = "BACKEND:WRITE_RESPONSE"

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

// ServeError 定义网关处理请求的服务错误；
// 它包含：错误定义的状态码、错误消息、内部错误等元数据
type ServeError struct {
	StatusCode int                    // 响应状态码
	ErrorCode  interface{}            // 业务错误码
	Message    string                 // 错误消息
	CauseError error                  // 内部错误对象；错误对象不会被输出到请求端；
	Header     http.Header            // 响应Header
	Extras     map[string]interface{} // 用于定义和跟踪的额外信息；额外信息不会被输出到请求端；
}

func (e *ServeError) Error() string {
	if nil != e.CauseError {
		return fmt.Sprintf("ServeError: StatusCode=%d, ErrorCode=%s, Message=%s, Extras=%+v, Error=%s", e.StatusCode, e.ErrorCode, e.Message, e.Extras, e.CauseError)
	} else {
		return fmt.Sprintf("ServeError: StatusCode=%d, ErrorCode=%s, Message=%s, Extras=%+v", e.StatusCode, e.ErrorCode, e.Message, e.Extras)
	}
}

func (e *ServeError) GetErrorCode() string {
	return cast.ToString(e.ErrorCode)
}

func (e *ServeError) ExtraByKey(key string) interface{} {
	return e.Extras[key]
}

func (e *ServeError) SetExtra(key string, value interface{}) {
	if e.Extras == nil {
		e.Extras = make(map[string]interface{}, 4)
	}
	e.Extras[key] = value
}

func (e *ServeError) Merge(header http.Header) *ServeError {
	if e.Header == nil {
		e.Header = header.Clone()
	} else {
		for key, values := range header {
			for _, value := range values {
				e.Header.Add(key, value)
			}
		}
	}
	return e
}
