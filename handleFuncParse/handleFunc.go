package handleFuncParse

import (
	"errors"
	"fmt"
	"github.com/solaa51/swagger/context"
	"reflect"
	"strconv"
	"strings"
)

// struct解析后的路由默认规则为:struct名称首字母小写或别名 / 方法名称首字母小写
// func解析后的路由规则为：自定义
var handleFuncs = make(map[string]*HandleFunc)

//func init() {
//	handleFuncs = make(map[string]*HandleFunc)
//}

// MatchAndDelHandler key完全匹配返回handler并将其从库中删除
func MatchAndDelHandler(pattern string) *HandleFunc {
	if pattern == "" {
		return nil
	}

	if handler, ok := handleFuncs[pattern]; ok {
		delete(handleFuncs, pattern)
		return handler
	}

	return nil
}

// ClearHandlerToRoute 返回剩余未处理的handler
func ClearHandlerToRoute() map[string]*HandleFunc {
	if len(handleFuncs) == 0 {
		return nil
	}

	t := make(map[string]*HandleFunc, len(handleFuncs))
	for k := range handleFuncs {
		t[k] = handleFuncs[k]
	}

	clear(handleFuncs)

	return t
}

// MatchPrefixAndDelHandler 根据前缀匹配返回handler
func MatchPrefixAndDelHandler(prefix string) map[string]*HandleFunc {
	if len(handleFuncs) == 0 {
		return nil
	}

	t := make(map[string]*HandleFunc)
	for k := range handleFuncs {
		if strings.HasPrefix(k, prefix) {
			t[k] = handleFuncs[k]
			delete(handleFuncs, k)
		}
	}

	return t
}

// LastHandler 返回未处理的handler
func LastHandler() {
	for k := range handleFuncs {
		fmt.Println(k, "=>", handleFuncs[k].StructFuncName)
	}
}

// HandleFunc 处理方法
type HandleFunc struct {
	StructFuncName string //原始struct名称和方法名称 如果为函数则名称为初始匹配名称
	methodValue    reflect.Value
	inType         []string //入参类型
}

// Call 调用函数或方法
func (h *HandleFunc) Call(ctx *context.Context, args ...string) error {
	if len(h.inType) != len(args)+1 {
		return errors.New("参数不匹配:" + h.StructFuncName)
	}

	in := make([]reflect.Value, len(h.inType))
	in[0] = reflect.ValueOf(ctx)

	for i := 1; i < len(h.inType); i++ {
		switch h.inType[i] {
		case "int":
			ki, _ := strconv.Atoi(args[i-1])
			in[i] = reflect.ValueOf(ki)
		case "int64":
			ki, _ := strconv.ParseInt(args[i-1], 10, 64)
			in[i] = reflect.ValueOf(ki)
		case "float64":
			ki, _ := strconv.ParseFloat(args[i-1], 64)
			in[i] = reflect.ValueOf(ki)
		case "string":
			in[i] = reflect.ValueOf(args[i-1])
		}
	}

	h.methodValue.Call(in)

	return nil
}

type Control interface{}
type ControllerInstance func() Control
