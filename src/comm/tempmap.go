package comm

import (
	"sync"
	"time"
)

//零时的map,存放数据有时间限制
type tempMapValue struct {
	value string
	ct int64
}



type tempMap struct {
	m0 map[string]tempMapValue
	l *sync.RWMutex
	lastTime int64
}

/*
	定义一个全局的对象
*/
var TempMap *tempMap

/*
	提供init方法，默认加载
*/
func init() {
	//初始化GlobalObject变量，设置一些默认值
	TempMap = &tempMap{
		m0:            make(map[string]tempMapValue),
		l:new(sync.RWMutex),
		lastTime:0,
	}
	go TempMap.clearfor()
}

//把数据放入map
func (m *tempMap) Put(key string,v string,t int64) {
	tt := t+time.Now().Unix()
	m.m0[key] = tempMapValue{
		value:v,
		ct:tt,
	}
}

func (m *tempMap) Get(key string) (string,bool){
	v, ok := m.m0[key]
	return v.value,ok
}


func (m *tempMap) Remove(key string) {
	m.l.RLock()
	defer m.l.RUnlock()
	delete(m.m0,key)
}


//清除超时的数据，线程死循环清除
func (m *tempMap) clearfor(){
	for{
		m.clear()
		time.Sleep(10 * time.Second)
	}
}

func (m *tempMap) clear(){
	m.l.RLock()
	defer m.l.RUnlock()
	currtime := time.Now().Unix()
	for k, v := range m.m0{
		if v.ct<currtime{
			delete(m.m0,k)
		}
	}
}