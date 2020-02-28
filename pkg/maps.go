package pkg

func ToStringKVMap(attrs map[string]interface{}) map[string]string {
	attachment := make(map[string]string)
	for k, v := range attrs {
		attachment[k] = ToString(v)
	}
	return attachment
}
