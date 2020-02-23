package comm

import (
	"sync"
	"time"
)

//零时的map,存放数据有时间限制
//模仿实现类似临时缓存的实现
//暂不实现桶分组功能
type tempMapValue struct {
	value string
	ct    int64
}

type tempMap struct {
	m0       map[string]*tempMapValue
	l        *sync.RWMutex
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
		m0:       make(map[string]*tempMapValue),
		l:        new(sync.RWMutex),
		lastTime: 0,
	}
	go TempMap.clearfor()
}

//把数据放入map
func (m *tempMap) Put(key string, v string, t int64) {
	m.l.Lock()
	defer m.l.Unlock()
	m.m0[key] = &tempMapValue{
		value: v,
		ct:    t + time.Now().Unix(),
	}
}

func (m *tempMap) Get(key string) (string, bool) {
	m.l.RLock()
	defer m.l.RUnlock()
	v, ok := m.m0[key]
	if ok {
		return v.value, ok
	} else {
		return "", ok
	}
}

//是否存在某key
func (m *tempMap) Has(key string) bool {
	m.l.RLock()
	defer m.l.RUnlock()
	_, ok := m.m0[key]
	return ok
}

func (m *tempMap) Remove(key string) {
	m.l.Lock()
	defer m.l.Unlock()
	delete(m.m0, key)
}

//清除超时的数据，线程死循环清除
func (m *tempMap) clearfor() {
	for {
		m.clear()
		time.Sleep(2 * time.Second)
	}
}

func (m *tempMap) clear() {
	m.l.Lock()
	defer m.l.Unlock()
	cct := time.Now().Unix()
	for k, v := range m.m0 {
		if v.ct < cct {
			delete(m.m0, k)
		}
	}
}

func (m *tempMap) len() int {
	return len(m.m0)
}
