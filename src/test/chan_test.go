package test

import (
	"fmt"
	"testing"
)

func TestChan(t *testing.T){
	ch := make(chan int,10)
	ch<-1
	ch<-1
	ch<-1
	ch<-1
	ch<-1

	for i:=0;i<0;i++{
		<-ch
	}
	fmt.Println("ch len:",len(ch))
}
