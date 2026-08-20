package genericlab

// All 返回 s 的迭代器：按下标顺序产出每个元素；
// 消费方在 yield 中返回 false 时立即停止遍历，且之后不得再调用 yield
/*
	迭代器是「待执行的遍历」：返回值是个函数，被调用时才真正遍历
	for v := range All(s) 被编译器翻译为：All(s)(func(v T) bool { 循环体; return true })
	continue → return true，break → return false
	该函数类型即标准库的 iter.Seq[T]，可直接喂给 slices.Collect / slices.Sorted
*/
func All[T any](s []T) func(yield func(T) bool) {
	return func(yield func(T) bool) {
		for _, v := range s {
			// yield 返回 false = 消费方 break，必须立刻停（再调 yield 会 panic）
			if !yield(v) {
				break
			}
		}
	}
}
