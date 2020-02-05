package comm

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"zinx/ziface"
	"zinx/znet"
)

//枚举类型的消息
const (
	MID_NullPacket =iota
	MID_KeepAlive

	MID_Login
	MID_LoginRet




	MID_CheckFile	//校验某个文件的MD5
	MID_CheckFileRet //校验文件结果返回

	MID_CheckFileBlock	//校验某个文件的区块
	MID_CheckFileBlockRet //校验某个文件的区块的返回

	MID_SendFile	//上传某个文件
	MID_SendFileBlock	//上传某个文件区块

	MID_SendMessage
)

//消息处理工具类
type MessageUtils struct {

}


func (m *MessageUtils) WriteString(w io.Writer,data string){
	bs :=[]byte(data)
	binary.Write(w, binary.BigEndian, int32(len(bs)))
	w.Write(bs)
}

func (m *MessageUtils) ReadString(r io.Reader) string {
	var n int32
	binary.Read(r,binary.BigEndian,&n)
	bs :=make([]byte,n)
	io.ReadFull(r,bs)
	return string(bs)
}




//各种数据包定义
//1.KeepAliveMsg 客户端/服务器包
type KeepAliveMsg struct {
	time int64
}

func NewKeepAliveMsg(t int64) *KeepAliveMsg{
	return &KeepAliveMsg{
		time:t,
	}
}

func (m *KeepAliveMsg) GetMsg() ziface.IMessage{
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.time)
	return znet.NewMsgPackage(MID_KeepAlive,bytesBuffer.Bytes())
}


//2.LoginMsg 客户端包
type LoginMsg struct {
	MessageUtils
	UserName string
	Pwd string
}

func NewLoginMsg(u string,p string) *LoginMsg{
	return &LoginMsg{
		UserName:u,
		Pwd:p,
	}
}

func NewLoginMsgByByte(b []byte) *LoginMsg{
	bytesBuffer := bytes.NewBuffer(b)
	var m MessageUtils
	return &LoginMsg{
		UserName:m.ReadString(bytesBuffer),
		Pwd:m.ReadString(bytesBuffer),
	}
}

func (m *LoginMsg) GetMsg() ziface.IMessage{
	bytesBuffer := bytes.NewBuffer([]byte{})
	m.WriteString(bytesBuffer,m.UserName)
	m.WriteString(bytesBuffer,m.Pwd)
	return znet.NewMsgPackage(MID_Login,bytesBuffer.Bytes())
}


//3.登录结果 服务器发送到客户端的包
type LoginRetMsg struct {
	Uid  uint32              //用户的ID，sessionid
	Result uint16    //登录结果 0：成功 1：失败
}

func NewLoginRetMsg(uid uint32,r uint16) *LoginRetMsg{
	return &LoginRetMsg{
		Uid:uid,
		Result:r,
	}
}

func NewLoginSuccessByByte(b []byte) *LoginRetMsg{
	var p LoginRetMsg
	err:=json.Unmarshal(b,&p)
	if err!=nil{
		fmt.Println("解码错误",err)
		return nil
	}
	return &p
}

func (m *LoginRetMsg) GetMsg() ziface.IMessage{
	bytes,err:=json.Marshal(m)
	if err!=nil{
		fmt.Println("编码错误",err)
		return nil
	}
	return znet.NewMsgPackage(MID_LoginRet,bytes)
}


//4.CheckFileMsg
type CheckFileMsg struct {
	MessageUtils
	Filepaht  string    //校验文件路径
	Check []byte    	//校验文件MD5
	CheckType byte      //校验文件类型 0:size校验 1:fastmd5 2:fullmd5
}

func NewCheckFileMsg(fp string, ck []byte,ct byte) *CheckFileMsg{
	return &CheckFileMsg{
		Filepaht:fp,
		Check:ck,
		CheckType:ct,
	}
}

func NewCheckFileMsgByByte(b []byte) *CheckFileMsg{
	bytesBuffer := bytes.NewBuffer(b)
	var m MessageUtils
	var c CheckFileMsg
	c.Filepaht=m.ReadString(bytesBuffer)
	c.Check = make([]byte,16)
	bytesBuffer.Read(c.Check)
	c.CheckType,_ = bytesBuffer.ReadByte()
	return &c
}

func (m *CheckFileMsg) GetMsg() ziface.IMessage{
	bytesBuffer := bytes.NewBuffer([]byte{})
	m.WriteString(bytesBuffer,m.Filepaht)
	bytesBuffer.Write(m.Check)
	bytesBuffer.WriteByte(m.CheckType)

	return znet.NewMsgPackage(MID_CheckFile,bytesBuffer.Bytes())
}


//5.CheckFileRetMsg
type CheckFileRetMsg struct {
	MessageUtils
	Filepaht  string    //校验文件路径
	CheckRet byte      //校验文件结果 1：需要上传 2:不需要上传 3:被别的客户端锁定，无需上传
}

func NewCheckFileRetMsg(fp string, ct byte) *CheckFileRetMsg{
	return &CheckFileRetMsg{
		Filepaht:fp,
		CheckRet:ct,
	}
}

func NewCheckFileRetMsgByByte(b []byte) *CheckFileRetMsg{
	bytesBuffer := bytes.NewBuffer(b)
	var m MessageUtils
	var c CheckFileRetMsg
	c.Filepaht=m.ReadString(bytesBuffer)
	c.CheckRet,_ = bytesBuffer.ReadByte()
	return &c
}

func (m *CheckFileRetMsg) GetMsg() ziface.IMessage{
	bytesBuffer := bytes.NewBuffer([]byte{})
	m.WriteString(bytesBuffer,m.Filepaht)
	bytesBuffer.WriteByte(m.CheckRet)
	return znet.NewMsgPackage(MID_CheckFileRet,bytesBuffer.Bytes())
}

