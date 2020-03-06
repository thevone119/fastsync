package comm

import (
	"fmt"
	"strconv"
	"testing"
)

func TestDb(t *testing.T) {
	FileDB.Open()
	FileDB.PutString("db1", "kye", "value")


	//	//100万个放入，测试消耗的内存,放入100万，3秒

	for i := 0; i < 10000*100; i++ {
		if i%10000==0{
			fmt.Println(i)
			fmt.Println(FileDB.GetString("db1","key_"+strconv.FormatInt(int64(i), 10)))
		}

		//FileDB.PutString("db1", "key_"+strconv.FormatInt(int64(i), 10), "key_"+strconv.FormatInt(int64(i), 10))
		FileDB.RemoveString("db1", "key_"+strconv.FormatInt(int64(i), 10))
		//tx.Bucket([]byte("db1")).Put([]byte("key_"+strconv.FormatInt(int64(i), 10)), []byte("key_"+strconv.FormatInt(int64(i), 10)))
	}
	FileDB.Commit()
	FileDB.Close()


}
