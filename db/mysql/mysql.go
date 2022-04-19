package mysql

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
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
	fmt.Println("mysqldb->open db")
	var err error
	db, err = gorm.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", c.Account, c.Password, c.Addr, c.DbName))
	if err != nil {
		panic("connect db error")
	}
	db.DB().SetConnMaxLifetime(120 * time.Second)
}

func DB() *gorm.DB {
	return db
}
