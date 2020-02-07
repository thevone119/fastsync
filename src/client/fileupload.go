package client

//文件上传类
type FileUpload struct {
	netclient *NetWork
	Fpath string

}

func NewFileUpload(nc *NetWork,fp string) *FileUpload{
	//nc.AddCallBack()
	n:=FileUpload{
		netclient:nc,
		Fpath:fp,
	}

	return &n
}

//上传文件 本地路径，远程路径
func (n *FileUpload) Upload(lp string,rp string,callback func(byte)){
	//1.开启上传文件请求通道

	//2.进行文件上传

	//3.判断文件上传是否完成，完成后关闭流，回调
}
