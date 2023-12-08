package context

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"swagger/internal/log/bufWriter"
	"swagger/internal/snowflake"
	"unicode/utf8"
)

// Context 处理请求上下文
type Context struct {
	Ctx            context.Context
	ResponseWriter http.ResponseWriter
	Request        *http.Request

	Path   string
	Method string

	RequestId string

	StructName string //最终调用的struct name
	MethodName string //最终调用的方法

	//Post     url.Values //单纯的form-data请求数据 或者x-www-form-urlencoded请求数据
	GetPost  url.Values //get参数与 form-data或者x-www-form-urlencoded合集
	BodyData []byte     //body内包含的数据

	ClientIp string //客户端IP

	// 处理返回信息
	CustomRet bool //自定义处理返回信息 跳过统一返回处理
	RetError  string
	RetCode   int
	RetData   any
}

func NewContext(w http.ResponseWriter, r *http.Request, structName string, finalStructName string, methodName string) *Context {
	ctx := &Context{
		Ctx:            context.Background(),
		ResponseWriter: w,
		Request:        r,
		RequestId:      snowflake.ID(),
		Method:         r.Method,
		Path:           r.URL.Path,
		StructName:     finalStructName,
		MethodName:     methodName,
	}
	//fmt.Println("remoteAddr", r.RemoteAddr)

	//解析参数
	ctx.parseParam()
	ctx.ClientIp = ctx.clientIp()

	return ctx
}

// 批量检测参数
const (
	CheckInt int = iota
	CheckString
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
			if v.Def == nil {
				v.Def = int64(0)
			}
			_, value, err = c.ParamDataInt(v.Name, v.Desc, v.Request, v.Min, v.Max, v.Def.(int64))
		case CheckString:
			if v.Def == nil {
				v.Def = ""
			}
			_, value, err = c.ParamDataString(v.Name, v.Desc, v.Request, v.Min, v.Max, v.Def.(string))
			//校验正则规则
			if v.Reg != "" {
				b, err := regexp.MatchString(v.Reg, value.(string))
				if !b || err != nil {
					err = errors.New(v.Desc + "无法通过规则校验")
				}
			}
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

// ParamDataInt 检查参数并返回数字
// name 参数名
// desc 参数说明
// require 是否必须有值
// min 最小长度
// max 最大长度
// def 默认值
// return exist是否存在参数
func (c *Context) ParamDataInt(name string, desc string, require bool, min int64, max int64, def int64) (bool, int64, error) {
	exist := true
	if c.GetPost[name] == nil {
		exist = false
	}

	if require && !exist {
		return false, 0, errors.New(desc + "不能为空")
	}

	var tmp int64
	var err error
	if !exist {
		tmp = def
	} else {
		temp := strings.TrimSpace(c.GetPost[name][0])
		/*reg := regexp.MustCompile(`^\d+$`)
		if !reg.MatchString(temp) {
			return false, 0, errors.New(desc + "格式不合法")
		}*/

		tmp, err = strconv.ParseInt(temp, 10, 64)
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

// ParamDataString 检查参数并返回字符串值
// name 参数名
// desc 参数说明
// require 是否必须有值
// min 最小长度
// max 最大长度 0=65535
// def 默认值
// return exist是否存在参数
func (c *Context) ParamDataString(name string, desc string, require bool, min int64, max int64, def string) (bool, string, error) {
	exist := true
	if c.GetPost[name] == nil {
		exist = false
	}

	if require && !exist {
		return false, "", errors.New(desc + "不能为空")
	}

	var tmp string
	if !exist {
		tmp = def
	} else {
		tmp = c.GetPost[name][0]
	}
	tmp = strings.TrimSpace(tmp)

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

/*// End 返回请求结果 当前只考虑JSON
func (c *Context) End() {
	if c.CustomRet {
		return
	}

	success := true
	if c.RetError != "" && c.RetCode == 0 {
		c.RetCode = 2000
	}

	if c.RetCode != 0 {
		success = false
	}

	if c.RetData == nil {
		c.RetData = struct{}{}
	}

	retData, _ := json.Marshal(struct {
		Success      bool   `json:"success"`
		ErrorMessage string `json:"errorMessage"`
		ErrorCode    string `json:"errorCode"`
		Data         any    `json:"data"`
		ShowType     int    `json:"showType"`
		TraceId      string `json:"traceId"`
		Host         string `json:"host"`
	}{
		Success:      success,
		ErrorMessage: c.RetError,
		ErrorCode:    strconv.Itoa(c.RetCode),
		Data:         c.RetData,
		ShowType:     0,
		TraceId:      c.RequestId,
		Host:         c.ClientIp,
	})

	c.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
	c.ResponseWriter.Write(retData)

	c.Request.Context().Done()
}*/

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
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		bufWriter.Error("读取请求body失败", err, c.Request.URL.String())
	}
	c.BodyData = body
	defer c.Request.Body.Close()

	c.GetPost = c.Request.Form
}

// 解析请求客户端IP
func (c *Context) clientIp() string {
	xForwardedFor := c.Request.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(c.Request.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}
