package client

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/pawelsocha/kryptond/database"
)

type Session struct {
	CustomerId int64  `gorm:"column:customerid"`
	NodeId     int64  `gorm:"column:nodeid"`
	IP         uint32 `gorm:"column:ipaddr"`
	Mac        string `gorm:"column:mac"`
	Start      uint64 `gorm:"column:start"`
	Stop       uint64 `gorm:"column:stop"`
	Download   uint64 `gorm:"column:download"`
	Upload     uint64 `gorm:"column:upload"`
}

func (s Session) TableName() string {
	return "nodesessions"
}

func (s *Session) SetIP(ip string) {
	s.IP = binary.BigEndian.Uint32(net.ParseIP(ip))
}

func (s Session) Save() error {
	if database.Connection.NewRecord(s) {
		return nil
	}
	return fmt.Errorf("Can't create new session record")
}
