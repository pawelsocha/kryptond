package client

import (
	"encoding/binary"
	"net"
)

type Session struct {
	CustomerId int64  `gorm:"column:customerid"`
	NodeId     int64  `gorm:"column:nodeid"`
	IP         int64  `gorm:"column:ipaddr"`
	Mac        string `gorm:"column:mac"`
	Start      int64  `gorm:"column:start"`
	Stop       int64  `gorm:"column:stop"`
	Download   int64  `gorm:"column:download"`
	Upload     int64  `gorm:"column:upload"`
}

func (s Session) TableName() string {
	return "nodesessions"
}

func (s *Session) SetIP(ip string) {
	s.IP = binary.BigEndian.Uint64(net.ParseIP(ip))
}

func Store(s Session) error {
	db, err := database.Database(cfg)
	if err != nil {
		return nil, err
	}

	defer db.Connection.Close()

	return db.Connection.NewRecord(d)
}
