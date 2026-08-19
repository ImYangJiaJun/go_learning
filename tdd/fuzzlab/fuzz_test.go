package fuzzlab

import (
	"errors"
	"strings"
	"testing"
)

func FuzzArabicToRoman(f *testing.F) {
	// 种子语料：人给的"重点嫌疑值"
	f.Add(1)    // 下界
	f.Add(4)    // 减法形式代表
	f.Add(3999) // 上界
	f.Fuzz(func(t *testing.T, n int) {
		roman, err := ArabicToRoman(n)
		if n < 1 || n > 3999 {
			// 越界输入：必须返回 ErrOutOfRange（errors.Is 断言）
			if !errors.Is(err, ErrOutOfRange) {
				t.Errorf("Expected ErrOutOfRange, got %v", err)
			}
		} else {
			// 合法输入：err 为 nil，且满足行为 2 的两条性质
			const validRoman = "IVXLCDM"

			count := 0
			var prev rune
			for _, ch := range roman {
				// 性质 1：只能出现 IVXLCDM
				if !strings.ContainsRune(validRoman, ch) {
					t.Errorf("输入：%v  出现IVXLCDM以外的字符 %v", n, ch)
				}
				// 性质 2：同一字符连续出现不超过 3 次
				if ch == prev {
					count++
					if count > 3 {
						t.Errorf("输入：%v  字符%v连续出现超过3次", n, ch)
					}
				} else {
					prev = ch
					count = 1
				}
			}
		}
	})
}
