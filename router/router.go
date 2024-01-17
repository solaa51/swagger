package router

import (
	"goSid/internal/middleware"
	"regexp"
	"strconv"
	"strings"
)

var route *Router

type Router struct {
	compile []*regRule //路由正则规则
}

type regRule struct {
	key    string
	val    string
	regKey *regexp.Regexp // key正则规则初始化
	regVal *regexp.Regexp // val正则规则初始化

	middleware []middleware.Middleware
}

func init() {
	route = &Router{
		compile: make([]*regRule, 0),
	}
}

// AddCompileGroup 设置路由分组规则，用来适配局部中间件调用
func AddCompileGroup(name string, rules map[string]string, middle []middleware.Middleware) {
	for k, v := range rules {
		AddCompile(k, v, middle...)
	}
}

// AddCompile 设置路由规则
func AddCompile(key, value string, middle ...middleware.Middleware) {
	route.compile = append(route.compile, &regRule{
		key:        key,
		val:        value,
		regKey:     regexp.MustCompile(`^` + key + `$`),
		middleware: middle,
	})
}

// 初始化路由的val正则规则
var regVal = regexp.MustCompile(`\$\d`)

// ParseUrlPath 解析请求地址 返回cName[struct名称或别名] method[方法名] []参数
func ParseUrlPath(urlPath string) (string, string, []string, []middleware.Middleware) {
	if strings.HasSuffix(urlPath, "/") {
		urlPath = urlPath[0 : len(urlPath)-1]
	}

	if strings.HasPrefix(urlPath, "/") {
		urlPath = urlPath[1:]
	}

	//初始化返回值
	className := ""
	methodName := ""
	args := make([]string, 0)

	middle := make([]middleware.Middleware, 0)

	//可以用来处理正则匹配路由
	for _, v := range route.compile {
		matchs := v.regKey.FindStringSubmatch(urlPath)
		if len(matchs) > 0 { //匹配到了
			urlPath = v.val
			b := regVal.FindAllString(v.val, -1)
			for _, i := range b {
				ij, _ := strconv.Atoi(i[1:])
				urlPath = strings.ReplaceAll(urlPath, i, matchs[ij])
			}

			middle = append(middle, v.middleware...)
			break
		}
	}

	splitUri := strings.Split(urlPath, "/")

	switch len(splitUri) {
	case 0, 1:
		return "", "", nil, nil
	default:
		className = splitUri[0]
		methodName = strings.ToUpper(splitUri[1][:1]) + splitUri[1][1:]
		for k, v := range splitUri {
			if k > 1 {
				args = append(args, v)
			}
		}
	}

	return className, methodName, args, middle
}

/**
自定义路由匹配规则

由struct绑定生成的数据 首字母均必须为小写
cName/method [struct名称或别名/方法名] => [参数1类型，参数2类型]
*/
