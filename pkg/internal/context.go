package internal

// ContextKey 用于隐藏内部实现的Key
type ContextKey string

var (
	CtxkeyRequestId  = ContextKey("flux.go/context.request.id/926820fa-7ad8-4444-9080-d690ce31c93a")
	CtxkeyWebContext = ContextKey("flux.go/context.web.context/890b1fa9-93ad-4b44-af24-85bcbfe646b4")
)
