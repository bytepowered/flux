package flux

import "fmt"

const (
	ErrorCodeGatewayInternal  = "GATEWAY:INTERNAL"
	ErrorCodeGatewayBackend   = "GATEWAY:BACKEND"
	ErrorCodeGatewayEndpoint  = "GATEWAY:ENDPOINT"
	ErrorCodeGatewayCircuited = "GATEWAY:CIRCUITED"
	ErrorCodeRequestInvalid   = "REQUEST:INVALID"
	ErrorCodeRequestNotFound  = "REQUEST:NOT_FOUND"
)

// StateError 定义网关请求错误，它包含：状态码、消息、内部错误等元数据
type StateError struct {
	StatusCode int    // 错误状态码
	ErrorCode  string // 错误码
	Message    string // 错误消息
	Internal   error  // 内部错误
}

func (e *StateError) Error() string {
	if nil != e.Internal {
		return fmt.Sprintf("StateError: StatusCode=%d, ErrorCode=%s, Message=%s, Error=%s", e.StatusCode, e.ErrorCode, e.Message, e.Internal)
	} else {
		return fmt.Sprintf("StateError: StatusCode=%d, ErrorCode=%s, Message=%s", e.StatusCode, e.ErrorCode, e.Message)
	}
}

type (
	// FilterHandler 定义一个处理方法，处理请求Context；如果发生错误则返回 StateError。
	FilterHandler func(Context) *StateError
	// FilterSkipper 定义一个函数，用于Filter执行中跳过某些处理。返回True跳过某些处理，见具体Filter的实现逻辑。
	FilterSkipper func(Context) bool
)

// Filter 用于定义处理方法的顺序及内在业务逻辑
type Filter interface {
	// TypeId Filter的类型标识
	TypeId() string
	// DoFilter 执行Filter链
	DoFilter(next FilterHandler) FilterHandler
}
