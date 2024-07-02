package context

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/solaa51/swagger/cFunc"
	"github.com/solaa51/swagger/library/valid"
	"github.com/solaa51/swagger/log/bufWriter"
	"github.com/solaa51/swagger/snowflake"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
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

	GetPost  url.Values   //get参数与 form-data或者x-www-form-urlencoded合集
	BodyData *[]byte      //body内包含的数据
	Valid    *valid.Valid //参数校验

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
	ctx.CustomRet = false
	ctx.RetError = ""
	ctx.RetCode = 0
	ctx.RetData = nil
	ctx.Valid = nil

	//解析参数
	ctx.parseParam()
	ctx.ClientIp = ctx.clientIp()

	return ctx
}

func (c *Context) parseBody() error {
	if c.Valid == nil {
		if c.Request.Method == "POST" && c.Request.Header.Get("Content-Type") == "application/json" {
			vid, err := valid.NewValid(c.GetPost, *c.BodyData)
			if err != nil {
				return err
			}
			c.Valid = vid
		} else {
			c.Valid, _ = valid.NewValid(c.GetPost, nil)
		}
	}

	return nil
}

func (c *Context) CheckField(regs []*valid.Regulation) (map[string]any, error) {
	if len(regs) == 0 {
		return nil, nil
	}

	if err := c.parseBody(); err != nil {
		return nil, err
	}

	return c.Valid.RegData(regs)
}

func (c *Context) ParamDataArrayString(reg *valid.Regulation) (bool, []string, error) {
	if err := c.parseBody(); err != nil {
		return false, nil, err
	}

	return c.Valid.GetStringArray(reg)
}

func (c *Context) ParamDataArrayInt(reg *valid.Regulation) (bool, []int64, error) {
	if err := c.parseBody(); err != nil {
		return false, nil, err
	}

	return c.Valid.GetIntArray(reg)
}

// ParamDataInt 检查参数并返回数字
func (c *Context) ParamDataInt(reg *valid.Regulation) (bool, int64, error) {
	if err := c.parseBody(); err != nil {
		return false, 0, err
	}

	return c.Valid.GetInt64(reg)
}

// ParamDataString 检查参数并返回字符串值
// name 参数名
// desc 参数说明
// require 是否必须有值
// min 最小长度
// max 最大长度 0=65535
// def 默认值
// return exist是否存在参数
func (c *Context) ParamDataString(reg *valid.Regulation) (bool, string, error) {
	if err := c.parseBody(); err != nil {
		return false, "", err
	}

	return c.Valid.GetString(reg)
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
