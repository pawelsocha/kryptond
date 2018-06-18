package client

import (
	"context"
	"fmt"
	"time"

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
	Addr       uint32    `gorm:"column:addr"`
	Addr_pub   uint32    `gorm:"column:addr_pub"`
}

func (e Event) TableName() string {
	return "kryptond_events"
}

func (e *Event) AddrNtoa() string {
	var ip uint32 = e.Addr

	if e.Addr_pub > 0 {
		ip = e.Addr_pub
	}

	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24),
		byte(ip>>16),
		byte(ip>>8),
		byte(ip),
	)
}

func (e *Event) Save() error {
	return database.Connection.Save(e).Error
}

type Events []Event

func Subscribe(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 60):
				e, err := CheckEvents()
				Log.Infof("event: %v, err: %s", e, err)
				return
			}
		}
	}()
}

func CheckEvents() (Events, error) {
	err := database.Connection.Table("kryptond_events").Where("status <> ?", "DONE").Where("DATE_ADD(start, INTERVAL 12 HOUR) < NOW()").Updates(map[string]interface{}{"status": "EXPIRED"}).Error
	if err != nil {
		return nil, err
	}

	var events Events
	if err := database.Connection.Where("status IN (?)", []string{"NEW", "ERR"}).Where("customerid > 0").Find(&events).Error; err != nil {
		return nil, err
	}

	err = database.Connection.Table("kryptond_events").Where("status = ?", "NEW").Updates(map[string]interface{}{"status": "RUNNING"}).Error
	if err != nil {
		return nil, err
	}

	return events, nil
}
