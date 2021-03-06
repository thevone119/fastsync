package comm

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"zinx/ziface"
	"zinx/znet"
)

//这里定义所有的命令字（即所有的socket数据包）

//枚举类型的消息
const (
	MID_NullPacket = iota
	MID_KeepAlive

	MID_Login
	MID_LoginRet

	MID_CheckFile    //校验某个文件的MD5，作废
	MID_CheckFileRet //校验文件结果返回，作废

	MID_SendFileReq    //上传请求，请求进行某个文件上传
	MID_SendFileReqRet //上传请求的返回
	MID_SendFile       //上传某个文件（块）
	MID_SendFileRet    //上传某个文件（块）返回

	MID_Request  //发送req
	MID_Response //返回rsp

	MID_SendMessage

	MID_DeleteFileReq //删除服务器某个文件或者目录

	MID_MoveFileReq //移动服务器某个文件，复制文件，或者目录

	MID_CommRet //通用的返回信息

)
//全局变量，全局对象，避免重复开启对象
var MessageUtilsObj = &MessageUtils{}

//消息处理工具类
type MessageUtils struct {
}

func (m *MessageUtils) WriteString(w io.Writer, data string) {
	bs := []byte(data)
	binary.Write(w, binary.BigEndian, int16(len(bs)))
	w.Write(bs)
}

func (m *MessageUtils) ReadString(r io.Reader) string {
	var n int16
	binary.Read(r, binary.BigEndian, &n)
	bs := make([]byte, n)
	io.ReadFull(r, bs)
	return string(bs)
}

//各种数据包定义
//1.KeepAliveMsg 客户端/服务器包
//ping 包
type KeepAliveMsg struct {
	CTime int64
}

func NewKeepAliveMsg(t int64) *KeepAliveMsg {
	return &KeepAliveMsg{
		CTime: t,
	}
}
func NewKeepAliveMsgByByte(b []byte) *KeepAliveMsg {
	bytesBuffer := bytes.NewBuffer(b)
	km:=KeepAliveMsg{}
	binary.Read(bytesBuffer, binary.BigEndian, &km.CTime)
	return &km
}

func (m *KeepAliveMsg) GetMsg() ziface.IMessage {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.CTime)
	return znet.NewMsgPackage(MID_KeepAlive, bytesBuffer.Bytes())
}

//2.LoginMsg 客户端包
type LoginMsg struct {
	UserName string
	Pwd      string
}

func NewLoginMsg(u string, p string) *LoginMsg {
	return &LoginMsg{
		UserName: u,
		Pwd:      p,
	}
}

func NewLoginMsgByByte(b []byte) *LoginMsg {
	bytesBuffer := bytes.NewBuffer(b)
	return &LoginMsg{
		UserName: MessageUtilsObj.ReadString(bytesBuffer),
		Pwd:      MessageUtilsObj.ReadString(bytesBuffer),
	}
}

func (m *LoginMsg) GetMsg() ziface.IMessage {
	bytesBuffer := bytes.NewBuffer([]byte{})
	MessageUtilsObj.WriteString(bytesBuffer, m.UserName)
	MessageUtilsObj.WriteString(bytesBuffer, m.Pwd)
	return znet.NewMsgPackage(MID_Login, bytesBuffer.Bytes())
}

//3.登录结果 服务器发送到客户端的包
type LoginRetMsg struct {
	Uid    uint32 //用户的ID，sessionid
	Result uint16 //登录结果 0：成功 1：失败
}

func NewLoginRetMsg(uid uint32, r uint16) *LoginRetMsg {
	return &LoginRetMsg{
		Uid:    uid,
		Result: r,
	}
}

func NewLoginSuccessByByte(b []byte) *LoginRetMsg {
	var p LoginRetMsg
	err := json.Unmarshal(b, &p)
	if err != nil {
		fmt.Println("解码错误", err)
		return nil
	}
	return &p
}

func (m *LoginRetMsg) GetMsg() ziface.IMessage {
	bytes, err := json.Marshal(m)
	if err != nil {
		fmt.Println("编码错误", err)
		return nil
	}
	return znet.NewMsgPackage(MID_LoginRet, bytes)
}

