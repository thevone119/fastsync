package znet

import (
	"strconv"
	"zinx/utils"
	"zinx/ziface"
	"zinx/zlog"
)

type MsgHandle struct {
	Apis           map[uint32]ziface.IRouter //存放每个MsgId 所对应的处理方法的map属性
	WorkerPoolSize uint32                    //业务工作Worker池的数量
	TaskQueue      []chan ziface.IRequest    //Worker负责取任务的消息队列
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		//一个worker对应一个queue
		TaskQueue: make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
}

//将消息交给TaskQueue,由worker进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {
	//根据ConnID来分配当前的连接应该由哪个worker负责处理
	//轮询的平均分配法则

	//用客户端ID进行分配
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	//得到需要处理此条连接的workerID,这里改下，通过消息ID来分，不同消息用不同的线程处理
	//workerID := request.GetMsgID() % mh.WorkerPoolSize
	//fmt.Println("Add ConnID=", request.GetConnection().GetConnID()," request msgID=", request.GetMsgID(), "to workerID=", workerID)
	//将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- request
}

//马上以非阻塞方式处理消息
func (mh *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		zlog.Error("api msgId = ", request.GetMsgID(), " is not FOUND!")
		return
	}

	//错误拦截必须配合defer使用  通过匿名函数使用
	//这里是全局所有的业务逻辑调用方法，如果这里出错了。则直接中断客户端连接,避免程序死掉
	defer func() {
		//恢复程序的控制权
		err := recover()
		if err != nil {
			zlog.Error("程序出现意外错误，关闭客户端连接", err)
			request.GetConnection().Stop()
		}
	}()

	//执行对应处理方法
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

//为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgId uint32, router ziface.IRouter) {
	//1 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[msgId]; ok {
		panic("repeated api , msgId = " + strconv.Itoa(int(msgId)))
	}
	//2 添加msg与api的绑定关系
	mh.Apis[msgId] = router
	zlog.Debug("Add api msgId = ", msgId)
}

//启动一个Worker工作流程
func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	zlog.Debug("Worker ID = ", workerID, " is started.")
	//不断的等待队列中的消息
	for {
		select {
		//有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

//启动worker工作池
func (mh *MsgHandle) StartWorkerPool() {
	//遍历需要启动worker的数量，依此启动
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		//一个worker被启动
		//给当前worker对应的任务队列开辟空间
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		//启动当前Worker，阻塞的等待对应的任务队列是否有消息传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}
