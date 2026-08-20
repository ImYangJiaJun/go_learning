package genericlab

// Map 对 s 的每个元素应用 f，按原顺序返回结果组成的新切片；
// 输入为空（含 nil）时返回非 nil 的空切片
func Map[T, R any](s []T, f func(T) R) []R {
	result := make([]R, len(s))
	for k, t := range s {
		result[k] = f(t)
	}
	return result
}

// Filter 返回 s 中所有让 pred 返回 true 的元素，保持原相对顺序；
// 没有元素满足（含输入为空）时返回非 nil 的空切片
func Filter[T any](s []T, pred func(T) bool) []T {
	result := make([]T, 0, len(s))
	for _, t := range s {
		if pred(t) {
			result = append(result, t)
		}
	}
	return result
}

// Reduce 以 init 为初始累积值，从左到右用 f(累积值, 元素) 依次折叠整个切片；
// s 为空时原样返回 init。R 允许与 T 不同类型（如 []int 折叠成 string）
func Reduce[T, R any](s []T, init R, f func(R, T) R) R {
	acc := init
	for _, t := range s {
		acc = f(acc, t)
	}
	return acc
}
