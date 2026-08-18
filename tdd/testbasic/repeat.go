package testbasic

import "strings"

// Repeat 将 s 重复 n 次返回；n <= 0 时返回空字符串
func Repeat(s string, n int) string {
	var str strings.Builder
	for range n {
		str.WriteString(s)
	}
	return str.String()
}
