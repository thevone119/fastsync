package comm

import (
	"bytes"
	"encoding/binary"
	"os"
	"strings"
)

//文件校验类型定义
type CheckFileType byte

//枚举文件校验类型
const (
	FCHECK_NOT_CHECK           = CheckFileType(0) //无需校验直接上传
	FCHECK_SIZE_CHECK          = CheckFileType(1) //只校验大小，不同则上传
	FCHECK_FASTMD5_CHECK       = CheckFileType(2) //快速MD5校验，不同则上传（无法保证完整）
	FCHECK_FULLMD5_CHECK       = CheckFileType(3) //完整的MD5校验		（用于增量文章发布同步）
	FCHECK_SIZE_AND_TIME_CHECK = CheckFileType(4) //校验大小，如果大小一样，则校验时间，如果时间较新，则更新（用于全量快速同步）
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
func AppendPath(p1 string, p2 string) string {
	if strings.LastIndex(p1, "/") == len(p1)-1 && len(p1) > 0 {
		p1 = p1[0 : len(p1)-1]
	}
	if strings.LastIndex(p2, "/") == len(p2)-1 && len(p2) > 0 {
		p2 = p2[0 : len(p2)-1]
	}

	if strings.Index(p2, "/") == 0 {
		p2 = p2[1:]
	}
	return p1 + "/" + p2
}

//整形转换成字节
func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}
func UIntToBytes(n uint32) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

//字节转换成整形
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)

	return int(x)
}
