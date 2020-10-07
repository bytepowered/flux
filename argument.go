package flux

// ArgumentValueLookupFunc 参数值查找函数
type ArgumentValueLookupFunc func(scope, key string, context Context) (value interface{}, err error)

// 默认实现：查找Argument的值函数
func DefaultArgumentValueLookupFunc(scope, key string, ctx Context) (value interface{}, err error) {
	request := ctx.Request()
	switch scope {
	case ScopeQuery:
		return request.QueryValue(key), nil
	case ScopePath:
		return request.PathValue(key), nil
	case ScopeHeader:
		return request.HeaderValue(key), nil
	case ScopeForm:
		return request.FormValue(key), nil
	case ScopeBody:
		return request.RequestBodyReader()
	case ScopeAttrs:
		return ctx.Attributes(), nil
	case ScopeAttr:
		v, _ := ctx.GetAttribute(key)
		return v, nil
	case ScopeParam:
		if v := request.QueryValue(key); "" == v {
			return request.FormValue(key), nil
		} else {
			return v, nil
		}
	default:
		if v := request.PathValue(key); "" != v {
			return v, nil
		} else if v := request.QueryValue(key); "" != v {
			return v, nil
		} else if v := request.FormValue(key); "" != v {
			return v, nil
		} else if v := request.HeaderValue(key); "" != v {
			return v, nil
		} else if v, _ := ctx.GetAttribute(key); "" != v {
			return v, nil
		} else {
			return nil, nil
		}
	}
}
