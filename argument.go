package flux

// EndpointArgumentValueLookupFunc 参数值查找函数
type EndpointArgumentValueLookupFunc func(scope, key string, context Context) (value interface{}, err error)

// 默认实现：查找Argument的值函数
func DefaultEndpointArgumentValueLookup(scope, key string, ctx Context) (value interface{}, err error) {
	request := ctx.Request()
	switch scope {
	case ScopeQuery:
		return request.QueryValue(key), nil
	case ScopePath:
		return request.PathValue(key), nil
	case ScopeParam:
		if v := request.QueryValue(key); "" == v {
			return request.FormValue(key), nil
		} else {
			return v, nil
		}
	case ScopeHeader:
		return request.HeaderValue(key), nil
	case ScopeForm:
		return request.FormValue(key), nil
	case ScopeAttrs:
		return ctx.Attachments(), nil
	case ScopeAttr:
		value, _ := ctx.GetAttachment(key)
		return value, nil
	case ScopeAuto:
		fallthrough
	default:
		if v := request.PathValue(key); "" != v {
			return v, nil
		} else if v := request.QueryValue(key); "" != v {
			return v, nil
		} else if v := request.FormValue(key); "" != v {
			return v, nil
		} else if v := request.HeaderValue(key); "" != v {
			return v, nil
		} else if v, _ := ctx.GetAttachment(key); "" != v {
			return v, nil
		} else {
			return nil, nil
		}
	}
}
