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


	MID_SendFileReq	//上传请求，请求进行某个文件上传
	MID_SendFileReqRet	//上传请求的返回
	MID_SendFile	//上传某个文件（块）
	MID_SendFileRet	//上传某个文件（块）返回

	MID_Request		//发送req
	MID_Response	//返回rsp

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
	Check []byte    	//校验文件MD5（16 byte）
	CheckType byte      //校验文件类型 0:不校验  1:size校验 2:fastmd5 3:fullmd5
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


//----------------------------------------------------------------------------------------MID_Request
type RequestMsg struct {
	MessageUtils
	SecId  uint32    //发送请求的ID
	MsgId uint32 	//子消息ID
	Data []byte    	//返回的数据包
}

func NewRequestMsgMsg(sendid uint32,msgid uint32, data []byte) *RequestMsg{
	return &RequestMsg{
		SecId:sendid,
		MsgId:msgid,
		Data:data,
	}
}

func NewRequestMsgMsgByByte(b []byte) *RequestMsg{
	bytesBuffer := bytes.NewBuffer(b)
	var c RequestMsg
	binary.Read(bytesBuffer,binary.BigEndian,&c.SecId)
	binary.Read(bytesBuffer,binary.BigEndian,&c.MsgId)
	c.Data = make([]byte,bytesBuffer.Len())
	bytesBuffer.Read(c.Data)
	return &c
}

func (m *RequestMsg) GetMsg() ziface.IMessage{
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.SecId)
	binary.Write(bytesBuffer, binary.BigEndian, m.MsgId)
	bytesBuffer.Write(m.Data)
	return znet.NewMsgPackage(MID_Request,bytesBuffer.Bytes())
}




//----------------------------------------------------------------------------------------MID_Request
type ResponseMsg struct {
	MessageUtils
	SecId  uint32    //发送请求的ID
	MsgId uint32 	//子消息ID
	Data []byte    	//返回的数据包
}

func NewResponseMsg(sendid uint32,msgid uint32, data []byte) *ResponseMsg{
	return &ResponseMsg{
		SecId:sendid,
		MsgId:msgid,
		Data:data,
	}
}

func NewResponseMsgByByte(b []byte) *ResponseMsg{
	bytesBuffer := bytes.NewBuffer(b)
	var c ResponseMsg
	binary.Read(bytesBuffer,binary.BigEndian,&c.SecId)
	binary.Read(bytesBuffer,binary.BigEndian,&c.MsgId)
	c.Data = make([]byte,bytesBuffer.Len())
	bytesBuffer.Read(c.Data)
	return &c
}

func (m *ResponseMsg) GetMsg() ziface.IMessage{
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.SecId)
	binary.Write(bytesBuffer, binary.BigEndian, m.MsgId)
	bytesBuffer.Write(m.Data)
	return znet.NewMsgPackage(MID_Response,bytesBuffer.Bytes())
}




//----------------------------------------------------------------------------------------CheckFileRetMsg
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



//----------------------------------------------------------------------------------------MID_SendFileReq 发送上传文件请求
type SendFileReqMsg struct {
	MessageUtils
	ReqId uint32		//请求的ID
	Flen int64		//文件大小
	Check []byte    	//校验文件MD5（16 byte）
	CheckType byte      //校验文件类型 0:不校验  1:size校验 2:fastmd5 3:fullmd5
	IsUpload byte      //是否开启上传通道
	Filepaht  string    //目标文件路径
}

func NewSendFileReqMsg(reqid uint32,fl int64,cbyte []byte,ctype byte,isupload byte,fp string) *SendFileReqMsg{
	return &SendFileReqMsg{
		ReqId:reqid,
		Flen:fl,
		Check:cbyte,
		CheckType:ctype,
		IsUpload:isupload,
		Filepaht:fp,
	}
}

func NewSendFileReqMsgByByte(b []byte) *SendFileReqMsg{
	bytesBuffer := bytes.NewBuffer(b)
	var m MessageUtils
	var c SendFileReqMsg
	binary.Read(bytesBuffer,binary.BigEndian,&c.ReqId)
	binary.Read(bytesBuffer,binary.BigEndian,&c.Flen)
	c.Check = make([]byte,16)
	bytesBuffer.Read(c.Check)
	c.CheckType,_ = bytesBuffer.ReadByte()
	c.IsUpload,_ = bytesBuffer.ReadByte()
	c.Filepaht=m.ReadString(bytesBuffer)
	return &c
}

