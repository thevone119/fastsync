package btree2

import (
	"math/rand"
	"testing"
)

func TestBTree(t *testing.T) {
	//插入
	bsTree := BT{nil,2,[M+1]int{0,21,38,0,0},[M+1]*BT{}}
	newTree := bsTree.Insert(rand.Int())
	for i:=0;i<10000*100;i++{
		newTree = bsTree.Insert(rand.Int())
	}

	//删除
	newTree = newTree.Delete(38)
	newTree = newTree.Delete(39)
	newTree = newTree.Delete(42)

	//newTree.BTreeTraverse()
	//time.Sleep(time.Hour)
}