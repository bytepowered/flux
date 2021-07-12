package common

import "fmt"

type extkey struct {
	id string
}

func ToNumber(in interface{}) float64 {
	v, _ := ToNumber64E(in)
	return v
}

func ToNumber64E(in interface{}) (float64, error) {
	switch v := in.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	default:
		return float64(0), fmt.Errorf("invalid number: %s", v)
	}
}
