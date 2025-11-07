//go:build !solution

package fileleak

import (
	"fmt"
	"os"
)

type testingT interface {
	Errorf(msg string, args ...interface{})
	Cleanup(func())
}

func VerifyNone(t testingT) {
	before := snapshotOpenFiles()
	t.Cleanup(func() {
		after := snapshotOpenFiles()
		leaks := difference(after, before)
		if len(leaks) > 0 {
			t.Errorf("file leaks detected: %v", leaks)
		}
	})
}

func snapshotOpenFiles() map[string]int {
	files := make(map[string]int)
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return files
	}

	for _, e := range entries {
		path := "/proc/self/fd/" + e.Name()
		target, err := os.Readlink(path)
		if err == nil {
			files[target]++
		}
	}
	return files
}

func difference(after, before map[string]int) []string {
	var leaks []string
	for target, countAfter := range after {
		countBefore := before[target]
		if countAfter > countBefore {
			leaks = append(leaks,
				fmt.Sprintf("%s (extra %d)", target, countAfter-countBefore))
		}
	}
	return leaks
}
