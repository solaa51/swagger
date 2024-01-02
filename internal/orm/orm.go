package orm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	mysql2 "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"net"
	"os"
	"strconv"
	"strings"
	"swagger/internal/appPath"
	"swagger/internal/cFunc"
	"swagger/internal/log/bufWriter"
	"swagger/internal/watchConfig"
	"sync"
	"time"
)

//自动管理 数据库的连接与更新 配置文件更新时，实例也会同步更新

func GetDb(dbUName string) (*gorm.DB, error) {
	if _, ok := dbInstances[dbUName]; ok {
		return dbInstances[dbUName].dbIns, nil
	}

	return nil, errors.New("没找到对应数据库示例")
}

// ShowSql 为数据库连接实例开启sql日志
func ShowSql(db *gorm.DB) {
	db.Logger = logger.Default.LogMode(logger.Info)
}

func levelStrToGorm(levelStr string) logger.LogLevel {
	var lv logger.LogLevel
	switch levelStr {
	case "silent":
		lv = logger.Silent
	case "error":
		lv = logger.Error
	case "warn":
		lv = logger.Warn
	case "info":
		lv = logger.Info
	default:
		lv = logger.Warn
	}

	return lv
}

func setLogLevel(db *gorm.DB, levelStr string) {
	db.Logger = logger.Default.LogMode(levelStrToGorm(levelStr))
}

// 数据库配置文件
var dbConfigFile string

// 使用中的配置信息 用于比较是否发生变更
var dbConfigJson string

func init() {
	configDir := appPath.ConfigDir()

	dbConfigFile = configDir + "database.yaml"
	if _, err := os.Stat(dbConfigFile); err != nil {
		bufWriter.Fatal("没找到数据库配置文件", dbConfigFile, err.Error())
		return
	}

	//初始化 数据库连接池
	dbInstances = make(map[string]*dbInstance, 0)

	//实例化数据库连接
	connectDb()

	//开启数据库配置文件监控
	go func() {
		dbConfigNotifyChan, err := watchConfig.AddWatch(dbConfigFile)
		if err != nil {
			bufWriter.Fatal("监听数据库配置文件失败:", err.Error())
		}
		for {
			select {
			case <-dbConfigNotifyChan:
				bufWriter.Info("文件变更通知：", dbConfigFile)
				connectDb()
			}
		}
	}()
}

// 连接数据库
func link(conf DbConf) (*gorm.DB, error) {
	var dialer gorm.Dialector
	var logPrefix string
	var slowTime int
	if conf.SlowTime > 0 {
		slowTime = 200
	}

	switch conf.DBType {
	case "clickhouse":
		logPrefix = "clickhouse-"

		dsn := "clickhouse://" + conf.User + ":" + conf.Pass + "@" + conf.Host + ":" + conf.Port + "/" + conf.Name + "?dial_timeout=200ms&max_execution_time=60"
		dialer = clickhouse.Open(dsn)
	case "mysql", "":
		logPrefix = "mysql-"
		protocName := "tcp"
		if conf.TunnelSSHPort != "" {
			sshClient, err := getSshTunnel(conf)
			if err != nil {
				return nil, err
			}

			sshTunnelName := conf.TunnelSSHNetName
			if sshTunnelName == "" {
				sshTunnelName = conf.Name + "_ssh_tunnel"
			}

			// 注册ssh代理
			mysql2.RegisterDialContext(sshTunnelName, (&ViaSshDialer{client: sshClient}).Dial)

			protocName = sshTunnelName
		}

		dsn := conf.User + ":" + conf.Pass + "@" + protocName + "(" + conf.Host + ":" + conf.Port + ")/" + conf.Name + "?charset=utf8"
		dialer = mysql.Open(dsn)
	default:
		return nil, errors.New("数据库类型不支持:" + conf.DBType)
	}

	db, err := gorm.Open(dialer, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, //使用单表名
		},
		Logger: logger.New(newDbLogWriter(logPrefix), logger.Config{
			SlowThreshold: time.Duration(slowTime) * time.Millisecond,
			Colorful:      false,
			LogLevel:      levelStrToGorm(conf.LogLevel),
		}),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, _ := db.DB()
	sqlDB.SetConnMaxLifetime(time.Minute * 10)

	return db, nil
}