//4.CheckFileMsg
type CheckFileMsg struct {
	Filepath  string        //校验文件路径
	Check     []byte        //校验文件MD5（16 byte）
	CheckType CheckFileType //校验文件类型 0:不校验  1:size校验 2:fastmd5 3:fullmd5
}

func NewCheckFileMsg(fp string, ck []byte, ct CheckFileType) *CheckFileMsg {
	return &CheckFileMsg{
		Filepath:  fp,
		Check:     ck,
		CheckType: ct,
	}
}

func NewCheckFileMsgByByte(b []byte) *CheckFileMsg {
	bytesBuffer := bytes.NewBuffer(b)

	var c CheckFileMsg
	c.Filepath = MessageUtilsObj.ReadString(bytesBuffer)
	c.Check = make([]byte, 16)
	bytesBuffer.Read(c.Check)
	ct, _ := bytesBuffer.ReadByte()
	c.CheckType = CheckFileType(ct)
	return &c
}

func (m *CheckFileMsg) GetMsg() ziface.IMessage {
	bytesBuffer := bytes.NewBuffer([]byte{})
	MessageUtilsObj.WriteString(bytesBuffer, m.Filepath)
	bytesBuffer.Write(m.Check)
	bytesBuffer.WriteByte(byte(m.CheckType))

	return znet.NewMsgPackage(MID_CheckFile, bytesBuffer.Bytes())
}

//----------------------------------------------------------------------------------------MID_Request
type RequestMsg struct {
	SecId uint32 //发送请求的ID
	MsgId uint32 //子消息ID
	Data  []byte //返回的数据包
}

func NewRequestMsgMsg(sendid uint32, msgid uint32, data []byte) *RequestMsg {
	return &RequestMsg{
		SecId: sendid,
		MsgId: msgid,
		Data:  data,
	}
}

func NewRequestMsgMsgByByte(b []byte) *RequestMsg {
	bytesBuffer := bytes.NewBuffer(b)
	var c RequestMsg
	binary.Read(bytesBuffer, binary.BigEndian, &c.SecId)
	binary.Read(bytesBuffer, binary.BigEndian, &c.MsgId)
	c.Data = make([]byte, bytesBuffer.Len())
	bytesBuffer.Read(c.Data)
	return &c
}

func (m *RequestMsg) GetMsg() ziface.IMessage {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.SecId)
	binary.Write(bytesBuffer, binary.BigEndian, m.MsgId)
	bytesBuffer.Write(m.Data)
	return znet.NewMsgPackage(MID_Request, bytesBuffer.Bytes())
}

//----------------------------------------------------------------------------------------MID_Request
type ResponseMsg struct {
	SecId uint32 //发送请求的ID
	MsgId uint32 //子消息ID
	Data  []byte //返回的数据包
}

func NewResponseMsg(sendid uint32, msgid uint32, data []byte) *ResponseMsg {
	return &ResponseMsg{
		SecId: sendid,
		MsgId: msgid,
		Data:  data,
	}
}

func NewResponseMsgByByte(b []byte) *ResponseMsg {
	bytesBuffer := bytes.NewBuffer(b)
	var c ResponseMsg
	binary.Read(bytesBuffer, binary.BigEndian, &c.SecId)
	binary.Read(bytesBuffer, binary.BigEndian, &c.MsgId)
	c.Data = make([]byte, bytesBuffer.Len())
	bytesBuffer.Read(c.Data)
	return &c
}

func (m *ResponseMsg) GetMsg() ziface.IMessage {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.SecId)
	binary.Write(bytesBuffer, binary.BigEndian, m.MsgId)
	bytesBuffer.Write(m.Data)
	return znet.NewMsgPackage(MID_Response, bytesBuffer.Bytes())
}

//----------------------------------------------------------------------------------------CheckFileRetMsg
type CheckFileRetMsg struct {
	Filepaht string //校验文件路径
	CheckRet byte   //校验文件结果 1：需要上传 2:不需要上传 3:被别的客户端锁定，无需上传
}

