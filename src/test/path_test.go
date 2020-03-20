package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestPath(t *testing.T) {
	s:="e:/project/aa/xx.log"
	fmt.Println(	filepath.Dir(s))
	fi,err:=os.Stat(s)
	if err!=nil{
		fmt.Println("not find")
	}else{
		fmt.Println(fi.IsDir())
	}
	sv,_:=filepath.Rel(s,"e:/portal")
	fmt.Println("rel",sv)

	fmt.Println(filepath.Match("xx*.log",s))
	//通过Walk来遍历目录下f的所有子目录
	filepath.Walk("e:/project", func(p string, info os.FileInfo, err error) error {
		//fmt.Println(p)
		return nil
	})

}