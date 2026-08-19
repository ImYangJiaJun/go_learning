package slicelab

func growNoReturn(s []int, v int) {
	_ = append(s, v)
}

func growReturned(s []int, v int) []int {
	t := append(s, v)
	return t
}
