package context

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/solaa51/swagger/cFunc"
	"github.com/solaa51/swagger/log/bufWriter"
	"github.com/solaa51/swagger/snowflake"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var CtxPool *sync.Pool

func init() {
	CtxPool = &sync.Pool{
		New: func() any {
			return new(Context)
		},
	}
}

// Context 处理请求上下文
type Context struct {
	Ctx            context.Context
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	StartTime      time.Time //记录请求开始处理时间

	RequestId      int64
	StructFuncName string //最终调用的structFuncName 如果为方法则为自定义名称 如果为struct则为struct/method

	GetPost  url.Values        //get参数与 form-data或者x-www-form-urlencoded合集
	BodyData *[]byte           //body内包含的数据
	JsonData map[string]string //post + application/json 的情况下，解析后的json数据 [仅限简单结构-请看使用文档]

	ClientIp string //客户端IP

	// 处理返回信息
	CustomRet bool //自定义处理返回信息 跳过统一返回处理
	RetError  string
	RetCode   int
	RetData   any
}

func NewContext(w http.ResponseWriter, r *http.Request, structFuncName string) *Context {
	ctx := CtxPool.Get().(*Context)
	ctx.Ctx = context.Background()
	ctx.ResponseWriter = w
	ctx.Request = r
	ctx.StartTime = time.Now()
	ctx.RequestId = snowflake.IDInt64()
	ctx.StructFuncName = structFuncName
	ctx.BodyData = nil
	ctx.JsonData = make(map[string]string)
	ctx.CustomRet = false
	ctx.RetError = ""
	ctx.RetCode = 0
	ctx.RetData = nil

	//解析参数
	ctx.parseParam()
	ctx.ClientIp = ctx.clientIp()

	return ctx
}

// 批量检测参数
const (
	CheckInt int = iota
	CheckString
	CheckArrayInt
	CheckArrayString
)

// CheckField 待检测字段及规则
type CheckField struct {
	Name      string
	Desc      string
	CheckType int
	Request   bool   //是否必填
	Min       int64  //最小值或最小长度
	Max       int64  //最大值或最大长度
	Def       any    //默认值
	Reg       string //正则规则校验
}

func (c *Context) anyToInt64(a any) int64 {
	if a == nil {
		return 0
	}

	switch a.(type) {
	case int:
		return int64(a.(int))
	case int8:
		return int64(a.(int8))
	case int16:
		return int64(a.(int16))
	case int32:
		return int64(a.(int32))
	case int64:
		return a.(int64)
	case float32:
		return int64(a.(float32))
	case float64:
		return int64(a.(float64))
	case bool:
		if a.(bool) {
			return 1
		} else {
			return 0
		}
	case string:
		i, _ := strconv.ParseInt(a.(string), 10, 64)
		return i
	}

	return 0
}

func (c *Context) anyToString(a any) string {
	if a == nil {
		return ""
	}

	switch a.(type) {
	case int:
		return strconv.Itoa(a.(int))
	case int8:
		return strconv.Itoa(int(a.(int8)))
	case int16:
		return strconv.Itoa(int(a.(int16)))
	case int32:
		return strconv.Itoa(int(a.(int32)))
	case int64:
		return strconv.FormatInt(a.(int64), 10)
	case string:
		return a.(string)
	case []byte:
		return string(a.([]byte))
	case bool:
		if a.(bool) {
			return "1"
		} else {
			return "0"
		}
	}

	return ""
}

func (c *Context) CheckField(fields []*CheckField) (map[string]any, error) {
	if len(fields) == 0 {
		return nil, nil
	}

	var value any
	var err error

	ret := make(map[string]any, len(fields))
	for _, v := range fields {
		switch v.CheckType {
		case CheckInt:
			_, value, err = c.ParamDataInt(v.Name, v.Desc, v.Request, v.Min, v.Max, c.anyToInt64(v.Def))
		case CheckString:
			_, value, err = c.ParamDataString(v.Name, v.Desc, v.Request, v.Min, v.Max, c.anyToString(v.Def))
			//校验正则规则
			if v.Reg != "" && value != "" {
				b, e := regexp.MatchString(v.Reg, value.(string))
				if !b || e != nil {
					err = errors.New(v.Desc + "无法通过规则校验")
				}
			}
		case CheckArrayInt:
			_, value, err = c.ParamDataArrayInt(v.Name, v.Desc, v.Request, v.Min, v.Max)
		case CheckArrayString:
			_, value, err = c.ParamDataArrayString(v.Name, v.Desc, v.Request, v.Min, v.Max, v.Reg)
		default:
			return nil, errors.New("暂不支持的检查类型：" + v.Name)
		}

		if err != nil {
			return nil, err
		}

		ret[v.Name] = value
	}

	return ret, nil
}

