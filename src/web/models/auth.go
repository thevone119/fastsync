/**********************************************
** @Des: 菜单
** @Author: haodaquan
** @Date:   2017-09-09 20:50:36
** @Last Modified by:   haodaquan
** @Last Modified time: 2017-09-17 21:42:08
***********************************************/
package models



type Auth struct {
	Id         int
	AuthName   string
	AuthUrl    string
	UserId     int
	Pid        int
	Sort       int
	Icon       string
	IsShow     int
	Status     int
	CreateId   int
	UpdateId   int
	CreateTime int64
	UpdateTime int64
}

//菜单列表
var AuthList []*Auth

//
func init(){
	AuthList= make([]*Auth, 0)
	AuthList = append(AuthList, &Auth{Id: 2, AuthName: "网络", AuthUrl: "", Pid:1, Sort: 0,Status:1,IsShow:1})
	AuthList = append(AuthList, &Auth{Id: 3, AuthName: "日志", AuthUrl: "", Pid:1, Sort: 0,Status:1,IsShow:1})
	AuthList = append(AuthList, &Auth{Id: 4, AuthName: "配置", AuthUrl: "", Pid:1, Sort: 0,Status:1,IsShow:1})

	//服务器
	AuthList = append(AuthList, &Auth{Id: 21, AuthName: "网络状态", AuthUrl: "/server/list", Pid:2, Sort: 0,Status:1,IsShow:1})

	AuthList = append(AuthList, &Auth{Id: 31, AuthName: "传输日志", AuthUrl: "/log/list", Pid:3, Sort: 0,Status:1,IsShow:1})

}


func (a *Auth) TableName() string {
	return TableName("uc_auth")
}

func AuthGetList(page, pageSize int, filters ...interface{}) ([]*Auth, int64) {
	return AuthList, int64(len(AuthList))
}

func AuthGetListByIds(authIds string, userId int) ([]*Auth, error) {
	return AuthList, nil
}
