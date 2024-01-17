package handle

import (
	"errors"
	"github.com/solaa51/swagger/context"
	"github.com/solaa51/swagger/control"
	"github.com/solaa51/swagger/log/bufWriter"
	"reflect"
	"strconv"
)

// 自定义struct
type handleStruct struct {
	name      string             //struct名称
	method    map[string]*method //methodName对应的数据
	selfValue reflect.Value
	selfType  reflect.Type
}

// 调用方法
func (h *handleStruct) call(ctx *context.Context, methodName string, args ...string) error {
	m, ok := h.method[methodName]
	if !ok {
		return errors.New("未知方法:" + h.name + "/" + methodName)
	}

	if len(m.in) != len(args)+1 {
		return errors.New("入参数不匹配:" + h.name + "/" + methodName)
	}

	in := make([]reflect.Value, len(m.in))
	in[0] = reflect.ValueOf(ctx)

	for i := 1; i < len(m.in); i++ {
		switch m.in[i].Name() {
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

	m.methodValue.Call(in)

	return nil
}

// 自定义struct的方法
type method struct {
	methodType  reflect.Method
	methodValue reflect.Value
	in          []reflect.Type //入参类型
	out         []reflect.Type //返回类型
}

// struct名对应的定义
var hss map[string]*handleStruct

// AddHandleStruct 新增对外的绑定关系
// 默认使用structName 首字母小写
func AddHandleStruct(strut control.ControllerInstance, aliasName string) {
	typ := reflect.TypeOf(strut())
	tyv := reflect.ValueOf(strut())
	name := typ.Elem().Name()
	if aliasName != "" {
		name = aliasName
	}
	hss[name] = &handleStruct{
		name:      typ.Elem().Name(),
		method:    parseMethod(typ.Elem().Name(), typ, tyv),
		selfValue: tyv,
		selfType:  typ,
	}
}

func parseMethod(structName string, typ reflect.Type, tyv reflect.Value) map[string]*method {
	methods := make(map[string]*method, 0)

	for i := 0; i < typ.NumMethod(); i++ {
		methodType := typ.Method(i)  //方法类型
		methodValue := tyv.Method(i) //方法值
		methodName := methodType.Name

		//存在返回值的方法 不处理
		if methodType.Type.NumOut() > 0 {
			continue
		}

		// 入参0 为方法自己 第一个参数必须为*context.Context类型才可以对外
		if methodType.Type.NumIn() >= 2 {
			argv := make([]reflect.Type, 0, methodType.Type.NumIn())
			for j := 1; j < methodType.Type.NumIn(); j++ {
				if j == 1 {
					if methodType.Type.In(j).String() == "*context.Context" {
						argv = append(argv, methodType.Type.In(j))
					}
				} else {
					typeName := methodType.Type.In(j).Name()
					switch typeName {
					case "int", "int64", "float64", "string":
					default:
						bufWriter.Warn("暂不支持的参数类型:", structName+"/"+methodName, typeName)
						continue
					}
					argv = append(argv, methodType.Type.In(j))
				}
			}

			methods[methodName] = &method{
				methodType:  methodType,
				methodValue: methodValue,
				in:          argv,
			}
		}

		/*//返回值类型 带返回值的方法不绑定
		rets := make([]reflect.Type, 0, method.Type.NumOut())
		for j := 0; j < method.Type.NumOut(); j++ {
			rets = append(rets, method.Type.Out(j))
		}*/
	}

	return methods
}

func init() {
	hss = make(map[string]*handleStruct, 0)
}
