package swr

import (
	"fmt"
	"testing"
)

func TestSwr(t *testing.T) {
	swr := NewSw()
	swr.Add("a", 5)
	swr.Add("b", 1)
	swr.Add("c", 1)
	for i := 0; i < 7; i ++ {
		fmt.Println(swr.Get())
	}
}
