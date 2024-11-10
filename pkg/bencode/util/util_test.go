package util

import (
	"fmt"
	"testing"
)

func TestGeneratePeerID(t *testing.T) {
	l := GeneratePeerID("dsm")
	fmt.Println(l)
}
