package comm

import (
	"fmt"
	"strconv"
	"testing"
)

func TestDb(t *testing.T) {
	FileDB.PutString("db1", "kye", "value")
	for i := 0; i < 10000*1; i++ {
		fmt.Println(FileDB.GetString("db2", "key_"+strconv.FormatInt(int64(i), 10)))
		FileDB.RemoveString("db2", "key_"+strconv.FormatInt(int64(i), 10))
		//FileDB.PutString("db1","key_"+strconv.FormatInt(int64(i),10),"123fasdfas")
	}

}
