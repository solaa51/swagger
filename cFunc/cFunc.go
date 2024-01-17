package cFunc

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/solaa51/swagger/appPath"
	"github.com/xuri/excelize/v2"
	"golang.org/x/net/http2"
	"golang.org/x/sync/singleflight"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"hash/crc32"
	"io"
	"math"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CheckCustomStructPtrExport 检测自定义struct指针参数并且可导出
func CheckCustomStructPtrExport(dataPtr any) error {
	t := reflect.TypeOf(dataPtr)
	// if the type is not a pointer, return an error
	if t.Kind() != reflect.Ptr {
		return errors.New("参数必须为可导出的自定义struct的指针类型")
	}
	// get the type of the element that the pointer points to
	e := t.Elem()
	// if the element type is not a struct, return an error
	if e.Kind() != reflect.Struct {
		return errors.New("参数必须为可导出的自定义struct的指针类型")
	}

	b := []byte(e.Name())[0]
	if b >= 65 && b <= 90 {
		return nil
	}

	return errors.New("参数必须为可导出的自定义struct的指针类型")
}

// Div 计算除法 保留小数位数
func Div(num1, num2, point int64) float64 {
	return 0
}

// Md5File 计算文件的md5值
func Md5File(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := md5.New()
	_, _ = io.Copy(hash, f)

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func Md5(in []byte) string {
	md5Str := md5.New()
	md5Str.Write(in)

	return hex.EncodeToString(md5Str.Sum(nil))
}

func Sha256(in []byte) string {
	m := sha256.New()
	m.Write(in)
	b := m.Sum(nil)
	return hex.EncodeToString(b)
}

// ComparePoint 比较数据是否相同 相同返回true
//
// > - 布尔型、数值类型、字符串类型、指针类型和channel是严格可比较的。
//
// > - 如果结构体类型的所有字段的类型都是严格可比较的，那么该结构体类型就是严格可比较的。
//
// > - 如果数组元素的类型是严格可比较的，那么该数组类型就是严格可比较的。
//
// > - 如果类型形参的类型集合中的所有类型都是严格可比较的，那么该类型形参就是严格可比较的。
func ComparePoint[T comparable](t1, t2 T) bool {
	return t1 == t2
}

// UTF82GBK utf8编码 转 gbk编码
func UTF82GBK(str []byte) ([]byte, error) {
	r := transform.NewReader(bytes.NewReader(str), simplifiedchinese.GBK.NewEncoder())
	b, err := io.ReadAll(r)
	return b, err
}

// GBK2UTF8 gbk编码 转 utf8编码
func GBK2UTF8(str []byte) ([]byte, error) {
	r := transform.NewReader(bytes.NewReader(str), simplifiedchinese.GBK.NewDecoder())
	b, err := io.ReadAll(r)
	return b, err
}

// Mod 给数据库分表等计算
func Mod(id int64) int64 {
	str := strconv.FormatInt(id, 10)
	shu := crc32.ChecksumIEEE([]byte(str))

	return int64(math.Mod(float64(shu), 10))
}

// FloatStringToInt float类型的字符串转为整数
// point 小数位数
func FloatStringToInt(str string, point float64) int64 {
	s, err := strconv.ParseFloat(str, 10)
	if err != nil {
		return 0
	}

	i := math.Pow(10, point)
	return int64(s * i)
}

// WriteFile 追加写入文件内容
// 如果fileName为绝对路径则直接使用 如果为相对路径则获取当前程序路径
func WriteFile(fileName string, content []byte) error {
	//判断fileName路径
	var file string
	if filepath.IsAbs(fileName) {
		file = fileName
	} else {
		dir := appPath.AppDir()
		file = dir + fileName
	}

	fileInfo, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(fileInfo)
	_, err = writer.Write(content)
	if err != nil {
		return err
	}

	_ = writer.Flush()

	return nil
}

// ClientIP 尽最大努力实现获取客户端 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

// LocalIP 判断是否为本地局域网/内网ip
func LocalIP() bool {
	ip := LocalIPV4()
	if ip == "::1" || ip == "localhost" { //本机
		return true
	} else if strings.HasPrefix(ip, "192.168.") { //内网地址
		return true
	} else if strings.HasPrefix(ip, "172.16.206.") { //内网私有IP地址 eg:172.16.206.293
		return true
	}

	return false
}

// LocalIPV4 获取本地IPv4地址
func LocalIPV4() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				if !strings.HasPrefix(ipnet.IP.String(), "169.254") { //微软保留地址
					return ipnet.IP.String()
				}
			}
		}
	}

	return ""
}

