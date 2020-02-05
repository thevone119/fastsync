package demo

import (
	"fmt"
	"testing"
)

func TestDog(t *testing.T){
	username:="dfasdfa"

	fmt.Println(username)
	fmt.Println(username=="dfasdfa"&&username=="dfasdfa")
}

func Login(userName string,pwd string){
	if (userName=="dfasdfa"&&pwd=="dfasdfa" ){
		fmt.Println(userName=="dfasdfa"&&pwd=="dfasdfa")
	}
}