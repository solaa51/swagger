package router

import (
	"strings"
)

// 路由匹配链表

type Segment struct {
	Name   string //路由单节字符串
	Router *Router
	Parent *Segment
	Child  map[string]*Segment
}

// 生成路由匹配规则
func addSegment(path string, router *Router) {
	s := strings.Split(path, "/")

	parent := rootSegment
	for i := range s {
		parent.Child[s[i]] = &Segment{
			Name:   s[i],
			Router: router,
			Parent: parent,
			Child:  make(map[string]*Segment),
		}

		parent = parent.Child[s[i]]
	}
}

// MatchHandleFunc 匹配路由查找对应方法
func MatchHandleFunc(urlPath string) (*Segment, []string) {
	if strings.HasSuffix(urlPath, "/") {
		urlPath = urlPath[0 : len(urlPath)-1]
	}

	if strings.HasPrefix(urlPath, "/") {
		urlPath = urlPath[1:]
	}

	s := strings.Split(urlPath, "/")

	var sig *Segment
	temp := rootSegment
	j := 0
	for i := range s { //将urlPath一层层比对
		j = i
		if _, ok := temp.Child[s[i]]; !ok {
			//检测同级别下是否有通配符
			if _, ok = temp.Child["*"]; ok {
				return temp.Child["*"], []string{}
			}

			break
		}

		if len(temp.Child[s[i]].Child) == 0 {
			sig = temp.Child[s[i]]
			break
		}

		temp = temp.Child[s[i]]
	}

	if sig != nil {
		return sig, s[j+1:]
	}

	// 检测是否有全匹配规则
	if _, ok := rootSegment.Child["*"]; ok {
		return rootSegment.Child["*"], []string{}
	}

	return nil, nil
}

// 根节点
var rootSegment = &Segment{
	Name:   "/",
	Router: nil,
	Parent: nil,
	Child:  make(map[string]*Segment),
}

// InitRouterSegment 初始化链表路由规则
func InitRouterSegment() {
	initLastHandlerFunc()

	for k := range routers {
		addSegment(k, routers[k])
	}
}
