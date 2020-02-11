package comm

import (
	"fmt"
	"testing"
)

func TestSyncConfig(t *testing.T) {
	fmt.Println(len(SyncConfigObj.RemotePath))
}
