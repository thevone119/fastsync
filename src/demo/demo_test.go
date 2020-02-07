package demo

import (
	"fmt"
	"testing"
)

func TestDog(t *testing.T){
	s := [3]int{1, 2, 3}
	for i := 0; i < 2; i++ {
		defer func() {
			fmt.Println(s[i])
		}()
	}
	fmt.Println("end")
}

func Login(userName string,pwd string){
	if (userName=="dfasdfa"&&pwd=="dfasdfa" ){
		fmt.Println(userName=="dfasdfa"&&pwd=="dfasdfa")
	}
}