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
