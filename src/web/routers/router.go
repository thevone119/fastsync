package routers

import (
	"github.com/astaxie/beego"
	"web/controllers"
)

func init() {
	// 默认登录
	beego.Router("/", &controllers.LoginController{}, "*:LoginIn")
	beego.Router("/login", &controllers.LoginController{}, "*:LoginIn")
	beego.Router("/login_out", &controllers.LoginController{}, "*:LoginOut")
	beego.Router("/no_auth", &controllers.LoginController{}, "*:NoAuth")

	beego.Router("/home", &controllers.HomeController{}, "*:Index")
	beego.Router("/home/start", &controllers.HomeController{}, "*:Start")
	beego.AutoRouter(&controllers.ServerController{})
	beego.AutoRouter(&controllers.LogController{})
	//beego.AutoRouter(&controllers.ApiSourceController{})
	//beego.AutoRouter(&controllers.ApiPublicController{})
	//beego.AutoRouter(&controllers.TemplateController{})
	//beego.AutoRouter(&controllers.ApiDocController{})


}
