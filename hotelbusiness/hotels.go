//go:build !solution

package hotelbusiness

// package main

import (
	"slices"
)

type Guest struct {
	CheckInDate  int
	CheckOutDate int
}

type Load struct {
	StartDate  int
	GuestCount int
}

func sortedKeys(m map[int]int) []int {
	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func ComputeLoad(guests []Guest) []Load {
	diffPlan := make(map[int]int)
	var loadPlan []Load

	for _, elem := range guests {
		diffPlan[elem.CheckInDate]++
		diffPlan[elem.CheckOutDate]--
	}
	daysSorted := sortedKeys(diffPlan)
	var currentGuests int

	for _, day := range daysSorted {
		diff := diffPlan[day]

		if diff == 0 {
			continue
		}

		currentGuests += diff
		loadPlan = append(loadPlan, Load{day, currentGuests})
	}

	return loadPlan
}

// func main() {
// 	fmt.Println(ComputeLoad([]Guest{{1, 2}, {2, 3}}))

// }