// 初始化数据库连接
func connectDb() {
	configParse := &DbConfigParse{}

	f, err := os.ReadFile(dbConfigFile)
	if err != nil {
		bufWriter.Error("解析数据库配置文件出错", dbConfigFile, err)
		return
	}
	err = yaml.Unmarshal(f, configParse)
	if err != nil {
		bufWriter.Error("解析数据库配置文件出错", dbConfigFile, err)
		return
	}

	cJsonByte, _ := json.Marshal(configParse)
	tmpJson := string(cJsonByte)

	// 与当前配置比较 无变更则不处理
	if tmpJson == dbConfigJson {
		return
	}

	for _, v := range configParse.Dbs {
		if dd, ok := dbInstances[v.UName]; ok { //已存在连接
			//判断是否有变化
			if dd.dbConf.Host != v.Host || dd.dbConf.Pass != v.Pass || dd.dbConf.Port != v.Port || dd.dbConf.User != v.User || dd.dbConf.Name != v.Name {
				db, err := link(v)
				if err != nil {
					bufWriter.Error("连接数据库-"+v.Mark+"-失败：", err.Error(), "请检查配置", dbConfigFile)
					continue
				}
				dd.update(v, db)

				continue
			}

			if dd.dbConf.LogLevel != v.LogLevel {
				db, _ := GetDb(v.UName)
				setLogLevel(db, v.LogLevel)
			}
		} else {
			db, err := link(v)
			if err != nil {
				bufWriter.Error("连接数据库-"+v.Mark+"-失败：", err.Error(), "请检查配置", dbConfigFile)
				continue
			}

			dbInstances[v.UName] = &dbInstance{
				dbConf: v,
				dbIns:  db,
			}
		}
	}

	dbConfigJson = tmpJson
}

// DbConf 数据库配置格式
type DbConf struct {
	DBType    string `yaml:"dbType"`    //数据库类型 默认mysql
	UName     string `yaml:"uName"`     //唯一名称标识
	Mark      string `yaml:"mark"`      //备注
	Host      string `yaml:"host"`      //域名或ip
	User      string `yaml:"user"`      //用户
	Pass      string `yaml:"pass"`      //密码
	Port      string `yaml:"port"`      //端口
	Name      string `yaml:"name"`      //数据库名称
	LogLevel  string `yaml:"logLevel"`  //日志级别 [silent error warn info]
	SlowTime  int    `yaml:"slowTime"`  //慢日志记录时间 单位毫秒
	LogPrefix string `yaml:"logPrefix"` //日志前缀 默认前缀为 "[dbType]-"

	//ssh tunnel加密配置
	TunnelSSHHost string `yaml:"tunnelSSHHost"`
	TunnelSSHPort string `yaml:"tunnelSSHPort"`
	TunnelSSHUser string `yaml:"tunnelSSHUser"`
	//ssh 密码验证
	TunnelSSHPassword string `yaml:"tunnelSSHPassword"`
	// 秘钥验证 RSA PRIVATE KEY
	TunnelSSHKey string `yaml:"tunnelSSHKey"`
	// 秘钥生成时的密码 一般都没有
	TunnelSSHPassphrase string `yaml:"tunnelSSHPassphrase"`
	//自定义连接名名称
	TunnelSSHNetName string `yaml:"tunnelSSHNetName"`
}
type DbConfigParse struct {
	Dbs []DbConf `yaml:"dbConfig"`
}

// dbInstances 当前已连接到的数据库实例
var dbInstances map[string]*dbInstance

// dbInstance 单个数据库连接实例
type dbInstance struct {
	mux    sync.Mutex
	dbConf DbConf
	dbIns  *gorm.DB
}

// update 更新实例
func (d *dbInstance) update(conf DbConf, db *gorm.DB) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.dbConf = conf
	d.dbIns = db
}

