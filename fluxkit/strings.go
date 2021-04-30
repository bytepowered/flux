package fluxkit

// StringContains 字符串列表，是否包括指定字符串
func StringContains(given []string, s string) bool {
	for _, v := range given {
		if s == v {
			return true
		}
	}
	return false
}
