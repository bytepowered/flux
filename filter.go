package flux

import "fmt"

const (
	ErrorCodeGatewayInternal = "GATEWAY:INTERNAL"
	ErrorCodeGatewayExchange = "GATEWAY:EXCHANGE"
	ErrorCodeGatewayEndpoint = "GATEWAY:ENDPOINT"
	ErrorCodeRequestInvalid  = "REQUEST:INVALID"
	ErrorCodeRequestNotFound = "REQUEST:NOT_FOUND"
)

// StateError 定义网关请求错误，它包含：状态码、消息、内部错误等元数据
type StateError struct {
	StatusCode int    // 错误状态码
	ErrorCode  string // 错误码
	Message    string // 错误消息
	Internal   error  // 内部错误
}

func (e *StateError) Error() string {
	return fmt.Sprintf("StateError: StatusCode=%d, ErrorCode=%s, Message=%s, Error=%s", e.StatusCode, e.ErrorCode, e.Message, e.Internal)
}

// FilterInvoker 定义一个处理方法
type FilterInvoker func(Context) *StateError

// Filter 用于定义处理方法的顺序及内在业务逻辑
type Filter interface {
	// TypeId Filter的类型标识
	TypeId() string
	// Invoke 执行Filter链
	Invoke(next FilterInvoker) FilterInvoker
}
