package client

import (
	"fmt"

	"github.com/pawelsocha/kryptond/config"
	"github.com/pawelsocha/kryptond/database"
)

// Ratelimit Information about download and upload rate limits for
// specifig customer
type Ratelimit struct {
	Downceil int64 `gorm:"column:downceil"`
	Upceil   int64 `gorm:"column:upceil"`
}

//Node computer connected to customer
type Node struct {
	ClientID int64  `gorm:"column:ownerid"`
	Name     string `gorm:"column:name"`
	Passwd   string `gorm:"column:passwd"`
	IP       int64  `gorm:"column:ipaddr"`
	Public   int64  `gorm:"column:ipaddr_pub"`
	Auth     int64  `gorm:"column:authtype"`
}

//Name customer name
type Name struct {
	Name     string `gorm:"column:name"`
	Lastname string `gorm:"column:lastname"`
}

func (n Name) String() string {
	return fmt.Sprintf("%s %s", n.Name, n.Lastname)
}

//Nodes List of nodes
type Nodes []Node

// Client description
type Client struct {
	ID    int64
	Name  string
	Rate  *Ratelimit
	Nodes *Nodes
}

const (
	rateLimits = `SELECT c.id, t.downceil, t.upceil
		        FROM customers c
           LEFT JOIN assignments a on (a.customerid=c.id)
           LEFT JOIN tariffs t on (t.id=a.tariffid)
			   WHERE c.id = ?`
	nodes = `SELECT name, ipaddr, ipaddr_pub, passwd, authtype, ownerid
		        FROM nodes
			   WHERE ownerid = ?`
	clientname = `SELECT lastname, name FROM customers where id = ?`
	byip       = ``
)

//NewClient create client instance with rate limits and list of nodes
func NewClient(CustomerID int64, cfg *config.Config) (*Client, error) {
	db, err := database.Database(cfg)
	if err != nil {
		return nil, err
	}

	defer db.Disconnect()

	var rate Ratelimit
	err = db.Connection.Raw(rateLimits, CustomerID).Find(&rate).Error
	if err != nil {
		return nil, err
	}

	var clientNodes Nodes
	err = db.Connection.Raw(nodes, CustomerID).Find(&clientNodes).Error
	if err != nil {
		return nil, err
	}

	var name Name
	err = db.Connection.Raw(clientname, CustomerID).First(&name).Error
	return &Client{
		ID:    CustomerID,
		Name:  name.String(),
		Rate:  &rate,
		Nodes: &clientNodes,
	}, nil
}

//String convert client struct to string
func (c Client) String() string {
	return fmt.Sprint("Client%s", c.ID)
}

//Description returns client description for comment field in routeros
func (c Client) Description() string {
	return fmt.Sprint("Client:%s:%s", c.ID, c.Name)
}
