package control

// 给控制器定义一个空接口

type Control interface{}

type ControllerInstance func() Control