func (m *SendFileReqMsg) GetMsg() ziface.IMessage{
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.ReqId)
	binary.Write(bytesBuffer, binary.BigEndian, m.Flen)
	bytesBuffer.Write(m.Check)
	bytesBuffer.WriteByte(m.CheckType)
	bytesBuffer.WriteByte(m.IsUpload)
	m.WriteString(bytesBuffer,m.Filepaht)
	return znet.NewMsgPackage(MID_SendFileReq,bytesBuffer.Bytes())
}







//6.SendFileReqRetMsg 发送上传文件请求，返回
type SendFileReqRetMsg struct {
	ReqId uint32		//请求的ID
	RetId uint32		//返回的ID
	RetCode byte		//返回码，做相关逻辑的 0:可以上传，1：io失败，无法上传，2：文件一致，无需上传
}

func NewSendFileReqRetMsg(reqid uint32,retid uint32, retcode byte) *SendFileReqRetMsg{
	return &SendFileReqRetMsg{
		ReqId:reqid,
		RetId:retid,
		RetCode:retcode,
	}
}

func NewSendFileReqRetMsgByByte(b []byte) *SendFileReqRetMsg{
	bytesBuffer := bytes.NewBuffer(b)
	var c SendFileReqRetMsg
	binary.Read(bytesBuffer,binary.BigEndian,&c.ReqId)
	binary.Read(bytesBuffer,binary.BigEndian,&c.RetId)
	c.RetCode,_ = bytesBuffer.ReadByte()
	return &c
}

func (m *SendFileReqRetMsg) GetMsg() ziface.IMessage{
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.ReqId)
	binary.Write(bytesBuffer, binary.BigEndian, m.RetId)
	bytesBuffer.WriteByte(m.RetCode)
	return znet.NewMsgPackage(MID_SendFileReqRet,bytesBuffer.Bytes())
}





//-------------------------------------------------------------MID_SendFile 上传某个文件，一次上传4K
type SendFileMsg struct {
	SecId uint32 	//消息序列号ID
	FileId uint32 	//文件句柄ID
	Start int64		//开始位置
	Fbyte []byte	//文件二进制
}

func NewSendFileMsg(secid uint32,fildid uint32, start int64,fb []byte) *SendFileMsg{
	return &SendFileMsg{
		SecId:secid,
		FileId:fildid,
		Start:start,
		Fbyte:fb,
	}
}

func NewSendFileMsgByByte(b []byte) *SendFileMsg{
	bytesBuffer := bytes.NewBuffer(b)
	var c SendFileMsg
	binary.Read(bytesBuffer,binary.BigEndian,&c.SecId)
	binary.Read(bytesBuffer,binary.BigEndian,&c.FileId)
	binary.Read(bytesBuffer,binary.BigEndian,&c.Start)
	c.Fbyte=make([]byte,bytesBuffer.Len())
	io.ReadFull(bytesBuffer,c.Fbyte)
	return &c
}

func (m *SendFileMsg) GetMsg() ziface.IMessage{
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.SecId)
	binary.Write(bytesBuffer, binary.BigEndian, m.FileId)
	binary.Write(bytesBuffer, binary.BigEndian, m.Start)
	bytesBuffer.Write(m.Fbyte)
	return znet.NewMsgPackage(MID_SendFile,bytesBuffer.Bytes())
}


//-------------------------------------------------------------MID_SendFileRet 上传某个文件块，返回
type SendFileRetMsg struct {
	SecId uint32 	//消息序列号ID
	FileId uint32 	//文件句柄ID
	Start int64		//开始位置
	RetCode byte		//返回码 0:未成功，1：成功
}

func NewSendFileRetMsg(secid uint32,fileid uint32,start int64, retcode byte) *SendFileRetMsg{
	return &SendFileRetMsg{
		SecId:secid,
		FileId:fileid,
		Start:start,
		RetCode:retcode,
	}
}

func NewSendFileRetMsgByByte(b []byte) *SendFileRetMsg{
	bytesBuffer := bytes.NewBuffer(b)
	var c SendFileRetMsg
	binary.Read(bytesBuffer,binary.BigEndian,&c.SecId)
	binary.Read(bytesBuffer,binary.BigEndian,&c.FileId)
	binary.Read(bytesBuffer,binary.BigEndian,&c.Start)
	c.RetCode,_ = bytesBuffer.ReadByte()
	return &c
}

func (m *SendFileRetMsg) GetMsg() ziface.IMessage{
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.SecId)
	binary.Write(bytesBuffer, binary.BigEndian, m.FileId)
	binary.Write(bytesBuffer, binary.BigEndian, m.Start)
	bytesBuffer.WriteByte(m.RetCode)
	return znet.NewMsgPackage(MID_SendFileRet,bytesBuffer.Bytes())
}

