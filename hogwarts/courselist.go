//go:build !solution

package hogwarts

func Keys[T, V comparable](m map[T]V) []T {
	var keys []T
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func flattenMap[T comparable](Map map[T][]T) map[T]struct{} {
	flattenedMap := make(map[T]struct{})

	for key, vs := range Map {
		flattenedMap[key] = struct{}{}
		for _, v := range vs {
			flattenedMap[v] = struct{}{}
		}
	}
	return flattenedMap
}

func DFS(course string, visitedMap map[string]int, prereqs map[string][]string, courseList []string) []string {
	switch visitedMap[course] {
	case 1:
		// 1: node in process
		panic("Сycle")
	case 2:
		// 2: processed node
		return courseList
	}

	visitedMap[course] = 1

	for _, pre := range prereqs[course] {
		courseList = DFS(pre, visitedMap, prereqs, courseList)
	}

	visitedMap[course] = 2

	courseList = append(courseList, course)
	return courseList
}

func GetCourseList(prereqs map[string][]string) []string {
	allCourses := flattenMap(prereqs)
	visitedMap := make(map[string]int)
	courseList := []string{}

	for course := range allCourses {
		if visitedMap[course] == 0 {
			courseList = DFS(course, visitedMap, prereqs, courseList)
		}
	}
	return courseList
}
