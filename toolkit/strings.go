package toolkit

import "strings"

func MatchEqual(expects []string, s string) bool {
	return MatchOn(expects, s, stringEqual)
}

func MatchEqualFold(expects []string, s string) bool {
	return MatchOn(expects, s, strings.EqualFold)
}

func MatchPrefix(expects []string, s string) bool {
	return MatchOn(expects, s, strings.HasPrefix)
}

func MatchSuffix(expects []string, s string) bool {
	return MatchOn(expects, s, strings.HasSuffix)
}

func MatchContains(expects []string, s string) bool {
	return MatchOn(expects, s, strings.Contains)
}

func MatchOn(expects []string, in string, tf func(in, expect string) bool) bool {
	for _, exp := range expects {
		if tf(in, exp) {
			return true
		}
	}
	return false
}

func stringEqual(v, expect string) bool {
	return expect == v
}
