package flux

type Selection struct {
	Filters []string // 选中的FilterId列表
}

// Selector 用于请求路由前的组件选择
type Selector interface {
	Select(ctx Context) Selection
}
