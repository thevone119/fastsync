package comm

import (
	"bytes"
	"encoding/binary"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"zinx/zlog"
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

//定义一些全局变量，方便读取
var CURR_PID=os.Getpid()//当前启动的进程ID
var CURR_RUN_PATH=""	//当前运行程序目录  /home/ap/ccb/nitify
var CURR_RUN_NAME=""	//当前运行程序名称	xxx.exe

var NOTIFY_PATH = "notifylog"	//日志监控目录，没有则自动创建

var TRANSFER_PATH = "transferlog"	//传输日志，没有则自动创建

var WAIT_UP_PATH = "waituplog"			//等待上传的记录在这里，临时记录，上传完清空
var BASE_PATH=""			//基础的监控路径/home/nas/static

/*
	提供init方法，默认加载
*/
func init() {
	//初始化全局变量，设置一些默认值
	file, err := exec.LookPath(os.Args[0])
	path, err := filepath.Abs(file)
	if err != nil {
		zlog.Error("获取程序运行目录出错",err)
	}else{
		CURR_RUN_PATH = filepath.Dir(path)
		CURR_RUN_NAME = filepath.Base(path)

		//日志监控目录，没有则自动创建
		NOTIFY_PATH = filepath.Join(CURR_RUN_PATH,"notifylog")
		TRANSFER_PATH = filepath.Join(CURR_RUN_PATH,"transferlog")
		WAIT_UP_PATH= filepath.Join(CURR_RUN_PATH,"waituplog")

		//os.MkdirAll(NOTIFY_PATH,0755)
	}

}

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

//判断文件是否存在
func checkFileExist(filename string) bool {
	exist := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func Mkdir(dir string) (e error) {
	_, er := os.Stat(dir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, 0775); err != nil {
			if os.IsPermission(err) {
				e = err
			}
		}
	}
	return
}

//路径连接，串联
func AppendPath(p1 string, p2 string) string {

	return path.Join(p1, p2)
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

