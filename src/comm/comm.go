package comm

import (
	"os"
	"strings"
)

//判断一个文件是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}


//路径连接，串联
func AppendPath(p1 string,p2 string) string{
	if strings.LastIndex(p1,"/")==len(p1)-1{
		p1 = p1[0:len(p1)-2]
	}
	if(strings.Index(p2,"/")==0){
		p2=p2[1:]
	}
	return p1+"/"+p2
}