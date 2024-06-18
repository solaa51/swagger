package valid

import (
	"errors"
	"github.com/solaa51/swagger/cFunc"
	"maps"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type validType int

// 批量检测参数
const (
	Int validType = iota
	String
	ArrayInt    //仅支持在json的内层使用
	ArrayString //仅支持在json的内层使用
)

// Regulation 校验规则
type Regulation struct {
	Name      string    //参数名称
	Desc      string    //参数描述
	CheckType validType //校验类型
	Required  bool      //是否必填
	Min       int64     //最小值或最小长度
	Max       int64     //最大值或最大长度
	Def       any       //默认值
	Reg       string    //正则规则校验
}

type Valid struct {
	MapData map[string]string //解析后的键值对
}

// NewValid 创建校验对象
// url.Values url参数 或 post 键值对 参数只解析[0] 【数组不做处理】
// jsonStr json字符串 当key相同时 json权重高于url.Values数据
func NewValid(urlValues url.Values, jsonStr []byte) (*Valid, error) {
	v := &Valid{
		MapData: make(map[string]string),
	}

	if urlValues != nil {
		data := v.parseParamHttpFields(urlValues)
		maps.Copy(v.MapData, data)
	}

	if jsonStr != nil {
		data, err := v.parseJson(jsonStr)
		if err != nil {
			return nil, err
		}
		maps.Copy(v.MapData, data)
	}

	return v, nil
}

// RegData 根据规则校验参数 碰到错误立即返回
func (v *Valid) RegData(regs []*Regulation) (map[string]any, error) {
	ret := make(map[string]any, len(regs))

	var value any
	var err error

	for _, reg := range regs {
		switch reg.CheckType {
		case Int:
			_, value, err = v.GetInt64(reg)
		case String:
			_, value, err = v.GetString(reg)
		case ArrayInt:
			_, value, err = v.GetIntArray(reg)
		case ArrayString:
			_, value, err = v.GetStringArray(reg)
		default:
			return nil, errors.New("不支持的校验类型")
		}

		if err != nil {
			return nil, err
		}

		ret[reg.Name] = value
	}

	return ret, nil
}

func (v *Valid) parseJson(data []byte) (map[string]string, error) {
	return cFunc.ParseSimpleJson(&data)
}

func (v *Valid) parseParamHttpFields(data url.Values) map[string]string {
	ret := make(map[string]string, len(data))
	for k, v := range data {
		ret[k] = v[0]
	}

	return ret
}

func (v *Valid) GetString(reg *Regulation) (bool, string, error) {
	value, exist := v.MapData[reg.Name]

	if reg.Required && !exist {
		return false, "", errors.New(reg.Desc + "不能为空")
	}

	if reg.Required && (!exist || value == "") {
		return false, "", errors.New(reg.Desc + "不能为空")
	}

	var tmp string
	if !exist {
		tmp = reg.defToString()
	}

	//获得字符长度
	num := int64(utf8.RuneCountInString(tmp))

	//判断长度
	if num > 0 && reg.Min > 0 {
		if num < reg.Min {
			return exist, "", errors.New(reg.Desc + "最少" + strconv.FormatInt(reg.Min, 10) + "个字")
		}
	}

	if reg.Max == 0 { //限定下 数据库存储 普通情况 最大65535
		reg.Max = 65535
	}

	if num > reg.Max {
		return exist, "", errors.New(reg.Desc + "最多" + strconv.FormatInt(reg.Max, 10) + "个字")
	}

	if !reg.checkRegexp(value) {
		return exist, "", errors.New(reg.Desc + "正则校验失败")
	}

	return exist, tmp, nil
}

func (v *Valid) GetInt64(reg *Regulation) (bool, int64, error) {
	value, exist := v.MapData[reg.Name]

	if reg.Required && !exist {
		return false, 0, errors.New(reg.Desc + "不能为空")
	}

	var (
		tmp int64
		err error
	)
	if !exist {
		tmp = reg.defToInt64()
	} else {
		if !reg.checkRegexp(value) {
			return false, 0, errors.New(reg.Desc + "正则校验失败")
		}

		tmp, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return false, 0, errors.New(reg.Desc + "类型错误")
		}
	}

	//判断大小
	if tmp < reg.Min {
		return exist, 0, errors.New(reg.Desc + "不能小于" + strconv.FormatInt(reg.Min, 10))
	}

	if reg.Max > 0 {
		if tmp > reg.Max {
			return exist, 0, errors.New(reg.Desc + "不能大于" + strconv.FormatInt(reg.Max, 10))
		}
	}

	return exist, tmp, nil
}

