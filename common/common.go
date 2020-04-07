package common

import (
	"runtime"
	"strings"
)

// Goid returns the current goroutine id.
func Goid() string {
	var buf [18]byte
	n := runtime.Stack(buf[:], false)
	fields := strings.Fields(string(buf[:n]))
	if len(fields) <= 1 {
		return ""
	}
	return fields[1]
}