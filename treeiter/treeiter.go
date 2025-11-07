//go:build !solution

package treeiter

import "reflect"

type Tree[T any] interface {
	Left() T
	Right() T
}

func DoInOrder[T Tree[T]](tree T, f func(T)) {
	if isNil(tree) {
		return
	}

	left := tree.Left()
	if !isNil(left) {
		DoInOrder(left, f)
	}

	f(tree)

	right := tree.Right()
	if !isNil(right) {
		DoInOrder(right, f)
	}
}

func isNil[T any](v T) bool {
	rv := reflect.ValueOf(v)
	return rv.IsNil()
}
