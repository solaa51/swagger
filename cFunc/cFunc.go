package cFunc

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/solaa51/swagger/appPath"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"hash/crc32"
	"hash/fnv"
	"io"
	"math"
	rand2 "math/rand/v2"
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
	"unsafe"
)

// ParseJsonArrayObject 解析json数组对象到 [][]byte
// 内部对象可继续交给 ParseSimpleJson 解析
func ParseJsonArrayObject(str []byte) ([][]byte, error) {
	if !json.Valid(str) || len(str) < 4 {
		return nil, errors.New("invalid json array object")
	}

	if str[0] != '[' || str[len(str)-1] != ']' {
		return nil, errors.New("invalid json array")
	}

	if str[1] != '{' || str[len(str)-2] != '}' {
		return nil, errors.New("invalid json array object")
	}

	var arr [][]byte
	var j = 0
	var pos = 0
	for i := 1; i < len(str)-1; i++ {
		if str[i] == '{' {
			if j == 0 {
				pos = i
			}
			j++
		}
		if str[i] == '}' {
			j--
			if j == 0 {
				arr = append(arr, str[pos:i+1])
			}
		}
	}
	return arr, nil
}

// ParseSimpleJson 解析简单json数据为map[string]string结构
// 参考格式：{"a":"b","c":123,"d":[1,2,3],"e":["h","i","j"]}
// 参考格式：{"a":"b,c","c":123,"d":[1,2,3],"e":["h","i","j"]}
// 参考格式：{"role_id":165, "auto_operation":"{\"abc\":21212}"}
func ParseSimpleJson(jsonByte *[]byte) (map[string]string, error) {
	jsonBtr := *jsonByte

	if !json.Valid(jsonBtr) {
		return nil, errors.New("json格式错误")
	}

	l := len(jsonBtr)
	var iis [][3]int
	i := 1

	for i < l {
		if jsonBtr[i] == '"' && jsonBtr[i+1] == ':' {
			i += 2
			continue
		}

		if jsonBtr[i] == '"' && jsonBtr[i-1] == ':' {
			ks, b := findStrEndIndex(jsonBtr[i+1:])
			if ks == -1 {
				return nil, errors.New("json格式错误")
			}
			iis = append(iis, [3]int{i, i + ks + 1, b})
			i += ks + 2
			continue
		}

		if jsonBtr[i] == '[' && jsonBtr[i-1] == ':' {
			ks := bytes.Index(jsonBtr[i+1:], []byte(`]`))
			if ks == -1 {
				return nil, errors.New("json格式错误")
			}
			iis = append(iis, [3]int{i, i + ks + 1, 0})
			i += ks + 2
			continue
		}

		i++
	}

	var snew bytes.Buffer
	snew.Grow(l)
	tiStrs := make([]string, len(iis))

	if len(iis) == 0 {
		snew.Write(jsonBtr[1 : l-1])
	} else {
		prevEnd := 1
		for k, v := range iis {
			strValue := string(jsonBtr[v[0]+1 : v[1]])
			if v[2] == 0 {
				strValue = strings.ReplaceAll(strValue, `"`, "")
			} else {
				strValue = strings.ReplaceAll(strValue, `\"`, `"`)
			}
			tiStrs[k] = strings.TrimSpace(strValue)
			snew.Write(jsonBtr[prevEnd:v[0]])
			snew.WriteString("{{}}")
			prevEnd = v[1] + 1
		}
		snew.Write(jsonBtr[prevEnd : l-1])
	}

	arr := make(map[string]string)
	sl := strings.Split(snew.String(), ",")
	rPIndex := 0
	for _, v := range sl {
		if v == "" {
			continue
		}
		vv := strings.SplitN(v, `:`, 2)
		if len(vv) != 2 {
			return nil, errors.New("json格式错误")
		}
		key := strings.TrimSpace(strings.ReplaceAll(vv[0], `"`, ""))
		val := strings.TrimSpace(vv[1])
		if val == "{{}}" {
			arr[key] = tiStrs[rPIndex]
			rPIndex++
		} else {
			arr[key] = val
		}
	}

	return arr, nil
}

// 返回位置和 是否包含转义符 0和1为int类型一致方便处理
func findStrEndIndex(sb []byte) (kk int, b int) {
	kk = -1
	for k := range sb {
		if sb[k] == '"' {
			if k > 0 {
				if sb[k-1] == '\\' {
					b = 1
					continue
				}
			}

			kk = k
			break
		}
	}

	return
}

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

func Sha256File(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha256.New()
	_, _ = io.Copy(hash, f)
	return hex.EncodeToString(hash.Sum(nil)), nil
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
	xof := r.Header.Get("X-Original-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xof, ",")[0])
	if ip != "" {
		return ip
	}

	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip = strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
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
	//timeZoneStr := "Asia/Shanghai"
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
		//Transport: &http2.Transport{
		//	DialTLSContext: nil,
		//	TLSClientConfig: &tls.Config{
		//		InsecureSkipVerify: true, //跳过https验证
		//	},
		//	ConnPool:                   nil,
		//	DisableCompression:         false,
		//	AllowHTTP:                  true,
		//	MaxHeaderListSize:          0,
		//	MaxReadFrameSize:           0,
		//	MaxDecoderHeaderTableSize:  0,
		//	MaxEncoderHeaderTableSize:  0,
		//	StrictMaxConcurrentStreams: false,
		//	ReadIdleTimeout:            0,
		//	PingTimeout:                0,
		//	WriteByteTimeout:           0,
		//	CountError:                 nil,
		//},
	}
	response, err := client.Do(req)
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

// CheckMobile 校验手机号码
func CheckMobile(mobile string) bool {
	regRuler := "^1\\d{10}$"
	reg := regexp.MustCompile(regRuler)

	return reg.MatchString(mobile)
}

// RandRangeInt 生成随机数[n - m)
func RandRangeInt(start, end uint64) int64 {
	if end < start {
		return int64(start)
	}

	if end == start {
		return int64(start)
	}

	n := RandInt(end - start)

	return int64(n + start)
}

// RandInt 生成随机数[0 - max)
func RandInt(max uint64) uint64 {
	if max == 0 {
		return 0
	}
	return rand2.Uint64N(max)
}

// Daemon 将进程放入后台执行
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

func Bytes2String(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func String2Bytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BreakSensitive 对铭感字符串脱敏处理
func BreakSensitive(str string, frontLen int, behindLen int) string {
	if len(str) > frontLen+behindLen {
		return str[0:frontLen] + "****" + str[len(str)-behindLen:]
	}

	return str
}

// Fnv32a 计算字符串的hash值，将字符串转成整形
func Fnv32a(str string) uint32 {
	h := fnv.New32()
	_, _ = h.Write(String2Bytes(str))
	return h.Sum32()
}

// HashShard 求字符串的hash值分片数
// @param shardCount 分配总数
func HashShard(str string, shardCount int) int {
	return int(Fnv32a(str)) % shardCount
}