func (c *Context) ParamDataArrayString(name string, desc string, require bool, min int64, max int64, reg string) (bool, []string, error) {
	exist, value, err := c.getParam(name)
	if err != nil {
		return false, []string{}, errors.New("无法获取参数")
	}

	if require && !exist {
		return false, []string{}, errors.New(desc + "不能为空")
	}

	if max == 0 { //限定下 数据库存储 普通情况 最大65535
		max = 65535
	}

	var rc *regexp.Regexp
	if reg != "" {
		rc, err = regexp.Compile(reg)
		if err != nil {
			return exist, []string{}, errors.New("正则规则错误")
		}
	}

	ret := make([]string, 0)
	is := strings.Split(value, ",")
	for _, v := range is {
		tmp := strings.TrimSpace(v)
		if tmp == "" {
			continue
		}

		num := int64(utf8.RuneCountInString(tmp))

		if min > 0 {
			if num < min {
				return exist, []string{}, errors.New(desc + "最少" + strconv.FormatInt(min, 10) + "个字")
			}
		}

		if num > max {
			return exist, []string{}, errors.New(desc + "最多" + strconv.FormatInt(max, 10) + "个字")
		}

		if rc != nil {
			b := rc.MatchString(tmp)
			if !b {
				return exist, []string{}, errors.New(desc + "无法通过规则校验" + tmp)
			}
		}

		ret = append(ret, tmp)
	}

	return true, ret, nil
}

func (c *Context) ParamDataArrayInt(name string, desc string, require bool, min int64, max int64) (bool, []int64, error) {
	exist, value, err := c.getParam(name)
	if err != nil {
		return false, []int64{}, errors.New("无法获取参数")
	}

	if require && !exist {
		return false, []int64{}, errors.New(desc + "不能为空")
	}

	ret := make([]int64, 0)
	is := strings.Split(value, ",")
	for _, v := range is {
		if v == "" {
			continue
		}

		i, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)

		if i < min {
			return exist, []int64{}, errors.New(desc + "不能小于" + strconv.FormatInt(min, 10))
		}

		if max > 0 {
			if i > max {
				return exist, []int64{}, errors.New(desc + "不能大于" + strconv.FormatInt(max, 10))
			}
		}

		ret = append(ret, i)
	}

	return true, ret, nil
}

// ParamDataInt 检查参数并返回数字
func (c *Context) ParamDataInt(name string, desc string, require bool, min int64, max int64, def int64) (bool, int64, error) {
	exist, value, err := c.getParam(name)
	if err != nil {
		return false, 0, errors.New("无法获取参数")
	}

	if require && !exist {
		return false, 0, errors.New(desc + "不能为空")
	}

	var tmp int64
	if !exist {
		tmp = def
	} else {
		tmp, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return false, 0, errors.New(desc + "格式不合法")
		}
	}

	//判断大小
	if tmp < min {
		return exist, 0, errors.New(desc + "不能小于" + strconv.FormatInt(min, 10))
	}

	if max > 0 {
		if tmp > max {
			return exist, 0, errors.New(desc + "不能大于" + strconv.FormatInt(max, 10))
		}
	}

	return exist, tmp, nil
}

// 获取参数的值
// name 参数名
// 当请求方式为json时，懒解析json数据
func (c *Context) getParam(name string) (exist bool, value string, err error) {
	// post请求并且为json格式
	if c.Request.Method == "POST" && c.Request.Header.Get("Content-Type") == "application/json" {
		if len(c.JsonData) == 0 {
			c.JsonData, err = cFunc.ParseSimpleJson(c.BodyData)
			if err != nil {
				return false, "", err
			}
		}

		value, ok := c.JsonData[name]
		if ok {
			return true, value, nil
		}
	}

	if c.GetPost[name] == nil {
		return false, "", nil
	}

	return true, strings.TrimSpace(c.GetPost[name][0]), nil
}

// ParamDataString 检查参数并返回字符串值
// name 参数名
// desc 参数说明
// require 是否必须有值
// min 最小长度
// max 最大长度 0=65535
// def 默认值
// return exist是否存在参数
func (c *Context) ParamDataString(name string, desc string, require bool, min int64, max int64, def string) (bool, string, error) {
	exist, tmp, err := c.getParam(name)
	if err != nil {
		return false, "", errors.New("无法获取参数")
	}

	if require && (!exist || tmp == "") {
		return false, "", errors.New(desc + "不能为空")
	}

	if !exist {
		tmp = def
	}

	//获得字符长度
	num := int64(utf8.RuneCountInString(tmp))

	//判断长度
	if num > 0 && min > 0 {
		if num < min {
			return exist, "", errors.New(desc + "最少" + strconv.FormatInt(min, 10) + "个字")
		}
	}

	if max == 0 { //限定下 数据库存储 普通情况 最大65535
		max = 65535
	}

	if num > max {
		return exist, "", errors.New(desc + "最多" + strconv.FormatInt(max, 10) + "个字")
	}

	return exist, tmp, nil
}

// AddRetError 向返回错误信息中追加错误内容
func (c *Context) AddRetError(err error) {
	if err == nil {
		return
	}

	if c.RetError != "" {
		c.RetError = c.RetError + "\n"
	}

	c.RetError = c.RetError + err.Error()
}

// WebSocket 升级请求为websocket
func (c *Context) WebSocket() (*websocket.Conn, error) {
	upgrade := websocket.Upgrader{}
	upgrade.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	conn, err := upgrade.Upgrade(c.ResponseWriter, c.Request, nil)
	return conn, err
}

// 解析请求get post参数
func (c *Context) parseParam() {
	// 解析key=value值
	_ = c.Request.ParseMultipartForm(32 << 20)

	// 获取body内容
	if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			bufWriter.Error("读取请求body失败", err, c.Request.URL.String())
		}
		c.BodyData = &body
		defer c.Request.Body.Close()
	}

	c.GetPost = c.Request.Form
}

// 解析请求客户端IP
func (c *Context) clientIp() string {
	return cFunc.ClientIP(c.Request)
}
