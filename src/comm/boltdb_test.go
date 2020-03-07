package comm

import (
	"fmt"
	"testing"
)

func TestDb(t *testing.T) {
	BoltDB.Open()
	BoltDB.PutString("db1", "22key22212", "value")
	BoltDB.PutInt64("db1","flee12",1212)
	//FileDB.ForceCommit()
	fmt.Println(BoltDB.GetString("db1","22key22212"))
	fmt.Println(BoltDB.GetInt64("db1","flee12"))

	//	//100万个放入，测试消耗的内存,放入100万，3秒

	//FileDB.Close()
	select {

	}


}
