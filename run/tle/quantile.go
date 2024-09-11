package tle

import (
	"sort"
)

func Quantile(xs []int64, q float64) int64 {
	if !(0 <= q && q < 1) {
		panic("q must be in [0, 1)")
	}
	xs = append([]int64{}, xs...)
	sort.Slice(xs, func(i, j int) bool {
		return xs[i] < xs[j]
	})
	return xs[int(float64(len(xs))*q)]
}
