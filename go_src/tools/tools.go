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

// スライスを反転する
func ReverseSlice(a []uint8) []uint8 {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
	return a
}

// startとfinishの間にあるかcheck
func CheckSpan(start, finish, target float64) bool {
	if start <= target && target <= finish {
		return true
	}
	return false
}

// return same components in two lists
func CheckDuplicate(a, b []int) []int {
	var dup []int
	for _, aa := range a {
		for _, bb := range b {
			if aa == bb {
				if !CheckSameCom(aa, dup) {
					dup = append(dup, aa)
				}
			}
		}
	}
	return dup
}

// remove duplicated components in a single list
func RemoveDuplicate(l []int) []int {
	keys := make(map[int]bool)
	list := []int{}

	for _, entry := range l {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// check an component:t is in list:l
func CheckSameCom(t int, l []int) bool {
	for _, a := range l {
		if t == a {
			return true
		}
	}
	return false
}
