package errhandling

import (
	"errors"
	"fmt"
)

var ErrEmptyName = errors.New("empty name")
var ErrInvalidAge = errors.New("invalid age")

// ValidationError 自定义错误类型：调用方用 errors.As 取出 Field 字段
type ValidationError struct {
	Field string
	Msg   string
}

// Error 是指针接收者：只有 *ValidationError 实现了 error 接口。
// 所以 Validate 必须返回 &ValidationError{...}，errors.As 的目标类型也必须是 *ValidationError。
func (e *ValidationError) Error() string {
	return fmt.Sprintf("field %s: %s", e.Field, e.Msg)
}

// Validate 只校验 Age：Age < 0 → 返回 &ValidationError{Field: "age", ...}；否则 nil
func Validate(u User) error {
	if u.Age < 0 {
		return &ValidationError{Field: "age", Msg: "age must be greater than zero"}
	}
	return nil
}

// ValidateAll 校验 Name 和 Age，收集全部违规（不是遇到第一个就返回）：
//   - Name == "" → ErrEmptyName；Age < 0 → ErrInvalidAge
//   - 有违规 → errors.Join 聚合返回；全合法 → nil
func ValidateAll(u User) error {
	// 违规先逐个 append 进切片，互不覆盖。
	// （若写成 err = errors.Join(...) 赋值两次，后一次会覆盖前一次，丢违规。）
	var errs []error
	if u.Name == "" {
		errs = append(errs, ErrEmptyName)
	}
	if u.Age < 0 {
		errs = append(errs, ErrInvalidAge)
	}
	// errors.Join 把多个错误聚合成一棵树，errors.Is 对树做深度遍历，两个哨兵都能命中；
	// errs 为空（全合法）时 Join 返回 nil，"全合法返回 nil"由此自然成立。
	return errors.Join(errs...)
}