// TableToStruct 将数据库表 转换为struct结构输出
func TableToStruct(dbUName string, tableName string) {
	var d *dbInstance
	if _, ok := dbInstances[dbUName]; ok {
		d = dbInstances[dbUName]
	}

	if d == nil {
		bufWriter.Fatal("获取数据库连接失败：", dbUName)
	}

	//表信息
	type TableInfo struct {
		TableName    string `gorm:"column:TABLE_NAME;size:64"`
		TableComment string `gorm:"column:TABLE_COMMENT;size:255"`
	}
	var tInfo []TableInfo
	d.dbIns.Raw("SELECT * FROM information_schema.tables WHERE table_schema = ? AND table_name = ?", d.dbConf.Name, tableName).Scan(&tInfo)

	if len(tInfo) != 1 {
		fmt.Println("没有找到对应的表", tableName, d.dbConf.Name)
		return
	}

	//表字段结构
	type Result struct {
		TableName     string `gorm:"column:TABLE_NAME;size:64"`
		ColumnName    string `gorm:"column:COLUMN_NAME;size:64"`      //列名
		ColumnDefault string `gorm:"column:COLUMN_DEFAULT;size:1024"` //默认值
		IsNullable    string `gorm:"column:IS_NULLABLE;size:3"`       //是否允许为空 yes no
		DataType      string `gorm:"column:DATA_TYPE;size:64"`        //数据精确类型 int tinyint smallint decimal varchar
		CharMaxLen    int64  `gorm:"column:CHARACTER_MAXIMUM_LENGTH"` //字符串允许的最大长度
		NumPre        int64  `gorm:"column:NUMERIC_PRECISION"`        //数字类型最大长度
		NumScale      int64  `gorm:"column:NUMERIC_SCALE"`            //decimal类型 后面的精度
		ColumnType    string `gorm:"column:COLUMN_TYPE;size:50"`      //列类型
		ColumnKey     string `gorm:"column:COLUMN_KEY;size:3"`        //主键 唯一等 pri uni mul
		Extra         string `gorm:"column:EXTRA;size:30"`            //自增 auto_increment
		ColumnComment string `gorm:"column:COLUMN_COMMENT;size:255"`  //备注
		ListOrder     int64  `gorm:"column:ORDINAL_POSITION"`         //位置
	}

	var result []Result
	d.dbIns.Raw("SELECT * FROM information_schema.columns WHERE table_schema = ? AND table_name = ? ORDER BY ordinal_position ASC", d.dbConf.Name, tableName).Scan(&result)

	fmt.Println("// " + cFunc.ConvertStr(tInfo[0].TableName) + " " + tInfo[0].TableComment)
	fmt.Println("type " + cFunc.ConvertStr(tInfo[0].TableName) + " struct {")

	//字段名和备注
	type columnComment struct {
		columnName string
		comment    string
	}
	columns := make([]columnComment, len(result))

	for k, v := range result {
		str := ""
		str += "    " + cFunc.ConvertStr(v.ColumnName) //字段名

		//gorm属性
		var g []string
		g = append(g, "column:"+v.ColumnName)

		comment := "\t\t//"

		switch v.DataType {
		case "int", "tinyint", "bigint", "smallint":
			str += "\tint64\t`json:\"" + v.ColumnName + "\" gorm:\""
			if v.NumPre > 0 {
				g = append(g, "size:"+strconv.FormatInt(v.NumPre, 10))
			}
		case "float", "double", "decimal": //精度问题
			str += "\tfloat64\t`json:\"" + v.ColumnName + "\" gorm:\""
			g = append(g, "type:"+v.ColumnType)

		case "varchar", "char", "text", "longtext":
			if v.IsNullable == "YES" {
				str += "\tsql.NullString\t`json:\"" + v.ColumnName + "\" gorm:\""
			} else {
				str += "\tstring\t`json:\"" + v.ColumnName + "\" gorm:\""
			}

			g = append(g, "size:"+strconv.FormatInt(v.CharMaxLen, 10))
		case "datetime", "date", "timestamp":
			if v.IsNullable == "YES" {
				str += "\tsql.NullString\t`json:\"" + v.ColumnName + "\" gorm:\""
			} else {
				str += "\tstring\t`json:\"" + v.ColumnName + "\" gorm:\""
			}

			if v.DataType == "datetime" {
				g = append(g, "size:20")
			}

			if v.DataType == "date" {
				g = append(g, "size:10")
			}

			if v.DataType == "timestamp" { //mysql的timestamp实际为19位的时间字符串
				g = append(g, "size:19")
			}

		case "json":
			str += "\r\n\t //JSON结构需自己定义 手动补充"
			str += "\t" + cFunc.ConvertStr(v.TableName) + cFunc.ConvertStr(v.ColumnName) + "\t`json:\"" + v.ColumnName + "\" gorm:\""
			g = append(g, "TYPE:json")
		default:
			bufWriter.Fatal("暂未支持的类型-快快修改工具源码：", v.DataType)
		}

		//检查default
		if v.IsNullable == "NO" && v.ColumnDefault != "" {
			defaultInfo := "default:" + v.ColumnDefault
			if v.Extra != "" && v.Extra != "DEFAULT_GENERATED" {
				defaultInfo += " " + v.Extra
			}
			g = append(g, defaultInfo)
		}

		//主键
		if v.ColumnKey == "pri" {
			g = append(g, "primaryKey")
		}

		//自增
		if v.Extra == "auto_increment" {
			g = append(g, "autoIncrement")
		}

		comment += "" + v.ColumnType + " " + v.ColumnComment

		//构建一行列数据
		gs := strings.Join(g, ";")

		fmt.Println(str + gs + "\"`" + comment)

		columns[k] = columnComment{
			columnName: v.ColumnName,
			comment:    comment,
		}
	}

	fmt.Println("}")

	//自定义表名
	fmt.Println("func (" + cFunc.ConvertStr(tInfo[0].TableName) + ") TableName() string {")
	fmt.Println("    return \"" + tInfo[0].TableName + "\"")
	fmt.Println("}")

	//生成易操作的字段结构名
	fmt.Println("// " + cFunc.ConvertStr(tInfo[0].TableName) + "Column " + tInfo[0].TableComment + "字段名")
	fmt.Println("var " + cFunc.ConvertStr(tInfo[0].TableName) + "Column = " + "struct {")
	for _, v := range columns {
		fmt.Println("    " + cFunc.ConvertStr(v.columnName) + "    string " + v.comment)
	}
	fmt.Println("}{")
	for _, v := range columns {
		fmt.Println("    " + cFunc.ConvertStr(v.columnName) + ":    \"" + v.columnName + "\",")
	}
	fmt.Println("}")
}

