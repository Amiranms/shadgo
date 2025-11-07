//go:build !solution

package genericsum

import (
	"cmp"
	"math/cmplx"
	"slices"
	"sync"

	"golang.org/x/exp/constraints"
)

func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func SortSlice[S ~[]T, T cmp.Ordered](a S) {
	slices.Sort(a)
}

func MapsEqual[S ~map[KeyT]ValT, KeyT, ValT comparable](a, b S) bool {
	small := a
	big := b

	if len(small) > len(big) {
		small, big = big, small
	}

	for k, v1 := range big {
		v2, ok := small[k]
		if !ok {
			return false
		}
		if v2 != v1 {
			return false
		}
	}
	return true
}

func SliceContains[T comparable](s []T, v T) bool {
	for _, k := range s {
		if k == v {
			return true
		}
	}
	return false
}

func MergeChans[T any](chs ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	wg.Add(len(chs))
	for _, ch := range chs {
		go func(ch <-chan T) {
			for t := range ch {
				out <- t
			}
			wg.Done()
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

type ComplexLike interface {
	constraints.Integer | constraints.Float | constraints.Complex
}

func toComplex[T ComplexLike](v T) complex128 {
	switch tv := any(v).(type) {
	case int:
		return complex(float64(tv), 0)
	case int8:
		return complex(float64(tv), 0)
	case int32:
		return complex(float64(tv), 0)
	case int64:
		return complex(float64(tv), 0)
	case float32:
		return complex(float64(tv), 0)
	case float64:
		return complex(tv, 0)
	case complex64:
		return complex(float64(real(tv)), float64(imag(tv)))
	case complex128:
		return tv
	default:
		panic("invalid value")
	}
}

func IsHermitianMatrix[T ComplexLike](m [][]T) bool {
	for i, r := range m {
		for j, _ := range r {
			v1 := toComplex(m[i][j])
			v2 := toComplex(m[j][i])
			if v1 != cmplx.Conj(v2) || v2 != cmplx.Conj(v1) {
				return false
			}

		}
	}
	return true
}
