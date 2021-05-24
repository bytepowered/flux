package flux

import (
	"fmt"
	"net/http"
)

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

var _ error = new(ServeError)

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

func (e *ServeError) GetExtra(key string) interface{} {
	return e.Extras[key]
}

func (e *ServeError) SetExtra(key string, value interface{}) {
	if e.Extras == nil {
		e.Extras = make(map[string]interface{}, 4)
	}
	e.Extras[key] = value
}

func (e *ServeError) MergeHeader(header http.Header) *ServeError {
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

type (
	// ServeResponseWriter 用于解析和序列化响应数据结构的接口，并将序列化后的数据写入Http响应流。
	ServeResponseWriter interface {
		// Write 写入正常响应数据
		Write(ctx *Context, response *ServeResponse)
		// WriteError 写入发生错误响应数据
		WriteError(ctx *Context, err *ServeError)
	}
	// ServeResponse 表示后端服务(Dubbo/Http/gRPC/Echo)返回响应数据结构，
	// 包含后端期望透传的状态码、Header和Attachment等数据
	ServeResponse struct {
		StatusCode  int                    // Http状态码
		Headers     http.Header            // Http Header
		Attachments map[string]interface{} // Attachment
		Body        interface{}            // 响应数据体
	}
)

func NewServeResponse(status int, body interface{}) *ServeResponse {
	return &ServeResponse{
		StatusCode:  status,
		Headers:     make(http.Header, 0),
		Attachments: make(map[string]interface{}, 0),
		Body:        body,
	}
}
