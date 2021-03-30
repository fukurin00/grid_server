package tool

import (
	"sort"
)

func MinFloat(a []float64) float64 {
	sort.Float64s(a)
	return a[0]
}

func MaxFloat(a []float64) float64 {
	sort.Float64s(a)
	return a[len(a)-1]
}

func ReverseSlice(a []uint8) []uint8 {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
	return a
}
