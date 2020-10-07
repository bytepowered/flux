package flux

import "net/url"

// ArgumentValueLookupFunc 参数值查找函数
type ArgumentValueLookupFunc func(scope, key string, context Context) (value MIMETypeValue, err error)

// 默认实现：查找Argument的值函数
func DefaultArgumentValueLookupFunc(scope, key string, ctx Context) (value MIMETypeValue, err error) {
	request := ctx.Request()
	switch scope {
	case ScopeQuery:
		return wrapTextMIMETypeValue(request.QueryValue(key)), nil
	case ScopePath:
		return wrapTextMIMETypeValue(request.PathValue(key)), nil
	case ScopeHeader:
		return wrapTextMIMETypeValue(request.HeaderValue(key)), nil
	case ScopeForm:
		return wrapTextMIMETypeValue(request.FormValue(key)), nil
	case ScopeBody:
		reader, err := request.RequestBodyReader()
		return MIMETypeValue{Value: reader, MIMEType: request.HeaderValue("Content-Type")}, err
	case ScopeAttrs:
		return MIMETypeValue{Value: ctx.Attributes(), MIMEType: ValueMIMETypeLangStringMap}, nil
	case ScopeAttr:
		v, _ := ctx.GetAttribute(key)
		return wrapObjectMIMETypeValue(v), nil
	case ScopeParam:
		if v := request.QueryValue(key); "" != v {
			return wrapTextMIMETypeValue(v), nil
		} else {
			return wrapTextMIMETypeValue(request.FormValue(key)), nil
		}
	default:
		find := func(key string, sources ...url.Values, ) (string, bool) {
			for _, source := range sources {
				if vs, ok := source[key]; ok {
					return vs[0], true
				}
			}
			return "", false
		}
		if v, ok := find(key, request.PathValues(), request.QueryValues(), request.FormValues()); ok {
			return wrapTextMIMETypeValue(v), nil
		} else if v := request.HeaderValue(key); "" != v {
			return wrapTextMIMETypeValue(v), nil
		} else if v, _ := ctx.GetAttribute(key); "" != v {
			return wrapObjectMIMETypeValue(v), nil
		} else {
			return wrapObjectMIMETypeValue(value), nil
		}
	}
}

func wrapTextMIMETypeValue(value interface{}) MIMETypeValue {
	return MIMETypeValue{Value: value, MIMEType: ValueMIMETypeLangText}
}

func wrapObjectMIMETypeValue(value interface{}) MIMETypeValue {
	return MIMETypeValue{Value: value, MIMEType: ValueMIMETypeLangObject}
}
