package server

import "zinx/znet"

//定义所有的数据，路由，包等
//ping test 自定义路由
type PingRouter struct {
	znet.BaseRouter
}