package router

import (
	"fmt"
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
func MatchHandleFunc(path string) *Segment {
	s := strings.Split(path, "/")

	temp := rootSegment
	for i := range s {
		if _, ok := temp.Child[s[i]]; !ok {
			//检测同级别下是否有通配符
			if _, ok = temp.Child["*"]; ok {
				return temp.Child["*"]
			}

			continue
		}

		if temp.Router != nil {
			return temp.Child[s[i]]
		}

		temp = temp.Child[s[i]]
	}

	// 检测是否有全匹配规则
	if _, ok := rootSegment.Child["*"]; ok {
		return rootSegment.Child["*"]
	}

	return nil
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

	//TODO 可根据routers生成配置 可放到config下，监听变更，允许在线微调
	fmt.Println(routers)

	fmt.Println(rootSegment.Child)
}
