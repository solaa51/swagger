package model

import (
	"github.com/solaa51/swagger/orm"
	"gorm.io/gorm"
	"log"
)

var Db *gorm.DB

var err error

func init() {
	dbName := "admin"
	Db, _, err = orm.GetDb(dbName)
	if err != nil {
		log.Fatal("无法获取数据库连接实例:", dbName)
	}
}
