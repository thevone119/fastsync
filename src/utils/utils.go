package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"io"
	"math"
	"os"
	"sync"
)

/*
ID 生成器
*/
var uintId uint32 = 1  //ID生成器
var intId int32 = 1  //ID生成器
var IdLock sync.Mutex //保护PidGen的互斥机制
func GetNextUint() uint32{
	IdLock.Lock()
	if uintId>=math.MaxUint32{
		uintId=1
	}
	uintId++
	IdLock.Unlock()
	return uintId
}
func GetNextInt() int32{
	IdLock.Lock()
	if intId>=math.MaxInt32{
		intId=1
	}
	intId++
	IdLock.Unlock()
	return intId
}


func GetFileMd5(fp string,ct byte ) ([]byte, error){
	hash := md5.New()
	var result []byte
	switch ct {
	case 0:
		info, err := os.Lstat(fp)
		bytesBuffer := bytes.NewBuffer([]byte{})
		if err!=nil{
			bytesBuffer.WriteByte(0)
			hash.Write(bytesBuffer.Bytes())
			return hash.Sum(result), nil
		}else{
			binary.Write(bytesBuffer, binary.BigEndian, int64(info.Size()))
			hash.Write(bytesBuffer.Bytes())
			return hash.Sum(result), nil
		}

	case 1:
		file, _ := os.Open(fp)
		defer file.Close()
		var result []byte
		//获取文件大小
		buf_len, _ := file.Seek(0, io.SeekEnd)
		file.Seek(0, io.SeekStart)
		//只取10块内容做MD5
		var clean = buf_len / 10
		hash := md5.New()
		var temp = make([]byte, 1024)
		var count = 0
		for {
			rn, err := file.Read(temp)
			if err != nil || rn <= 0 {
				break
			}
			file.Seek(clean,io.SeekCurrent)
			hash.Write(temp)
			count = count + 1
			if count > 10 {
				//break
			}
		}
		len_b := make([]byte, 8)
		binary.BigEndian.PutUint64(len_b, uint64(buf_len))
		hash.Write(len_b)
		return hash.Sum(result), nil
	case 2:
		file, _ := os.Open(fp)
		defer file.Close()
		var result []byte
		hash := md5.New()
		if _, err := io.Copy(hash, file); err != nil {
			return result, err
		}
		return hash.Sum(result), nil

	}
	hash.Write([]byte{0})
	return hash.Sum(result), nil
}

//获取文件的MD5
func ComputeMd5(file *os.File) ([]byte, error) {
	defer file.Close()
	var result []byte
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}
	//file.Seek(0, os.SEEK_SET)
	//file.Close()
	return hash.Sum(result), nil
}

//获取文件的MD5
func ComputeFastMd5(file *os.File) ([]byte, error) {
	defer file.Close()
	var result []byte
	//获取文件大小
	buf_len, _ := file.Seek(0, os.SEEK_END)
	file.Seek(0, os.SEEK_SET)
	//只取10块内容做MD5
	var clean = buf_len / 10
	hash := md5.New()
	var temp = make([]byte, 1024)
	var count = 0
	for {
		rn, err := file.Read(temp)
		if err != nil || rn <= 0 {
			break
		}
		file.Seek(clean, os.SEEK_CUR)
		hash.Write(temp)
		count = count + 1
		if count > 10 {
			//break
		}
	}

	len_b := make([]byte, 8)
	binary.BigEndian.PutUint64(len_b, uint64(buf_len))
	hash.Write(len_b)
	//file.Seek(0, os.SEEK_SET)
	return hash.Sum(result), nil
}


