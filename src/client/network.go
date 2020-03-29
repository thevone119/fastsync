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
	//id,按配置循序来，0-100？
	Id int
	//名称
	Name string
	//服务绑定的IP地址
	IP string
	//服务绑定的端口
	Port      int
	UserName  string
	PassWord  string
	Connected bool

	//ping time

	//告知该链接已经退出/停止的channel
	ExitBuffChan chan bool
	sendMsgs     chan ziface.IMessage //发送消息管道 1024
	receive      chan ziface.IMessage //接受到消息的管道 1024
	//当前连接的socket TCP套接字
	Conn          net.Conn
	//统计信息
	ConnCount      int64	//连接次数，掉线重新连次数
	SuccConnCount	int64	//成功连接次数
	PingTime      int64	//ping的时长，毫秒
	MaxPingTime	  int64	//最大的ping值
	OnlineTime	  int64	//在线时长，秒
	OfflineTime   int64 //离线时长，秒

	//逻辑处理
	TimeOutTime   int64
	TimeConnected int64
	ActivityTime  int64 //活动时间，存活时间，如果超过5秒没有活动了。则认为这个连接有问题了哦
	CurrTime      int64 //当前的时间，秒，每秒循环，更新这个时间
	//是否已登录
	Login bool

	receiveCallBack map[uint32]func(ziface.IMessage) //存放每个MsgId 处理方法

	//独立的request通道，实现同步的request
	requestChanLock chan bool
	requestChan     []chan *comm.ResponseMsg //request的管道，10个管道并发

	dataPack *znet.DataPack //数据封包处理类，避免每次都实例化

	secId      uint32
	secIdMutex sync.Mutex
}

//一个新的网络
func NewNetWork(id int,ip string, port int, username string, password string,name string) *NetWork {
	currtime:=time.Now().Unix()
	n := NetWork{
		Id:id,
		Name :			name,
		IP:              ip,
		Port:            port,
		UserName:        username,
		PassWord:        password,
		Connected:       false,
		Login:           false,
		ExitBuffChan:    make(chan bool, 1),
		sendMsgs:        make(chan ziface.IMessage, 10), //缓存10个数据包 每个4K算的话，就是40K缓存
		receive:         make(chan ziface.IMessage, 10),
		TimeOutTime:     currtime,
		TimeConnected:   currtime,
		ActivityTime:    currtime,
		CurrTime:			currtime,
		Conn:            nil,
		dataPack:        znet.NewDataPack(),
		receiveCallBack: make(map[uint32]func(ziface.IMessage)),
		requestChanLock: make(chan bool, 10),
		requestChan:     make([]chan *comm.ResponseMsg, 10),
	}
	//10个阻塞管道
	for i := 0; i < 10; i++ {
		n.requestChan[i] = make(chan *comm.ResponseMsg)
	}

	//加入几个默认的处理
	n.AddCallBack(comm.MID_KeepAlive, n.doKeepalive)
	n.AddCallBack(comm.MID_LoginRet, n.doLoginRet)
	n.AddCallBack(comm.MID_Response, n.doResponse)
	//启动线程做自处理，保持会话，链接
	n.connect()
	go n.process()
	return &n
}

//连接网络
func (n *NetWork) connect() {
	n.TimeConnected = time.Now().Unix()
	n.TimeOutTime = n.TimeConnected
	adder := fmt.Sprintf("%s:%d", n.IP, n.Port)
	zlog.Debug("connect to:", adder)
	n.ConnCount++
	conn, err := net.DialTimeout("tcp", adder, time.Duration(3)*time.Second)
	if err != nil {
		zlog.Error("NetWork Connect err!", adder)
		return
	}
	//这里才重新开启管道哦
	n.sendMsgs = make(chan ziface.IMessage, 10)
	n.receive = make(chan ziface.IMessage, 10)
	zlog.Debug("NetWork Connect succ!", n.IP)
	n.Conn = conn
	n.Connected = true
	n.ActivityTime = n.TimeConnected
	//登录认证
	n.SendData(comm.NewLoginMsg(n.UserName, n.PassWord).GetMsg())
	n.SuccConnCount++

	//开启读写线程
	go n.receiveData()
	go n.gosendData()
}

//断开连接
func (n *NetWork) Disconnect() {
	n.Connected = false
	n.Login = false
	n.ExitBuffChan <- true
	n.Conn.Close()

	//关闭管道
	close(n.sendMsgs)
	close(n.receive)
}

//死循环处理,主线程死循环调用,每秒循环调用
func (n *NetWork) process() {
	if !n.Connected {
		n.connect()
	}
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			if n.Connected{
				n.OnlineTime++
			}else{
				n.OfflineTime++
			}
			n.CurrTime = time.Now().Unix()
			//超过5秒没有连接上，则再次发起连接？
			if !n.Connected && n.CurrTime > n.TimeConnected+5 {
				n.connect()
			}
			if n.Connected == false {
				continue
			}
			//超时发送心跳包？每10秒发送一个？心跳包是空的？
			if n.Connected && n.CurrTime > n.TimeOutTime {
				n.TimeOutTime = n.CurrTime + 5
				n.Enqueue(comm.NewKeepAliveMsg(time.Now().UnixNano()/ 1e6).GetMsg())
			}
		}
	}
}

