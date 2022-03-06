package mysql

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	db *gorm.DB
)

func OpenDB() {
	fmt.Println("mysqldb->open db")
	db1, err := gorm.Open("mysql", "root:123456@tcp(localhost:3306)/game?parseTime=true")
	if err != nil {
		panic("connect db error")
	}
	db = db1
}

func DB() *gorm.DB {
	return db
}
