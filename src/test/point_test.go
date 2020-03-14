package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"zinx/zlog"
)

type TPoint struct {
	Name string
	EUpLoadTime []int64				//各个服务器端结束上传时间,毫秒
}

func NewPoint() *TPoint{
	r:= &TPoint{Name:"123"}
	r.EUpLoadTime = make([]int64,10)

	return r
}

func ChangePoint(p *TPoint){
	p.Name="cChangePoint"
}

func GetPoint(p *TPoint) string{
	return p.Name
}

func print(lf *TPoint){
	for _,stime:=range lf.EUpLoadTime{
		zlog.Debug("EUpLoadTime:",stime)

	}
	fmt.Println(lf.EUpLoadTime[1])
}

func TestPoint(t *testing.T) {
	fmt.Println(os.Args[0])
	fmt.Println(filepath.Dir(os.Args[0]))
	c:=make(chan *TPoint, 10)
	p:=NewPoint()
	p.Name="fff"

	c<-p
	data, ok := <-c
	if ok{
		data.Name="chan"
		data.EUpLoadTime[1]=10
		fmt.Println(GetPoint(data))
		print(data)
	}
	fmt.Println(GetPoint(p))
	ChangePoint(p)
	fmt.Println(GetPoint(p))
}