func (v *Valid) GetIntArray(reg *Regulation) (bool, []int64, error) {
	value, exist := v.MapData[reg.Name]

	if reg.Required && !exist {
		return false, []int64{}, errors.New(reg.Desc + "不能为空")
	}

	ret := make([]int64, 0)
	is := strings.Split(value, ",")
	for _, va := range is {
		if va == "" {
			continue
		}

		if !reg.checkRegexp(value) {
			return exist, []int64{}, errors.New(reg.Desc + "正则校验失败")
		}

		i, _ := strconv.ParseInt(strings.TrimSpace(va), 10, 64)

		if i < reg.Min {
			return exist, []int64{}, errors.New(reg.Desc + "不能小于" + strconv.FormatInt(reg.Min, 10))
		}

		if reg.Max > 0 {
			if i > reg.Max {
				return exist, []int64{}, errors.New(reg.Desc + "不能大于" + strconv.FormatInt(reg.Max, 10))
			}
		}

		ret = append(ret, i)
	}

	return true, ret, nil
}

func (v *Valid) GetStringArray(reg *Regulation) (bool, []string, error) {
	value, exist := v.MapData[reg.Name]

	if reg.Required && (!exist || value == "") {
		return false, []string{}, errors.New(reg.Desc + "不能为空")
	}

	if reg.Max == 0 { //数据库存储时 一般最大65535
		reg.Max = 65535
	}

	var (
		rc  *regexp.Regexp
		err error
	)
	if reg.Reg != "" {
		rc, err = regexp.Compile(reg.Reg)
		if err != nil {
			return exist, []string{}, errors.New("正则初始化错误:" + err.Error())
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

		if reg.Min > 0 {
			if num < reg.Min {
				return exist, []string{}, errors.New(reg.Desc + "最少" + strconv.FormatInt(reg.Min, 10) + "个字")
			}
		}

		if num > reg.Max {
			return exist, []string{}, errors.New(reg.Desc + "最多" + strconv.FormatInt(reg.Max, 10) + "个字")
		}

		if rc != nil {
			b := rc.MatchString(tmp)
			if !b {
				return exist, []string{}, errors.New(reg.Desc + "无法通过规则校验" + tmp)
			}
		}

		ret = append(ret, tmp)
	}

	return true, ret, nil
}

func (r *Regulation) defToString() string {
	if r.Def == nil {
		return ""
	}

	switch r.Def.(type) {
	case int:
		return strconv.Itoa(r.Def.(int))
	case int8:
		return strconv.Itoa(int(r.Def.(int8)))
	case int16:
		return strconv.Itoa(int(r.Def.(int16)))
	case int32:
		return strconv.Itoa(int(r.Def.(int32)))
	case int64:
		return strconv.FormatInt(r.Def.(int64), 10)
	case string:
		return r.Def.(string)
	case []byte:
		return string(r.Def.([]byte))
	case bool:
		if r.Def.(bool) {
			return "1"
		} else {
			return "0"
		}
	}

	return ""
}

func (r *Regulation) defToInt64() int64 {
	if r.Def == nil {
		return 0
	}

	switch r.Def.(type) {
	case int:
		return int64(r.Def.(int))
	case int8:
		return int64(r.Def.(int8))
	case int16:
		return int64(r.Def.(int16))
	case int32:
		return int64(r.Def.(int32))
	case int64:
		return r.Def.(int64)
	case float32:
		return int64(r.Def.(float32))
	case float64:
		return int64(r.Def.(float64))
	case bool:
		if r.Def.(bool) {
			return 1
		} else {
			return 0
		}
	case string:
		i, _ := strconv.ParseInt(r.Def.(string), 10, 64)
		return i
	}

	return 0
}

func (r *Regulation) checkRegexp(value string) bool {
	if r.Reg != "" && value != "" {
		b, e := regexp.MatchString(r.Reg, value)
		if !b || e != nil {
			return false
		}
	}

	return true
}
