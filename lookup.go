package flux

// ArgumentLookupFunc 参数值查找函数
type ArgumentLookupFunc func(argument Argument, context Context) (interface{}, error)

// 默认实现：查找Argument的值函数
func DefaultArgumentValueLookupFunc(arg Argument, ctx Context) (interface{}, error) {
	request := ctx.Request()
	switch arg.HttpScope {
	case ScopeQuery:
		return request.QueryValue(arg.HttpKey), nil
	case ScopePath:
		return request.PathValue(arg.HttpKey), nil
	case ScopeParam:
		if v := request.QueryValue(arg.HttpKey); "" == v {
			return request.FormValue(arg.HttpKey), nil
		} else {
			return v, nil
		}
	case ScopeHeader:
		return request.HeaderValue(arg.HttpKey), nil
	case ScopeForm:
		return request.FormValue(arg.HttpKey), nil
	case ScopeAttrs:
		return ctx.Attributes(), nil
	case ScopeAttr:
		value, _ := ctx.GetAttribute(arg.HttpKey)
		return value, nil
	case ScopeAuto:
		fallthrough
	default:
		if v := request.PathValue(arg.HttpKey); "" != v {
			return v, nil
		} else if v := request.QueryValue(arg.HttpKey); "" != v {
			return v, nil
		} else if v := request.FormValue(arg.HttpKey); "" != v {
			return v, nil
		} else if v := request.HeaderValue(arg.HttpKey); "" != v {
			return v, nil
		} else if v, _ := ctx.GetAttribute(arg.HttpKey); "" != v {
			return v, nil
		} else {
			return nil, nil
		}
	}
}
