package server

import (
	"utils"
	"zinx/ziface"
)

//连接用户对象
type User struct {
	Uid  uint32              //用户的ID，sessionid
	Conn ziface.IConnection //当前用户的连接
	UserName string    //用户名

}

//创建一个玩家对象
func NewUser(usrName string,conn ziface.IConnection) *User {
	return &User{
		Uid:utils.GetNextUint(),
		 Conn:conn,
		 UserName:usrName,
	}
}

//用户登录
func (u *User) Login(userName string,pwd string) bool{
	if (userName=="admin" &&pwd=="admin"){
		return true
	}
	return false
}
