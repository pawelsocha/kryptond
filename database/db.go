package database

import (
	"github.com/jinzhu/gorm"
	//mysql
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Connection - database client
var Connection *gorm.DB
