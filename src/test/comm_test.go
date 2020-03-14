package test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestDate(t *testing.T) {
	now  := time.Now()
	//Year = now.Year()
	//Mouth  = now.Month()
	//Day  =  now.Day()
	//时间格式化输出 Printf输出
	fmt.Println(time.Now().YearDay())
	fmt.Println(filepath.Abs("e:/project222/\\affw"))
	fmt.Printf("当前时间为： %d-%d-%d %d:%d:%d\n",now.Year(),now.Month(),now.Day(),now.Hour(),now.Minute(),now.Second())
	//fmt.Sprintf 格式化输出
	dateString := fmt.Sprintf("当前时间为： %d-%d-%d %d:%d:%d\n",now.Year(),now.Month(),now.Day(),now.Hour(),now.Minute(),now.Second())
	fmt.Println(dateString)
	//now.Format 方法格式化
	fmt.Println(now.Format("20060102"))
	fmt.Println(now.Format("2001/01/02 15:04:05"))
	fmt.Println(now.Format("2006/01/02"))//年月日
	fmt.Println(now.Format("15:04:05"))//时分秒
}