package client

import (
	"comm"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"sync"
	"time"
	"zinx/ziface"
	"zinx/zlog"
	"zinx/znet"
)

//网络数据发送，处理
type NetWork struct {
	//服务绑定的IP地址
	IP string
	//服务绑定的端口
	Port int
	Connected bool

	//告知该链接已经退出/停止的channel
	ExitBuffChan chan bool
	sendMsgs chan ziface.IMessage //发送消息管道 1024
	receive chan ziface.IMessage //接受到消息的管道 1024
	//当前连接的socket TCP套接字
	Conn net.Conn
	TimeOutTime int64
	TimeConnected int64
	//是否已登录
	Login bool

	receiveCallBack           map[uint32] func(ziface.IMessage)  //存放每个MsgId 处理方法

	//独立的request通道，实现同步的request
	requestChanMutex sync.RWMutex
	requestChan           map[uint32] chan *comm.ResponseMsg  //request的管道，每个请求一个管道，知道管道中有数据，才返回

	secId uint32
	secIdMutex sync.Mutex
}

//一个新的网络
func NewNetWork(ip string,port int) *NetWork{
	n:=NetWork{
		IP:ip,
		Port:port,
		Connected:false,
		Login:false,
		ExitBuffChan: make(chan bool, 1),
		sendMsgs: make(chan ziface.IMessage, 20),//缓存20个数据包 每个4K算的话，就是8M缓存
		receive: make(chan ziface.IMessage, 1024),
		TimeOutTime:0,
		TimeConnected:0,
		Conn:nil,
		receiveCallBack: make(map[uint32]func(ziface.IMessage)),
		requestChan: make(map[uint32] chan *comm.ResponseMsg),
	}

	//加入2个默认的处理
	n.AddCallBack(comm.MID_KeepAlive,n.doKeepalive)
	n.AddCallBack(comm.MID_LoginRet,n.doLoginRet)
	n.AddCallBack(comm.MID_Response,n.doResponse)
	return &n
}


//连接网络
func (n *NetWork) connect(){
	n.TimeConnected = time.Now().Unix()
	n.TimeOutTime= time.Now().Unix()
	adder :=fmt.Sprintf("%s:%d", n.IP, n.Port)
	fmt.Println("connect to:",adder)
	conn,err := net.Dial("tcp", adder)
	if err != nil {
		fmt.Println("NetWork Connect err!",adder)
		return
	}
	//登录认证
	n.Enqueue(comm.NewLoginMsg("admin","admin").GetMsg())


	fmt.Println("NetWork Connect succ!",n.IP)
	n.Conn=conn
	n.Connected = true

	//开启读写线程
	go n.receiveData()
	go n.sendData()
}

//断开连接
func (n *NetWork) disconnect(){
	n.Connected = false
	n.Login = false
	n.ExitBuffChan<-true
	n.Conn.Close()
	//close(n.SendMsgs)
	//close(n.Receive)
}



//死循环处理,主线程死循环调用,每秒循环调用
func (n *NetWork) Process(){
	for {
		currtime:=time.Now().Unix()
		//超过5秒没有连接上，则再次发起连接？
		if (!n.Connected &&  currtime > n.TimeConnected+5 ){
			n.connect()
		}
		time.Sleep(1 * time.Second)

		if n.Connected == false{
			continue
		}
		//超时发送心跳包？每5秒发送一个？心跳包是空的？
		if (currtime> n.TimeOutTime ){
			n.TimeOutTime=currtime+5
			n.Enqueue(comm.NewKeepAliveMsg(currtime).GetMsg())
		}
		//如果有接受管道中有数据，则开启线程处理管道中的数据

	}
}

//把数据放入待发送队列
func (n *NetWork) Enqueue(msg ziface.IMessage){
	n.sendMsgs<-msg
}

