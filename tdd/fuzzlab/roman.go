package fuzzlab

import (
	"errors"
	"strings"
)

var ErrOutOfRange = errors.New("out of range")

var romanNumerals = []struct {
	value  int
	symbol string
}{
	{1000, "M"},
	{900, "CM"},
	{500, "D"},
	{400, "CD"},
	{100, "C"},
	{90, "XC"},
	{50, "L"},
	{40, "XL"},
	{10, "X"},
	{9, "IX"},
	{5, "V"},
	{4, "IV"},
	{1, "I"},
}

func ArabicToRoman(num int) (string, error) {
	if num > 3999 || num < 1 {
		return "", ErrOutOfRange
	}

	var result strings.Builder

	for _, roman := range romanNumerals {
		for num >= roman.value {
			result.WriteString(roman.symbol)
			num -= roman.value
		}
	}

	return result.String(), nil
}