// GetFreePort 获取一个可用的端口号
func GetFreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	_ = l.Close()

	return port, nil
}

/*// CanTypes 基础类型泛型限定
type CanTypes interface {
	~int | ~int8 | string | []byte
}

func Md5T[T CanTypes](in T) string {
	m := md5.New()
	b := m.Sum([]byte(time.Now().String()))
	return hex.EncodeToString(b)
}*/

var g singleflight.Group

// SingleFlight 防止缓存击穿
//
// 适用场景: key的生成需要谨慎
//
// 1.单机模式下的生成缓存；如果分布式模式下，请使用其他方式
//
// 2.将多个请求合并成一个请求,比如post短时多次提交；get同一内容
//
// eg: ret, err := cFunc.SingleFlight(context.Background(), key, jisuan)
//
// eg: fmt.Println(ret.(int64), err)
//
// eg: func jisuan(ctx context.Context) (any, error) {
//
// eg:	return time.Now().Unix(), nil
//
// eg: }
func SingleFlight(ctx context.Context, key string, method func(context.Context) (any, error)) (any, error) {
	result := g.DoChan(key, func() (interface{}, error) {
		return method(ctx)
	})

	//防止 一个出错，全部出错
	go func() {
		time.Sleep(100 * time.Millisecond)
		g.Forget(key)
	}()

	select {
	case r := <-result:
		return r.Val, r.Err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// ConvertStr 将数据库中的表字段 转换为go中使用的名称
func ConvertStr(col string) string {
	var s string
	//s = strings.ToUpper(col[:1]) + col[1:]
	flag := 0
	for k, v := range col {
		if k == 0 {
			s += strings.ToUpper(string(v))
			flag = 0
		} else {
			if v == '_' {
				flag = 1
			} else {
				if flag == 1 {
					s += strings.ToUpper(string(v))
					flag = 0
				} else {
					s += string(v)
					flag = 0
				}
			}
		}
	}

	return s
}

// 默认为Local时区
// 主要适配PHP的时间格式
// Go 时间格式：2006-01-02 15:04:05000 -0700

var curZone *time.Location

// SetTimeZone 全局修改该包的时区
func SetTimeZone(zoneStr string) error {
	var err error
	curZone, err = time.LoadLocation(zoneStr)
	return err
}

func init() {
	//timeZoneStr = "Asia/Shanghai"
	curZone, _ = time.LoadLocation("Local")
}

// Time 返回时间戳
func Time() int64 {
	return time.Now().In(curZone).Unix()
}

var dateFormat = map[string]string{
	"Y": "2006", //年
	"y": "06",   //2位年
	"m": "01",   //带前导零的月
	"n": "1",    //不带前导零的月
	"d": "02",
	"j": "2",
	"H": "15",
	//"G": "03",
	"i": "04",
	"s": "05",
	//"O": "0700", //有问题
	//"P": "07:00", //有问题
}

// Date 返回格式化后的时间字符串
// format 支持PHP的时间格式
// unix 时间戳默认为当前时间
func Date(fmt string, unix int64) string {
	//按字母匹配格式
	newFormat := ""
	for _, v := range fmt {
		if _, ok := dateFormat[string(v)]; ok {
			newFormat += dateFormat[string(v)]
			continue
		}
		newFormat += string(v)
	}

	if unix == 0 {
		return time.Now().In(curZone).Format(newFormat)
	} else if unix == -1 {
		return time.Unix(0, 0).In(curZone).Format(newFormat)
	} else {
		return time.Unix(unix, 0).In(curZone).Format(newFormat)
	}
}

// StrToTime 将时间字符串 转换为 时间戳
// 将时间戳 转换为 指定时间格式 对应的 时间戳
// 仅支持最常用的 Y-m-d H:i:s 和 Y-m-d
// stamp 时间戳 如果为0则处理为当前时间
func StrToTime(phpFormat string, timeStr string) (int64, error) {
	format := phpFormat
	for k, v := range dateFormat {
		format = strings.ReplaceAll(format, k, v)
	}

	tt, err := time.ParseInLocation(format, timeStr, curZone)
	if err != nil {
		return 0, errors.New("cFunc:解析时间格式错误:" + format + " " + timeStr + " " + err.Error())
	}

	return tt.Unix(), nil
}

// StampToTimeStamp 将时间戳 按指定格式 转换为新的时间戳
// stamp 时间戳 如果为0则处理为当前时间
// 仅支持最常用的 Y-m-d H:i:s
// 仅支持最常用的 Y-m-d
func StampToTimeStamp(stamp int64, phpFormat string) int64 {
	var st time.Time
	if stamp == 0 {
		st = time.Now().In(curZone)
	} else {
		st = time.Unix(stamp, 0).In(curZone)
	}

	format := phpFormat
	for k, v := range dateFormat {
		format = strings.ReplaceAll(format, k, v)
	}

	str := st.Format(format)
	tt, _ := time.ParseInLocation(format, str, curZone)
	return tt.Unix()
}

// ExcelDateToDate excel读取到的日期转换为time类型
func ExcelDateToDate(excelDate string) time.Time {
	excelTime := time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
	var days, _ = strconv.Atoi(excelDate)
	return excelTime.Add(time.Second * time.Duration(days*86400))
}

// SignPost 内部使用 请求参数加密验证 发送post请求到接口
func SignPost(domain string, key string, secret string, control string, method string, data map[string]string) (string, error) {
	param, _ := json.Marshal(data)
	type Param struct {
		AppKey  string `json:"app_key"`
		Control string `json:"control"`
		Method  string `json:"method"`
		Ip      string `json:"ip"`
		Sign    string `json:"sign"`
		Param   string `json:"param"`
	}
	d := Param{
		AppKey:  key,
		Control: control,
		Ip:      LocalIPV4(),
		Method:  method,
		Param:   string(param),
	}
	d.Sign = Md5([]byte("app_key=" + d.AppKey + "&control=" + d.Control + "&ip=" + d.Ip + "&" + "method=" + d.Method + "&param=" + url.QueryEscape(string(param)) + secret))

	pJson, err := json.Marshal(d)
	if err != nil {
		return "", err
	}

	dt := map[string]string{"param": string(pJson)}
	return GetPost("POST", domain+control+"/"+method, dt, nil, nil)
}

// GetPost 发送get 或 post请求 获取数据
func GetPost(method string, sUrl string, data map[string]string, head map[string]string, cookie []*http.Cookie) (string, error) {
	//请求体数据
	var postBody *strings.Reader
	if data != nil {
		pData := url.Values{}
		for k, v := range data {
			pData.Add(k, v)
		}
		postBody = strings.NewReader(pData.Encode())
	} else {
		postBody = strings.NewReader("")
	}

	req, err := http.NewRequest(method, sUrl, postBody)
	if err != nil {
		return "", err
	}

	if _, ok := head["User-Agent"]; !ok {
		req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.120 Safari/537.36")
	}
	if _, ok := head["Content-Type"]; !ok {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	}
	if head != nil {
		for k, v := range head {
			if v != "" {
				req.Header.Add(k, v)
			}
		}
	}

	if cookie != nil {
		for _, c := range cookie {
			req.AddCookie(c)
		}
	}

	client := &http.Client{
		Timeout: time.Second * 15,
		//Transport: &http.Transport{
		//	TLSClientConfig: &tls.Config{
		//		InsecureSkipVerify: true, //跳过https验证
		//	},
		//},
		Transport: &http2.Transport{
			DialTLSContext:             nil,
			TLSClientConfig:            nil,
			ConnPool:                   nil,
			DisableCompression:         false,
			AllowHTTP:                  true,
			MaxHeaderListSize:          0,
			MaxReadFrameSize:           0,
			MaxDecoderHeaderTableSize:  0,
			MaxEncoderHeaderTableSize:  0,
			StrictMaxConcurrentStreams: false,
			ReadIdleTimeout:            0,
			PingTimeout:                0,
			WriteByteTimeout:           0,
			CountError:                 nil,
		},
	}
	response, err := client.Do(req)
	_ = req.Body.Close()
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", errors.New(response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// CreateXls 导出为excel文件 *.xlsx
// filename 导出的文件名
// titles 表头信息
// data 每行的数据信息
func CreateXls(filename string, titles []any, data [][]any) error {
	f := excelize.NewFile()
	defer f.Close()

	//流式写入器
	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return err
	}

	//行ID
	rowID := 1

	//设置第一行
	if len(titles) > 0 {
		_ = sw.SetRow("A1", titles)
		rowID = 2
	}

	if len(data) > 0 {
		for _, v := range data {
			cell, err := excelize.CoordinatesToCellName(1, rowID)
			if err != nil {
				return err
			}

			if err := sw.SetRow(cell, v); err != nil {
				return err
			}

			rowID++
		}
	}

	if err = sw.Flush(); err != nil {
		return err
	}
	if err = f.SaveAs(filename); err != nil {
		return err
	}

	return err
}

// ReadXls 读取excel文件中的数据 仅读取第一个工作表
func ReadXls(filename string) ([][]string, error) {
	//读取excel数据
	f, err := excelize.OpenFile(filename, excelize.Options{})
	if err != nil {
		return nil, err
	}

	//按索引获取工作表名 0默认为Sheet1
	name := f.GetSheetName(0)

	data := make([][]string, 0)

	//按行获取全部单元格的值
	rows, _ := f.GetRows(name)
	for _, row := range rows {
		d := make([]string, 0)
		for _, colCell := range row {
			d = append(d, colCell)
		}
		data = append(data, d)
	}

	return data, err
}

// CheckMobile 校验手机号码
func CheckMobile(mobile string) bool {
	regRuler := "^1\\d{10}$"
	reg := regexp.MustCompile(regRuler)

	return reg.MatchString(mobile)
}

// RandRangeInt 生成随机数[n - m)
func RandRangeInt(start, end int64) int64 {
	if end < start {
		return start
	}

	if end == start {
		return start
	}

	n, _ := rand.Int(rand.Reader, big.NewInt(end-start))

	return n.Int64() + start
}

// Daemon 讲进程放入后台执行
func Daemon() {
	if os.Getppid() != 1 { //判断父进程  父进程为1则表示已被系统接管
		filePath, _ := filepath.Abs(os.Args[0]) //将启动命令 转换为 绝对地址命令
		cmd := exec.Command(filePath, os.Args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		_ = cmd.Start()

		os.Exit(0)
	}
}

func CreateOrderId(id int64) string {
	stamp := Time()
	var cstSh, _ = time.LoadLocation("Asia/Shanghai")
	st := time.Unix(stamp, 0).In(cstSh)

	orderId := ""
	date := strconv.Itoa(st.Year()) + sup(int64(st.Month()), 2) + sup(int64(st.Day()), 2) + sup(int64(st.Hour()), 2) + sup(int64(st.Minute()), 2) + sup(int64(st.Second()), 2)
	r := sup(RandRangeInt(0, 1000000), 6)

	orderId = date + sup(id, 5) + r

	return orderId
}

func sup(i int64, n int) string {
	m := fmt.Sprintf("%d", i)
	for len(m) < n {
		m = fmt.Sprintf("0%s", m)
	}
	return m
}

// ParseSimpleJson 解析简单json数据为map[string]string结构
// 参考格式：{"a":"b","c":123,"d":[1,2,3],"e":["h","i","j"]}
func ParseSimpleJson(jsonStr string) (map[string]string, error) {
	if jsonStr[0] != '{' || jsonStr[len(jsonStr)-1] != '}' {
		return nil, errors.New("json格式错误")
	}

	var iis [][2]int
	keys := 0

	//特殊 :" -- "  :[ -- ]
	for k, v := range jsonStr {
		if k == 0 {
			continue
		}

		if v == '"' && jsonStr[k+1] == ':' {
			keys++
		}

		if v == '"' && jsonStr[k-1] == ':' {
			iis = append(iis, [2]int{
				k, k + strings.Index(jsonStr[k+1:], `"`) + 1,
			})
		}

		if v == '[' && jsonStr[k-1] == ':' {
			iis = append(iis, [2]int{
				k, k + strings.Index(jsonStr[k+1:], `]`) + 1,
			})
		}
	}

	var snew bytes.Buffer
	snew.Grow(len(jsonStr))
	tiStrs := make([]string, len(iis))

	if len(iis) == 0 {
		snew.WriteString(jsonStr[1 : len(jsonStr)-1])
	} else {
		for k, v := range iis {
			tiStrs[k] = jsonStr[v[0]+1 : v[1]]

			if k == 0 {
				snew.WriteString(jsonStr[1:v[0]] + "{{}}")
			}

			if k > 0 {
				snew.WriteString(jsonStr[iis[k-1][1]+1:v[0]] + "{{}}")
			}

			if k == len(iis)-1 {
				snew.WriteString(jsonStr[v[1]+1 : len(jsonStr)-1])
			}
		}
	}

	arr := make(map[string]string)

	sl := strings.Split(snew.String(), ",")
	rPIndex := 0
	for _, v := range sl {
		if v == "" {
			continue
		}
		vv := strings.Split(v, `:`)
		key := strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(vv[0]), `"`, ""))
		val := strings.TrimSpace(vv[1])
		if val == "{{}}" {
			arr[key] = strings.TrimSpace(strings.ReplaceAll(tiStrs[rPIndex], `"`, ""))
			rPIndex++
		} else {
			arr[key] = val
		}
	}

	return arr, nil
}