/***使用ssh tunnel加密时使用***/

// ssh 隧道代理
func getSshTunnel(con DbConf) (*ssh.Client, error) {
	sshAuth := make([]ssh.AuthMethod, 0)

	if con.TunnelSSHKey != "" {
		pemKey, err := os.ReadFile(appPath.ConfigDir() + con.TunnelSSHKey)
		if err != nil {
			return nil, err
		}

		var singer ssh.Signer
		if con.TunnelSSHPassphrase != "" {
			singer, err = ssh.ParsePrivateKeyWithPassphrase(pemKey, []byte(con.TunnelSSHPassphrase))
		} else {
			singer, err = ssh.ParsePrivateKey(pemKey)
		}
		if err != nil {
			return nil, err
		}

		sshAuth = append(sshAuth, ssh.PublicKeys(singer))
	}

	if con.TunnelSSHPassword != "" {
		sshAuth = append(sshAuth, ssh.Password(con.TunnelSSHPassword))
	}

	sshConfig := &ssh.ClientConfig{
		Config:            ssh.Config{},
		User:              con.TunnelSSHUser,
		Auth:              sshAuth,
		HostKeyCallback:   ssh.InsecureIgnoreHostKey(),
		BannerCallback:    nil,
		ClientVersion:     "",
		HostKeyAlgorithms: nil,
		Timeout:           0,
	}

	sshClient, err := ssh.Dial("tcp", con.TunnelSSHHost+":"+con.TunnelSSHPort, sshConfig)
	if err != nil {
		return nil, err
	}

	return sshClient, nil
}

type ViaSshDialer struct {
	client *ssh.Client
}

func (v *ViaSshDialer) Dial(ctx context.Context, addr string) (net.Conn, error) {
	return v.client.Dial("tcp", addr)
}
