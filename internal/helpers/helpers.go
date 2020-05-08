package helpers

func Dedupe(s []string) []string {
	keys := make(map[string]struct{}, len(s))
	ret := make([]string, 0, len(s))
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = struct{}{}
			ret = append(ret, entry)
		}
	}
	return ret
}

func MaskAccessToken(s string) string {
	rs := []rune(s)
	for i := 0; i < len(rs)-4; i++ {
		rs[i] = '*'
	}
	return string(rs)
}
