package libs

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	fmt.Println(TimeFormat(50))
	fmt.Println(TimeFormat(500))
	fmt.Println(TimeFormat(5000))
	fmt.Println(TimeFormat(50000))
	fmt.Println(TimeFormat(500000))
	fmt.Println(TimeFormat(5000000))
}