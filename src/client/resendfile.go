package client

type ReSendFileHandle struct {

}


//1个小时内重发3次，超过3次，则不重发了
//文件重发,第一次1分钟后，第2次5分钟后，第3次30分钟后
type ReSendFile struct {
	LPath        string             //本机路径（绝对路径）
	RetCodes 	[]byte				//各个服务器端上传返回的码，默认初始化为255
	ReSendCount  int16				//重发次数
	NextSendTime		int64			//下次重发时间
}

//重发文件
func (n *ReSendFile) NewSendFile(l *ReSendFile) {

}