func NewCheckFileRetMsg(fp string, ct byte) *CheckFileRetMsg {
	return &CheckFileRetMsg{
		Filepaht: fp,
		CheckRet: ct,
	}
}

func NewCheckFileRetMsgByByte(b []byte) *CheckFileRetMsg {
	bytesBuffer := bytes.NewBuffer(b)
	var c CheckFileRetMsg
	c.Filepaht = MessageUtilsObj.ReadString(bytesBuffer)
	c.CheckRet, _ = bytesBuffer.ReadByte()
	return &c
}

func (m *CheckFileRetMsg) GetMsg() ziface.IMessage {
	bytesBuffer := bytes.NewBuffer([]byte{})
	MessageUtilsObj.WriteString(bytesBuffer, m.Filepaht)
	bytesBuffer.WriteByte(m.CheckRet)
	return znet.NewMsgPackage(MID_CheckFileRet, bytesBuffer.Bytes())
}

//----------------------------------------------------------------------------------------MID_SendFileReq 发送上传文件请求
type SendFileReqMsg struct {
	ReqId        uint32        //请求的ID
	Flen         int64         //文件大小
	FlastModTime int64         //文件的最后修改时间,秒
	Check        []byte        //校验文件MD5（16 byte）
	CheckType    CheckFileType //校验文件类型 0:不校验  1:size校验 2:fastmd5 3:fullmd5
	IsUpload     byte          //是否开启上传通道
	Filepath     string        //目标文件路径
}

func NewSendFileReqMsg(reqid uint32, fl int64, modtime int64, cbyte []byte, ctype CheckFileType, isupload byte, fp string) *SendFileReqMsg {
	fp=strings.Replace(fp,"\\","/",-1)
	return &SendFileReqMsg{
		ReqId:        reqid,
		Flen:         fl,
		FlastModTime: modtime,
		Check:        cbyte,
		CheckType:    ctype,
		IsUpload:     isupload,
		Filepath:     fp,
	}
}

func NewSendFileReqMsgByByte(b []byte) *SendFileReqMsg {
	bytesBuffer := bytes.NewBuffer(b)

	var c SendFileReqMsg
	binary.Read(bytesBuffer, binary.BigEndian, &c.ReqId)
	binary.Read(bytesBuffer, binary.BigEndian, &c.Flen)
	binary.Read(bytesBuffer, binary.BigEndian, &c.FlastModTime)

	c.Check = make([]byte, 16)
	bytesBuffer.Read(c.Check)
	ct, _ := bytesBuffer.ReadByte()
	c.CheckType = CheckFileType(ct)
	c.IsUpload, _ = bytesBuffer.ReadByte()
	c.Filepath = MessageUtilsObj.ReadString(bytesBuffer)
	return &c
}

func (m *SendFileReqMsg) GetMsg() ziface.IMessage {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.ReqId)
	binary.Write(bytesBuffer, binary.BigEndian, m.Flen)
	binary.Write(bytesBuffer, binary.BigEndian, m.FlastModTime)
	bytesBuffer.Write(m.Check)
	bytesBuffer.WriteByte(byte(m.CheckType))
	bytesBuffer.WriteByte(m.IsUpload)
	MessageUtilsObj.WriteString(bytesBuffer, m.Filepath)
	return znet.NewMsgPackage(MID_SendFileReq, bytesBuffer.Bytes())
}

//-------------------------------------------------------------SendFileReqRetMsg 发送上传文件请求，返回
type SendFileReqRetMsg struct {
	ReqId   uint32 //请求的ID
	RetId   uint32 //返回的ID
	RetCode byte   //返回码，做相关逻辑的 0:可以上传，1：io失败，无法上传，2：文件一致，无需上传 3:文件正被其他客户端上传锁定 4：未登录，无法上传
}

func NewSendFileReqRetMsg(reqid uint32, retid uint32, retcode byte) *SendFileReqRetMsg {
	return &SendFileReqRetMsg{
		ReqId:   reqid,
		RetId:   retid,
		RetCode: retcode,
	}
}

