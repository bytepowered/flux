package flux

// ArgumentLookupResolver 参数值查找函数
type ArgumentLookupResolver func(argument Argument, context Context) (interface{}, error)

// 默认实现：查找Argument的值函数
func DefaultArgumentLookupResolver(arg Argument, ctx Context) (interface{}, error) {
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
		return ctx.Attachments(), nil
	case ScopeAttr:
		value, _ := ctx.GetAttachment(arg.HttpKey)
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
		} else if v, _ := ctx.GetAttachment(arg.HttpKey); "" != v {
			return v, nil
		} else {
			return nil, nil
		}
	}
}
