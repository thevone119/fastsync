/**********************************************
** @Des: This file ...
** @Author: haodaquan
** @Date:   2017-09-08 00:24:25
** @Last Modified by:   haodaquan
** @Last Modified time: 2017-09-17 10:12:06
***********************************************/

package libs

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"regexp"
	"time"
)

var emailPattern = regexp.MustCompile("[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[a-zA-Z0-9](?:[\\w-]*[\\w])?")

func Md5(buf []byte) string {
	hash := md5.New()
	hash.Write(buf)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func SizeFormat(size float64) string {
	units := []string{"Byte", "KB", "MB", "GB", "TB"}
	n := 0
	for size > 1024 {
		size /= 1024
		n += 1
	}
	return fmt.Sprintf("%.2f %s", size, units[n])
}

//秒
func TimeFormat(size int64) string {
	d:=time.Duration(size)*time.Second
	if d.Hours()<1{
		return fmt.Sprintf("%d分%d秒", int(d.Minutes()), int64(d.Seconds())%60)
	} else if d.Hours()<24{
		return fmt.Sprintf("%d时%d分%d秒", int64(d.Hours()),int64(d.Minutes())%60, int64(d.Seconds())%60)
	}else{
		return fmt.Sprintf("%d天%d时%d分", int64(d.Hours())/24, int64(d.Hours())%24,int64(d.Minutes())%60)
	}
}

func IsEmail(b []byte) bool {
	return emailPattern.Match(b)
}

func Password(len int, pwdO string) (pwd string, salt string) {
	salt = GetRandomString(4)
	defaultPwd := "george518"
	if pwdO != "" {
		defaultPwd = pwdO
	}
	pwd = Md5([]byte(defaultPwd + salt))
	return pwd, salt
}

// 生成32位MD5
// func MD5(text string) string{
//    ctx := md5.New()
//    ctx.Write([]byte(text))
//    return hex.EncodeToString(ctx.Sum(nil))
// }

//生成随机字符串
func GetRandomString(lens int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < lens; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