func NewSendFileReqRetMsgByByte(b []byte) *SendFileReqRetMsg {
	bytesBuffer := bytes.NewBuffer(b)
	var c SendFileReqRetMsg
	binary.Read(bytesBuffer, binary.BigEndian, &c.ReqId)
	binary.Read(bytesBuffer, binary.BigEndian, &c.RetId)
	c.RetCode, _ = bytesBuffer.ReadByte()
	return &c
}

func (m *SendFileReqRetMsg) GetMsg() ziface.IMessage {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.ReqId)
	binary.Write(bytesBuffer, binary.BigEndian, m.RetId)
	bytesBuffer.WriteByte(m.RetCode)
	return znet.NewMsgPackage(MID_SendFileReqRet, bytesBuffer.Bytes())
}

//-------------------------------------------------------------MID_SendFile 上传某个文件，一次上传4K
type SendFileMsg struct {
	SecId  uint32 //消息序列号ID
	FileId uint32 //文件句柄ID
	Start  int64  //开始位置
	Fbyte  []byte //文件二进制
}

func NewSendFileMsg(secid uint32, fildid uint32, start int64, fb []byte) *SendFileMsg {
	return &SendFileMsg{
		SecId:  secid,
		FileId: fildid,
		Start:  start,
		Fbyte:  fb,
	}
}

func NewSendFileMsgByByte(b []byte) *SendFileMsg {
	bytesBuffer := bytes.NewBuffer(b)
	var c SendFileMsg
	binary.Read(bytesBuffer, binary.BigEndian, &c.SecId)
	binary.Read(bytesBuffer, binary.BigEndian, &c.FileId)
	binary.Read(bytesBuffer, binary.BigEndian, &c.Start)
	c.Fbyte = make([]byte, bytesBuffer.Len())
	io.ReadFull(bytesBuffer, c.Fbyte)
	return &c
}

func (m *SendFileMsg) GetMsg() ziface.IMessage {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.SecId)
	binary.Write(bytesBuffer, binary.BigEndian, m.FileId)
	binary.Write(bytesBuffer, binary.BigEndian, m.Start)
	bytesBuffer.Write(m.Fbyte)
	return znet.NewMsgPackage(MID_SendFile, bytesBuffer.Bytes())
}

//-------------------------------------------------------------MID_SendFileRet 上传某个文件块，返回
type SendFileRetMsg struct {
	SecId   uint32 //消息序列号ID
	FileId  uint32 //文件句柄ID
	Start   int64  //开始位置
	RetCode byte   //返回码 0:未成功，1：成功  2:服务器读写错误 3：传输完成（最后的块都传输完了）
}

func NewSendFileRetMsg(secid uint32, fileid uint32, start int64, retcode byte) *SendFileRetMsg {
	return &SendFileRetMsg{
		SecId:   secid,
		FileId:  fileid,
		Start:   start,
		RetCode: retcode,
	}
}

func NewSendFileRetMsgByByte(b []byte) *SendFileRetMsg {
	bytesBuffer := bytes.NewBuffer(b)
	var c SendFileRetMsg
	binary.Read(bytesBuffer, binary.BigEndian, &c.SecId)
	binary.Read(bytesBuffer, binary.BigEndian, &c.FileId)
	binary.Read(bytesBuffer, binary.BigEndian, &c.Start)
	c.RetCode, _ = bytesBuffer.ReadByte()
	return &c
}

func (m *SendFileRetMsg) GetMsg() ziface.IMessage {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.SecId)
	binary.Write(bytesBuffer, binary.BigEndian, m.FileId)
	binary.Write(bytesBuffer, binary.BigEndian, m.Start)
	bytesBuffer.WriteByte(m.RetCode)
	return znet.NewMsgPackage(MID_SendFileRet, bytesBuffer.Bytes())
}

//-------------------------------------------------------------MID_DeleteFileReq 删除某个文件/或者目录
type DeleteFileReqMsg struct {
	SecId    uint32 //消息序列号ID
	FileType byte   //0:文件  1：目录
	Filepath string //目标文件路径
}

func NewDeleteFileRetMsg(secid uint32, fileType byte, fp string) *DeleteFileReqMsg {
	return &DeleteFileReqMsg{
		SecId:    secid,
		FileType: fileType,
		Filepath: fp,
	}
}

