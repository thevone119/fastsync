package web

import (
	"github.com/astaxie/beego"
	"github.com/patrickmn/go-cache"
	"time"
	"web/models"
	_ "web/routers"
	"web/utils"
)

type WebApp struct {
	httpport int

}


func NewWebApp(p int) *WebApp{
	return &WebApp{
		httpport:p,
	}
}

//开始
func (w *WebApp) Start() {
	//注册路由
	models.Init()
	utils.Che = cache.New(60*time.Minute, 120*time.Minute)
	beego.Run()

}