package main

import (
	"fmt"
	"sort"
)

func main() {
	//选择排序
	sliceOrder := []int{9, 5, 4, 2, 7, 3, 0, 6, 1, 8}
	for i := 0; i < len(sliceOrder); i++ {
		for j := i + 1; j < len(sliceOrder); j++ {
			if sliceOrder[i] > sliceOrder[j] {
				sliceOrder[i], sliceOrder[j] = sliceOrder[j], sliceOrder[i]
			}
		}
	}
	fmt.Println(sliceOrder)

	//冒泡排序
	sliceBubbling := []int{9, 5, 4, 2, 7, 3, 0, 6, 1, 8}
	for i := 0; i < len(sliceBubbling); i++ {
		for j := 0; j < len(sliceBubbling)-1-i; j++ {
			if sliceBubbling[j] > sliceBubbling[j+1] {
				sliceBubbling[j], sliceBubbling[j+1] = sliceBubbling[j+1], sliceBubbling[j]
			}
		}
	}
	fmt.Println(sliceBubbling)

	//sort包
	intList := []int{1, 5, 6, 7, 8, 9, 2, 3, 4}
	floatList := []float64{1.0, 4.0, 5.0, 6.0, 2.0, 3.0}
	stringList := []string{"a", "e", "f", "g", "A", "C", "b", "c", "d"}
	//默认升序
	sort.Ints(intList)
	sort.Float64s(floatList)
	sort.Strings(stringList)
	fmt.Println("intList:", intList)
	fmt.Println("floatList:", floatList)
	fmt.Println("stringList:", stringList)
	//降序
	sort.Sort(sort.Reverse(sort.IntSlice(intList)))
	sort.Sort(sort.Reverse(sort.Float64Slice(floatList)))
	sort.Sort(sort.Reverse(sort.StringSlice(stringList)))
	fmt.Println("intList:", intList)
	fmt.Println("floatList:", floatList)
	fmt.Println("stringList:", stringList)
}
