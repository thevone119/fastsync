package comm

import (
	"fmt"
	"testing"
	"time"
	"utils"
)

func TestTempMap(t *testing.T) {
	TempMap.Put("key","value",10)
	TempMap.Put("key2","value",10)
	TempMap.Put("key3","value",10)
	TempMap.Remove("key5")
	for{
		v,ok:=TempMap.Get("key")
		if ok==false{
			fmt.Println("---------err")
		}else{
			fmt.Println("value:",v)
		}
		currtime := time.Now().UnixNano() / 1e6
		mdb,err:=utils.GetFileMd5("/test/UnityPlayer.dll",1)
		fmt.Println("md5 usetime :",time.Now().UnixNano() / 1e6-currtime)

		if err!=nil{
			fmt.Println("md5 :",err)
		}else{
			fmt.Println("md5 len:",len(mdb))
			fmt.Println("md5:",fmt.Sprintf("%x",mdb))
		}


		time.Sleep(1* time.Second)
	}

}
