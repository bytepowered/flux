package flux

import "fmt"

// InvokeError 定义网关请求错误，它包含：状态码、消息、内部错误等元数据
type InvokeError struct {
	StatusCode int    // 错误状态码
	Message    string // 错误消息
	Internal   error  // 内部错误
}

func (e *InvokeError) Error() string {
	return fmt.Sprintf("InvokeError: StatusCode=%d, Message=%s, Error=%s", e.StatusCode, e.Message, e.Internal)
}

// FilterInvoker 定义一个处理方法
type FilterInvoker func(Context) *InvokeError

// Filter 用于定义处理方法的顺序及内在业务逻辑
type Filter interface {
	Identity
	Invoke(next FilterInvoker) FilterInvoker
}