//把数据放入待发送队列
func (n *NetWork) Enqueue(msg ziface.IMessage) {
	if n.Connected {
		n.sendMsgs <- msg
	}
}

//发送数据异常
func (n *NetWork) SendData(msg ziface.IMessage) error {
	if !n.Connected {
		return errors.New("Conn is colse")
	}
	//有数据要发送
	_d, _ := n.dataPack.Pack(msg)
	if _, err := n.Conn.Write(_d); err != nil {
		zlog.Error("Send Data error:, ", err, " Conn Writer exit")
		return errors.New("Conn is colse")
	}
	return nil
}

//go线程调用
func (n *NetWork) gosendData() {
	for {
		select {
		case data, ok := <-n.sendMsgs:
			if ok {
				//有数据要发送
				_d, _ := n.dataPack.Pack(data)
				if _, err := n.Conn.Write(_d); err != nil {
					zlog.Error("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				zlog.Info("msgBuffChan is Closed")
				break
			}
		case <-n.ExitBuffChan:
			return
		}
	}
}

//接受到数据,这个方法不对外,go方法，循环读取数据
func (n *NetWork) receiveData() {
	defer n.Disconnect()
	for {
		if n.Connected == false {
			return
		}
		//活动时间
		n.ActivityTime = n.CurrTime
		//发封包message消息
		dp := znet.NewDataPack()

		//先读出流中的head部分
		headData := make([]byte, dp.GetHeadLen())
		_, err := io.ReadFull(n.Conn, headData) //ReadFull 会把msg填充满为止
		if err != nil {
			zlog.Error("read head error")
			break
		}
		//将headData字节流 拆包到msg中
		msgHead, err := dp.Unpack(headData)
		if err != nil {
			zlog.Error("server unpack err:", err)
			return
		}

		if msgHead.GetDataLen() > 0 {
			//msg 是有data数据的，需要再次读取data数据
			msg := msgHead.(*znet.Message)
			msg.Data = make([]byte, msg.GetDataLen())

			//根据dataLen从io中读取字节流
			_, err := io.ReadFull(n.Conn, msg.Data)
			if err != nil {
				zlog.Error("server unpack data err:", err)
				return
			}
			//不放管道了，接受到所有数据都直接处理,单独开启线程处理
			//n.Receive<-msg
			n.doReceiveData(msg)
		}
	}
}

//是否活动中（超过10秒没有活动，则认为没有活动了）
func (n *NetWork) IsActivity() bool {
	if !n.Connected {
		return false
	}

	if n.ActivityTime+10 <= n.CurrTime {
		return false
	}

	return true
}

func (n *NetWork) AddCallBack(msgid uint32, cb func(ziface.IMessage)) {
	n.receiveCallBack[msgid] = cb
}
func (n *NetWork) RemoveCallBack(msgid uint32, cb func(ziface.IMessage)) {
	delete(n.receiveCallBack, msgid)
}

func (n *NetWork) doKeepalive(msg ziface.IMessage) {
	//zlog.Debug("Receive keepalive back")
	//计算ping值
	ret := comm.NewKeepAliveMsgByByte(msg.GetData())
	n.PingTime=(time.Now().UnixNano()/1e6)-ret.CTime
	if n.PingTime>n.MaxPingTime{
		n.MaxPingTime=n.PingTime
	}
}

func (n *NetWork) doResponse(msg ziface.IMessage) {
	ret := comm.NewResponseMsgByByte(msg.GetData())
	if ret != nil {
		zlog.Debug("request back secid:", ret.SecId)
		n.requestChan[ret.SecId%10] <- ret
	}
}

func (n *NetWork) doLoginRet(msg ziface.IMessage) {
	lret := comm.NewLoginSuccessByByte(msg.GetData())
	if lret != nil {
		if lret.Result == 0 {
			n.Login = true
			zlog.Debug("Login succ")
		} else {
			zlog.Debug("Login error")
			n.Login = false
		}
	}
}

//所有数据在这里处理
func (n *NetWork) doReceiveData(msg ziface.IMessage) {
	handler, ok := n.receiveCallBack[msg.GetMsgId()]
	if !ok {
		zlog.Debug("Receive data msgId = ", msg.GetMsgId(), " is not handler func !")
		return
	} else {
		handler(msg)
	}
}

//发送请求，获得返回，柱塞
//10秒超时
func (n *NetWork) Request(msg ziface.IMessage) ([]byte, error) {
	n.secIdMutex.Lock()
	if n.secId >= math.MaxUint32 {
		n.secId = 1
	}
	n.secId++
	_secId := n.secId
	n.secIdMutex.Unlock()

	//10最多10个线程进入此
	n.requestChanLock <- true
	defer func() { <-n.requestChanLock }()

	//发送
	n.Enqueue(comm.NewRequestMsgMsg(_secId, msg.GetMsgId(), msg.GetData()).GetMsg())

	//10秒超时
	for {
		select {
		case data, ok := <-n.requestChan[_secId%10]:
			if ok {
				if data.SecId != _secId {
					break
				}
				return data.Data, nil
			}
		case <-time.After(time.Duration(10) * time.Second):
			return nil, errors.New("request time out")
		}
	}
	return nil, errors.New("request time out")
}
