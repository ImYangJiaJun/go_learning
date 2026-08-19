package dictionary

import (
	"errors"
	"testing"
)

func TestSearch(t *testing.T) {
	// 两个子测试都只读，共享一份字典是安全的；从 TestAdd 开始出现写操作，就必须每个用例各自新建
	d := Dictionary{"test": "this is just a test"}

	tests := []struct {
		name string
		word string
		want string
		err  error
	}{
		{name: "查存在的词", word: "test", want: "this is just a test", err: nil},
		{name: "查不存在的词", word: "unknown", want: "", err: ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := d.Search(tt.word)

			// err 不符合预期时 got 没有意义，后续断言必须中止，所以这里用 Fatalf 而不是 Errorf
			if !errors.Is(err, tt.err) {
				t.Fatalf("Search %q: 期望错误 %v，得到 %v", tt.word, tt.err, err)
			}
			if got != tt.want {
				t.Errorf("Search %q: 期望 %q，得到 %q", tt.word, tt.want, got)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	t.Run("添加新词", func(t *testing.T) {
		// 写操作的用例各自新建字典，不共享可变状态——否则先跑的用例会污染后跑的，
		// 失败时分不清是实现错了还是数据被上一个用例改过
		d := Dictionary{}

		err := d.Add("test", "this is just a test")
		assertError(t, err, nil)

		// 补上原来缺的断言：只查 err 会放过"返回 nil 但实际没写入"的空实现，
		// 必须再 Search 一次确认释义真的写进去了
		assertDefinition(t, d, "test", "this is just a test")
	})

	t.Run("重复添加", func(t *testing.T) {
		// 旧值 "old" 与新值 "new" 故意不同，才能检出"被覆盖"
		d := Dictionary{"test": "old"}

		err := d.Add("test", "new")
		assertError(t, err, ErrWordExists)

		// 补上原来缺的断言：报错不等于没写——"返回 ErrWordExists 但照样覆盖"的实现
		// 只靠上面的错误断言抓不到，必须确认旧释义还在
		assertDefinition(t, d, "test", "old")
	})
}

func TestUpdate(t *testing.T) {
	t.Run("更新存在的词", func(t *testing.T) {
		d := Dictionary{"test": "old"}

		err := d.Update("test", "new")
		assertError(t, err, nil)

		// 补上原来缺的断言：确认新值真的生效，防"返回 nil 但没写"的实现
		assertDefinition(t, d, "test", "new")
	})

	t.Run("更新不存在的词", func(t *testing.T) {
		d := Dictionary{}

		err := d.Update("test", "new")
		assertError(t, err, ErrWordDoesNotExist)

		// 补上原来缺的断言：确认 Search 仍返回 ErrNotFound——
		// 防的是"先写 map 再报错"这类让 Update 悄悄退化成 Add 的实现
		_, searchErr := d.Search("test")
		assertError(t, searchErr, ErrNotFound)
	})
}

func TestDelete(t *testing.T) {
	t.Run("删除后查不到", func(t *testing.T) {
		d := Dictionary{"test": "this is just a test"}

		d.Delete("test")

		_, err := d.Search("test")
		assertError(t, err, ErrNotFound)
	})

	t.Run("删不存在的词", func(t *testing.T) {
		d := Dictionary{}

		// 删不存在的 key 是 no-op：不 panic、不算错误，所以 Delete 没有返回值
		d.Delete("ghost")

		_, err := d.Search("ghost")
		assertError(t, err, ErrNotFound)
	})
}

// assertError 判断哨兵错误一律用 errors.Is，禁止比较错误文案。
// t.Helper() 让断言失败时的行号指向调用方，而不是指向本函数内部
func assertError(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Errorf("期望错误 %v，得到 %v", want, got)
	}
}

// assertDefinition 封装"Search 后比对释义"这一重复出现的断言组合。
// 查不到词时 got 没有意义，用 Fatalf 直接中止当前用例
func assertDefinition(t *testing.T, d Dictionary, word, want string) {
	t.Helper()
	got, err := d.Search(word)
	if err != nil {
		t.Fatalf("期望查到 %q，得到错误 %v", word, err)
	}
	if got != want {
		t.Errorf("期望 %q，得到 %q", want, got)
	}
}