//go线程调用
func (n *NetWork) sendData(){
	//发封包message消息
	dp := znet.NewDataPack()
	for {
		if n.Connected == false{
			return
		}
		select {
		case data, ok := <-n.sendMsgs:
			if ok {
				//有数据要发送
				_d,_:=dp.Pack(data)
				if _, err := n.Conn.Write(_d); err != nil {
					zlog.Error("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				break
				zlog.Info("msgBuffChan is Closed")
			}
		case <-n.ExitBuffChan:
			return
		}
	}
}


//接受到数据,这个方法不对外,go方法，循环读取数据
func (n *NetWork) receiveData(){
	defer n.disconnect()
	for {
		if n.Connected==false{
			return
		}
		//发封包message消息
		dp := znet.NewDataPack()

		//先读出流中的head部分
		headData := make([]byte, dp.GetHeadLen())
		_, err := io.ReadFull(n.Conn, headData) //ReadFull 会把msg填充满为止
		if err != nil {
			fmt.Println("read head error")
			break
		}
		//将headData字节流 拆包到msg中
		msgHead, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("server unpack err:", err)
			return
		}

		if msgHead.GetDataLen() > 0 {
			//msg 是有data数据的，需要再次读取data数据
			msg := msgHead.(*znet.Message)
			msg.Data = make([]byte, msg.GetDataLen())

			//根据dataLen从io中读取字节流
			_, err := io.ReadFull(n.Conn, msg.Data)
			if err != nil {
				fmt.Println("server unpack data err:", err)
				return
			}
			//不放管道了，接受到所有数据都直接处理
			//n.Receive<-msg
			n.doReceiveData(msg)
		}
	}
}

func (n *NetWork) AddCallBack(msgid uint32,cb func(ziface.IMessage)){
	n.receiveCallBack[msgid]=cb
}
func (n *NetWork) RemoveCallBack(msgid uint32,cb func(ziface.IMessage)){
	delete(n.receiveCallBack,msgid)
}

func (n *NetWork) doKeepalive(msg ziface.IMessage){
	fmt.Println("Receive keepalive back")
}


func (n *NetWork) doResponse(msg ziface.IMessage){
	ret :=comm.NewResponseMsgByByte(msg.GetData())
	if ret!=nil{
		//zlog.Debug("request back secid:",ret.SecId)
		n.requestChanMutex.RLock()
		n.requestChan[ret.SecId]<-ret
		n.requestChanMutex.RUnlock()
	}
}

func (n *NetWork) doLoginRet(msg ziface.IMessage){
	lret :=comm.NewLoginSuccessByByte(msg.GetData())
	if lret!=nil{
		if(lret.Result==0){
			n.Login = true
			fmt.Println("Login succ")
		}else{
			n.Login = false
		}
	}
}

//所有数据在这里处理
func (n *NetWork) doReceiveData(msg ziface.IMessage){
	handler, ok := n.receiveCallBack[msg.GetMsgId()]
	if !ok {
		zlog.Debug("Receive data msgId = ", msg.GetMsgId(), " is not handler func !")
		return
	}else{
		handler(msg)
	}
}

//发送请求，获得返回，柱塞
//10秒超时
func (n *NetWork) Request(msg ziface.IMessage) ([]byte,error){
	n.secIdMutex.Lock()
	if n.secId>=math.MaxUint32{
		n.secId=1
	}
	n.secId++
	_secId:=n.secId
	n.secIdMutex.Unlock()
	//发送
	n.Enqueue(comm.NewRequestMsgMsg(_secId,msg.GetMsgId(),msg.GetData()).GetMsg())
	//锁
	n.requestChanMutex.Lock()
	n.requestChan[_secId]=make(chan *comm.ResponseMsg)
	n.requestChanMutex.Unlock()
	//柱塞
	select {
		case data, ok := <-n.requestChan[_secId]:
			if ok{
				//关闭通道，释放资源
				close(n.requestChan[_secId])
				n.requestChanMutex.Lock()
				delete(n.requestChan,_secId)
				n.requestChanMutex.Unlock()
				return data.Data,nil
			}
		case <-time.After((time.Second * 2))://10秒超时
			fmt.Println("request time out",_secId)
			return nil,errors.New("request time out")

	}
	return nil,errors.New("request time out")
}






