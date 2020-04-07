// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
//

package common

import (
	"crypto/sha256"
	"runtime"
	"strings"
)

func BytesReverse(u []byte) []byte {
	for i, j := 0, len(u)-1; i < j; i, j = i+1, j-1 {
		u[i], u[j] = u[j], u[i]
	}
	return u
}

func Sha256D(data []byte) [32]byte {
	once := sha256.Sum256(data)
	return sha256.Sum256(once[:])
}

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