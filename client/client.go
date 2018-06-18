package client

import (
	"fmt"

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
	ID       int64  `gorm:"column:id"`
	ClientID int64  `gorm:"column:ownerid"`
	Name     string `gorm:"column:name"`
	Passwd   string `gorm:"column:passwd"`
	IP       string `gorm:"column:ipaddr"`
	Gateway  string `gorm:"column:gateway"`
	Public   string `gorm:"column:ipaddr_pub"`
	Auth     int64  `gorm:"column:authtype"`
	Access   int    `gorm:"column:access"`
	Warning  int    `gorm:"column:warning"`
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
	Rate  Ratelimit
	Nodes Nodes
}

const (
	rateLimits = `SELECT c.id, t.downceil, t.upceil
		        FROM customers c
           LEFT JOIN assignments a on (a.customerid=c.id)
           LEFT JOIN tariffs t on (t.id=a.tariffid)
			   WHERE c.id = ? and a.suspended=0`
	nodes = `SELECT nodes.id, nodes.name, INET_NTOA(nodes.ipaddr) as ipaddr,
					INET_NTOA(nodes.ipaddr_pub) as ipaddr_pub, 
					networks.gateway as gateway,
					nodes.passwd, nodes.authtype, nodes.ownerid,
					nodes.access, nodes.warning
		        FROM nodes
		   LEFT JOIN networks ON (nodes.netid=networks.id)
			   WHERE ownerid = ?`
	clientname = `SELECT lastname, name FROM customers where id = ?`
	byip       = ``
)

//NewClient create client instance with rate limits and list of nodes
func NewClient(CustomerID int64) (*Client, error) {
	var rate Ratelimit

	if database.Connection.Raw(rateLimits, CustomerID).Find(&rate).RecordNotFound() {
		rate.Downceil = 0
		rate.Upceil = 0
	}

	if rate.Downceil == 0 {
		rate.Downceil = 256
	}

	if rate.Upceil == 0 {
		rate.Downceil = 256
	}

	var clientNodes Nodes
	err := database.Connection.Raw(nodes, CustomerID).Find(&clientNodes).Error
	if err != nil {
		return nil, fmt.Errorf("Nodes for client %d not found. %s", CustomerID, err)
	}

	var name Name
	err = database.Connection.Raw(clientname, CustomerID).First(&name).Error
	return &Client{
		ID:    CustomerID,
		Name:  name.String(),
		Rate:  rate,
		Nodes: clientNodes,
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