func NewDeleteFileReqMsgByByte(b []byte) *DeleteFileReqMsg {
	bytesBuffer := bytes.NewBuffer(b)
	var c DeleteFileReqMsg
	binary.Read(bytesBuffer, binary.BigEndian, &c.SecId)
	c.FileType, _ = bytesBuffer.ReadByte()
	c.Filepath = MessageUtilsObj.ReadString(bytesBuffer)
	return &c
}

func (m *DeleteFileReqMsg) GetMsg() ziface.IMessage {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.SecId)
	bytesBuffer.WriteByte(m.FileType)
	MessageUtilsObj.WriteString(bytesBuffer, m.Filepath)
	return znet.NewMsgPackage(MID_DeleteFileReq, bytesBuffer.Bytes())
}

//-------------------------------------------------------------MID_MoveFileReq 删除某个文件/或者目录
type MoveFileReqMsg struct {
	SecId       uint32 //消息序列号ID
	OpType      byte   //0:copy  1：move
	SrcFilepath string //源文件路径
	DstFilepath string //目标文件路径
}

func NewMoveFileReqMsg(secid uint32, opType byte, fp string, dstFilepath string) *MoveFileReqMsg {
	return &MoveFileReqMsg{
		SecId:       secid,
		OpType:      opType,
		SrcFilepath: fp,
		DstFilepath: dstFilepath,
	}
}

func NewMoveFileReqMsgByByte(b []byte) *MoveFileReqMsg {
	bytesBuffer := bytes.NewBuffer(b)
	var c MoveFileReqMsg
	binary.Read(bytesBuffer, binary.BigEndian, &c.SecId)
	c.OpType, _ = bytesBuffer.ReadByte()
	c.SrcFilepath = MessageUtilsObj.ReadString(bytesBuffer)
	c.DstFilepath = MessageUtilsObj.ReadString(bytesBuffer)
	return &c
}

func (m *MoveFileReqMsg) GetMsg() ziface.IMessage {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.SecId)
	bytesBuffer.WriteByte(m.OpType)
	MessageUtilsObj.WriteString(bytesBuffer, m.SrcFilepath)
	MessageUtilsObj.WriteString(bytesBuffer, m.DstFilepath)
	return znet.NewMsgPackage(MID_MoveFileReq, bytesBuffer.Bytes())
}

//-------------------------------------------------------------MID_CommRet 通用的返回信息
type CommRetMsg struct {
	ReqId   uint32 //请求的ID
	RetCode int16  //返回码 0:成功，1：失败  2:服务器错误
	RetMsg  string //返回消息
	ExtInt  int64  //扩展字段int
	ExtStr  string //扩展字段string
}

func NewCommRetMsg(reqId uint32, retCode int16, retMsg string, extInt int64, extStr string) *CommRetMsg {
	return &CommRetMsg{
		ReqId:   reqId,
		RetCode: retCode,
		RetMsg:  retMsg,
		ExtInt:  extInt,
		ExtStr:  extStr,
	}
}

func NewCommRetMsgByByte(b []byte) *CommRetMsg {
	bytesBuffer := bytes.NewBuffer(b)
	var c CommRetMsg
	binary.Read(bytesBuffer, binary.BigEndian, &c.ReqId)
	binary.Read(bytesBuffer, binary.BigEndian, &c.RetCode)
	binary.Read(bytesBuffer, binary.BigEndian, &c.ExtInt)

	c.RetMsg = MessageUtilsObj.ReadString(bytesBuffer)
	c.ExtStr = MessageUtilsObj.ReadString(bytesBuffer)
	return &c
}

func (m *CommRetMsg) GetMsg() ziface.IMessage {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, m.ReqId)
	binary.Write(bytesBuffer, binary.BigEndian, m.RetCode)
	binary.Write(bytesBuffer, binary.BigEndian, m.ExtInt)
	MessageUtilsObj.WriteString(bytesBuffer, m.RetMsg)
	MessageUtilsObj.WriteString(bytesBuffer, m.ExtStr)
	return znet.NewMsgPackage(MID_CommRet, bytesBuffer.Bytes())
}
