package client

import (
	"context"
	"time"

	"github.com/pawelsocha/kryptond/config"
	"github.com/pawelsocha/kryptond/database"
	. "github.com/pawelsocha/kryptond/logging"
)

type Event struct {
	ID         int64     `gorm:"column:id"`
	CustomerId int64     `gorm:"column:customerid"`
	NodeId     int64     `gorm:"column:nodeid"`
	Status     string    `gorm:"column:status"`
	Start      time.Time `gorm:"column:start"`
	Finish     time.Time `gorm:"column:finish"`
	Operation  string    `gorm:"column:type"`
	Table      string    `gorm:"column:source"`
}

func (e Event) TableName() string {
	return "kryptond_events"
}

func (e *Event) Save(cfg *config.Config) error {
	//TODO: better db handling
	db, err := database.Database(cfg)
	if err != nil {
		return err
	}

	defer db.Disconnect()
	return db.Connection.Save(e).Error
}

type Events []Event

func Subscribe(ctx context.Context, cfg *config.Config) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 60):
				e, err := CheckEvents(cfg)
				Log.Infof("event: %v, err: %s", e, err)
				return
			}
		}
	}()
}

func CheckEvents(cfg *config.Config) (Events, error) {
	db, err := database.Database(cfg)
	if err != nil {
		return nil, err
	}

	defer db.Disconnect()
	var events Events
	if err = db.Connection.Where("status = ?", "NEW").Find(&events).Error; err != nil {
		return nil, err
	}

	err = db.Connection.Table("kryptond_events").Where("status = ?", "NEW").Updates(map[string]interface{}{"status": "RUNNING"}).Error
	if err != nil {
		return nil, err
	}

	return events, nil
}
