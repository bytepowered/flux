package flux

import "fmt"

// Resolve 解析Argument参数值
func (a Argument) Resolve(ctx Context) (interface{}, error) {
	if nil != a.ValueLoader {
		return a.ValueLoader(), nil
	}
	if nil == a.LookupFunc {
		return nil, fmt.Errorf("LookupFunc is nil, name: %s", a.Name)
	}
	if nil == a.ValueResolver {
		return nil, fmt.Errorf("ValueResolver is nil, name: %s", a.Name)
	}
	// Single value
	if len(a.Fields) == 0 {
		mtv, err := a.LookupFunc(a.HttpScope, a.HttpName, ctx)
		if nil != err {
			return nil, err
		}
		return a.ValueResolver(mtv, a.Class, a.Generic)
	}
	// POJO Values
	sm := make(map[string]interface{}, len(a.Fields))
	sm["class"] = a.Class
	for _, field := range a.Fields {
		if fv, err := field.Resolve(ctx); nil != err {
			return nil, err
		} else {
			sm[field.Name] = fv
		}
	}
	return sm, nil
}
