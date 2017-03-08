package main

import (
	"fmt"
	"strconv"
)

// FormatID returns a new 'random' string id
func FormatID(id uint64) string {
	return fmt.Sprintf("%03s", strconv.FormatUint(id*46649%6125, 36))
}

// IDGenerator can be used to generate unique IDs
type IDGenerator struct {
	index uint64
}

// MakeIDGenerator makes an IDGenerator
func MakeIDGenerator() IDGenerator {
	return IDGenerator{
		index: 0,
	}
}

// UniqID returns and increments the current index
func (g *IDGenerator) UniqID() uint64 {
	res := g.index
	g.index++
	return res
}

// UniqIDf returns a new 'random' string id and increaes the index
func (g *IDGenerator) UniqIDf() string {
	return FormatID(g.UniqID())
}

var gen = MakeIDGenerator()

// UniqID returns and increments the current index
func UniqID() uint64 {
	return gen.UniqID()
}

// UniqIDf returns a new 'random' string id and increaes the index
func UniqIDf() string {
	return gen.UniqIDf()
}
