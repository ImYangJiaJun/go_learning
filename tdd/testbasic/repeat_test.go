package testbasic

import "testing"

func TestRepeat(t *testing.T) {
	cases := []struct {
		name  string
		input string
		times int
		want  string
	}{
		{name: "重复三次", input: "a", times: 3, want: "aaa"},
		{"重复两次", "ab", 2, "abab"},
		{"零次得空串", "x", 0, ""},
		{"空串重复", "", 5, ""},
		{"负数得空串", "x", -1, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Repeat(c.input, c.times)
			if got != c.want {
				t.Errorf("期望 %q，得到 %q", c.want, got)
			}
		})
	}
}

func BenchmarkRepeat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Repeat("a", 3)
	}
}
