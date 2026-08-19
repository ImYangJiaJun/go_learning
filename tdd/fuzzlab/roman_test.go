package fuzzlab

import (
	"errors"
	"strings"
	"testing"
)

func TestArabicToRoman(t *testing.T) {
	cases := []struct {
		name    string
		arabic  int
		want    string
		wantErr error
	}{
		{"超出上限", 4001, "", ErrOutOfRange},
		{"超出下限", -1, "", ErrOutOfRange},
		{"正常1", 1, "I", nil},
		{"正常3", 3, "III", nil},
		{"正常4", 4, "IV", nil},
		{"正常9", 9, "IX", nil},
		{"正常14", 14, "XIV", nil},
		{"正常40", 40, "XL", nil},
		{"正常2024", 2024, "MMXXIV", nil},
		{"正常3999", 3999, "MMMCMXCIX", nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ArabicToRoman(c.arabic)
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("错误断言失败：期望 %v，得到 %v", c.wantErr, err)
			}
			if got != c.want {
				t.Errorf("期望 %q，得到 %q", c.want, got)
			}
		})
	}
}

func TestArabicToRomanProperties(t *testing.T) {
	for n := 1; n <= 3999; n++ {
		roman, err := ArabicToRoman(n)
		// 性质 0：合法输入永不报错（err 必须先 Fatal 掉，否则 roman 不可用）
		if err != nil {
			t.Errorf("合法输入：%v 出现报错%v", n, err)
		}

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
}
