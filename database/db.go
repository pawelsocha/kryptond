package database

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// DbConn database connection
var DB *gorm.DB

// Connect open new session with database
func Connect(uri string) error {
	var err error
	DB, err = gorm.Open("mysql", uri)
	return err
}

// Disconnect close session with database
func Disconnect() {
	DB.Close()
}
