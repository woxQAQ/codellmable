package stringx

import "strings"

func HasSuffixSlice(src string, suffix []string) bool {
	for i := range suffix {
		if strings.HasSuffix(src, suffix[i]) {
			return true
		}
	}
	return false
}
