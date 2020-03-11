package comm

import (
	"os"
	"path/filepath"
	"time"
)

//文件信息记录，这个是记录文件的具体信息，比如文件的最后修改时间，文件的大小，文件的完整MD5值，文件的读取时间等。
//这个文件信息，可以保存到存储中，加速文件信息的读取，避免每次都读取文件的MD5,每次都进行上传等，减少文件的上传次数。
//这个文件存储，计划使用B树 的文件索引方式进行存储，加速访问性能。
//对于后续的全目录，全文件检索进行加速。

//文件信息
type FFileInfo struct {
	Path string			//文件路径
	Size int64			//文件大小
	ModTime int64		//文件的最后修改时间，纳秒
	MD5 []byte			//文件的MD5,完整的MD5值
	ReadTime int64		//文件上次读取时间，纳秒
	UpLoadTime int64	//文件的上传时间，纳秒
	IsDir	bool		//是否目录
	IsExist bool		//文件是否存在
}

func NewFFileInfo(p string) *FFileInfo{
	return &FFileInfo{
		Path:p,
		IsDir:false,
		IsExist:false,
	}
}

//重载文件的基础信息，文件大小，文件的最后修改时间，是否目录，文件是否存在等信息
func (f *FFileInfo) ReLoadBase(){
	fi, err := os.Stat(f.Path)
	if err!=nil{
		f.IsExist=false
		f.IsDir=filepath.Ext(f.Path) == ""
		f.ModTime=time.Now().UnixNano()
		return
	}
	f.Size=fi.Size()
	f.IsExist=true
	f.ModTime=fi.ModTime().UnixNano()
	f.IsDir=fi.IsDir()
}

//返回bytes信息，方便文件存储
func (f *FFileInfo) getBytes(){

}

