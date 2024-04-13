package handleFuncParse

import (
	"github.com/solaa51/swagger/context"
	"github.com/solaa51/swagger/log/bufWriter"
	"reflect"
	"strings"
)

// ParseFunc 直接注册方法
func ParseFunc(pattern string, f func(*context.Context)) {
	if _, ok := handleFuncs[pattern]; ok {
		bufWriter.Fatal("初始化方法存在路由冲突", pattern)
	}

	handleFuncs[pattern] = ParseFuncToRoute(pattern, f)
}

func ParseFuncToRoute(pattern string, f func(*context.Context)) *HandleFunc {
	return &HandleFunc{
		StructFuncName: pattern,
		methodValue:    reflect.ValueOf(f),
		inType:         []string{""},
	}
}

func ParseStructToRoute(strut ControllerInstance, aliasName string) map[string]*HandleFunc {
	return parseStruct(strut, aliasName)
}

func ParseStructsToRoute(strut ...ControllerInstance) map[string]*HandleFunc {
	ret := make(map[string]*HandleFunc)
	for k := range strut {
		ms := parseStruct(strut[k], "")
		for k1 := range ms {
			ret[k1] = ms[k1]
		}
	}

	return ret
}

func parseStruct(strut ControllerInstance, aliasName string) map[string]*HandleFunc {
	typ := reflect.TypeOf(strut())
	tyv := reflect.ValueOf(strut())
	sName := typ.Elem().Name()
	sName = strings.ToLower(sName[:1]) + sName[1:]
	if aliasName != "" {
		sName = aliasName
	}

	methods := parseMethod(typ.Elem().Name(), typ, tyv)
	fs := make(map[string]*HandleFunc, len(methods))
	for k := range methods {
		key := sName + "/" + strings.ToLower(k[:1]) + k[1:]
		fs[key] = methods[k]
	}

	return fs
}

// ParseStruct 解析struct 生成方法与路由的映射关系
// structName首字母转小写 如果aliasName不为空，则以aliasName为准
// funcName也会首字母转小写
func ParseStruct(strut ControllerInstance, aliasName string) {
	methods := parseStruct(strut, aliasName)

	for k := range methods {
		handleFuncs[k] = methods[k]
	}
}

// 解析struct下面对外的方法，将适合的方法绑定到路由
func parseMethod(structName string, typ reflect.Type, tyv reflect.Value) map[string]*HandleFunc {
	methods := make(map[string]*HandleFunc)

	for i := 0; i < typ.NumMethod(); i++ {
		methodType := typ.Method(i)  //方法类型
		methodValue := tyv.Method(i) //方法值
		methodName := methodType.Name

		if !methodType.IsExported() {
			continue
		}

		//存在返回值的方法 不处理
		if methodType.Type.NumOut() > 0 {
			continue
		}

		// 入参0 为方法自己 第一个参数必须为*context.Context类型才可以对外
		if methodType.Type.NumIn() >= 2 {
			argv := make([]string, 0, methodType.Type.NumIn())
			for j := 1; j < methodType.Type.NumIn(); j++ {
				if j == 1 {
					if methodType.Type.In(j).String() == "*context.Context" {
						argv = append(argv, methodType.Type.In(j).Name())
					}
				} else {
					typeName := methodType.Type.In(j).Name()
					switch typeName {
					case "int", "int64", "float64", "string":
					default:
						bufWriter.Warn("暂不支持的参数类型:", structName+"/"+methodName, typeName)
						continue
					}
					argv = append(argv, methodType.Type.In(j).Name())
				}
			}

			methods[methodName] = &HandleFunc{
				StructFuncName: structName + "." + methodName,
				methodValue:    methodValue,
				inType:         argv,
			}
		}
	}

	return methods
}
