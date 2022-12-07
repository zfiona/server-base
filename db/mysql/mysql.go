package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/zfiona/server-base/log"
)

var (
	db *gorm.DB
)

type Config struct {
	Account string   //"root"
	Password string  //"123456"
	Addr string      //"127.0.0.1:3306"
	DbName string	 //"game"
}

func OpenDB(c *Config) {
	log.Debug("redis->open db")
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true&loc=Local", c.Account, c.Password, c.Addr, c.DbName)
	db, err = gorm.Open("mysql", dsn)
	if err != nil {
		panic("connect db error")
	}
}

func DB() *gorm.DB {
	return db
}

func SqlDb() *sql.DB {
	return db.DB()
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
