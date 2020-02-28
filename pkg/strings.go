package pkg

func StrContains(item string, array []string) bool {
	for _, v := range array {
		if item == v {
			return true
		}
	}
	return false
}
