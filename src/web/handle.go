package web

import (
	"github.com/astaxie/beego"
	"net/http"
)



//所有handle均放在这里实现

//--------------------------------------------------------------首页处理----------------------------------------------
type IndexController struct {
	beego.Controller
}

func (this *IndexController) Get() {
	this.TplName = "login.html"
	//this.Ctx.WriteString("hello world")
}

//--------------------------------------------------------------登录处理----------------------------------------------

type LoginController struct {
	beego.Controller
}

func (this *LoginController) Post() {

	this.TplName = "login.html"
	//this.Ctx.WriteString("hello world")
}

//--------------------------------------------------------------主页处理----------------------------------------------

type MainController struct {
	beego.Controller
}

func (this *MainController) Get() {

	this.TplName = "main.html"
	//this.Ctx.WriteString("hello world")
}


//--------------------------------------------------------------获取目录列表，信息（最多显示1000条记录）--------------
func ListDir(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("hello world!"))
}


//--------------------------------------------------------------获取某个文件的具体上传信息json---------------------------
func GetFileSyncInfo(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("hello world!"))
}

//---------------------------------------------------------------查看同步队列------------------------------------------
func GetFileSyncQueue(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("hello world!"))
}


//---------------------------------------------------------------手工触发同步------------------------------------------
func ManualSync(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("hello world!"))
}

//---------------------------------------------------------------统计信息---------------------------------------------
