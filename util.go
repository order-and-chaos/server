package main

import (
	"fmt"
	"strconv"
)

var index uint64 = 1

// UniqID returns and increments the current index
func UniqID() uint64 {
	res := index
	index++
	return res
}

// FormatID returns a new 'random' string id
func FormatID(id uint64) string {
	return fmt.Sprintf("%03s", strconv.FormatUint(id*46649%6125, 36))
}

// UniqIDf returns a new 'random' string id and increaes the index
func UniqIDf() string {
	return FormatID(UniqID())
}
