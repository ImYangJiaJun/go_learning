package errhandling

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		name      string
		user      User
		wantErr   bool   // 是否期望校验失败
		wantField string // 期望提取出的 ValidationError.Field（仅 wantErr 时检查）
	}{
		{name: "年龄非法", user: User{Name: "Tom", Age: -1}, wantErr: true, wantField: "age"},
		{name: "合法", user: User{Name: "Tom", Age: 0}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := Validate(c.user)

			if !c.wantErr {
				if err != nil {
					t.Errorf("期望没有错误，得到 %v", err)
				}
				return
			}

			// errors.As 沿错误链按类型提取 *ValidationError，还能读出里面的 Field。
			// 第二参数必须传指针的指针（&ve 的类型是 **ValidationError），As 靠它把结果写回 ve。
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("期望 errors.As 提取出 *ValidationError，得到 %v", err)
			}
			if ve.Field != c.wantField {
				t.Errorf("期望 Field 为 %q，得到 %q", c.wantField, ve.Field)
			}
		})
	}

	// 这个用例的输入不是 User（不调用 Validate），放不进上面的表格，单独写一个子测试：
	// 链上没有 *ValidationError 这个类型时，errors.As 必须返回 false。
	t.Run("非校验错误", func(t *testing.T) {
		var ve *ValidationError
		if errors.As(ErrUserNotFound, &ve) {
			t.Errorf("ErrUserNotFound 不应被 errors.As 匹配为 *ValidationError")
		}
	})
}

func TestValidateAll(t *testing.T) {
	cases := []struct {
		name           string
		user           User
		wantEmptyName  bool // 期望命中 ErrEmptyName
		wantInvalidAge bool // 期望命中 ErrInvalidAge
	}{
		{name: "两个字段都非法", user: User{Name: "", Age: -1}, wantEmptyName: true, wantInvalidAge: true},
		{name: "只有一个非法", user: User{Name: "", Age: 1}, wantEmptyName: true},
		{name: "全合法", user: User{Name: "Tom", Age: 20}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateAll(c.user)

			// 每个哨兵都双向检查：期望命中的要中，不期望的不许误伤。
			// errors.Join 聚成错误树，errors.Is 深度遍历，两个哨兵可以各自命中。
			if got := errors.Is(err, ErrEmptyName); got != c.wantEmptyName {
				t.Errorf("errors.Is(err, ErrEmptyName) = %v，期望 %v（err = %v）", got, c.wantEmptyName, err)
			}
			if got := errors.Is(err, ErrInvalidAge); got != c.wantInvalidAge {
				t.Errorf("errors.Is(err, ErrInvalidAge) = %v，期望 %v（err = %v）", got, c.wantInvalidAge, err)
			}
			// 两个哨兵都不期望时，err 必须真的是 nil（Join 收到空切片返回 nil）。
			if !c.wantEmptyName && !c.wantInvalidAge && err != nil {
				t.Errorf("期望没有错误，得到 %v", err)
			}
		})
	}
}
