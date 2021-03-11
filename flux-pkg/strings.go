package fluxpkg

// StringSliceContains 字符串列表，是否包括指定字符串
func StringSliceContains(elements []string, ele string) bool {
	for _, v := range elements {
		if ele == v {
			return true
		}
	}
	return false
}
