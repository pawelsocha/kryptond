package database

import (
	"github.com/jinzhu/gorm"
	//mysql
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pawelsocha/kryptond/config"
)

// SQLStorage struct to keep database connection
type SQLStorage struct {
	Connection *gorm.DB
	Error      error
}

// Database open new session with database
func Database(cfg *config.Config) (*SQLStorage, error) {
	var err error
	db := new(SQLStorage)
	db.Error = nil
	db.Connection, err = gorm.Open("mysql", cfg.GetDatabaseDSN())
	if err != nil {
		db.Error = err
	}
	return db, err
}

// Disconnect close session with database
func (db *SQLStorage) Disconnect() {
	db.Connection.Close()
}
