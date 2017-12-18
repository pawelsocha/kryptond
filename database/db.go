package database

import (
	"github.com/jinzhu/gorm"
	//mysql
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pawelsocha/kryptond/config"
)

// SqlStorage struct to keep database connection
type SqlStorage struct {
	Connection *gorm.DB
	Error      error
}

// Database open new session with database
func Database(cfg *config.Config) *SqlStorage {
	var err error
	db := new(SqlStorage)
	db.Error = nil
	db.Connection, err = gorm.Open("mysql", cfg.GetDatabaseDSN())
	if err != nil {
		db.Error = err
	}
	return db
}

// Disconnect close session with database
func (db *SqlStorage) Disconnect() {
	db.Connection.Close()
}
