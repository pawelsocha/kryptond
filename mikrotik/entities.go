package mikrotik

import "fmt"

//TODO: move to separate file

//Entity describe a record in specific path
type Entity interface {
	Path() string
	Where() string
	GetId() string
}

type Queue struct {
	ID      string `routeros:".id"`
	Name    string `routeros:"name"`
	Target  string `routeros:"target"`
	Comment string `routeros:"comment"`
	Limits  string `routeros:"max-limit"`
}

func (q Queue) GetId() string {
	return q.ID
}

func (q Queue) Where() string {
	return fmt.Sprintf("?name=%s", q.Name)
}

func (q Queue) Path() string {
	return fmt.Sprintf("/queue/simple")
}

type Secret struct {
	ID       string `routeros:".id"`
	Name     string `routeros:"name"`
	Password string `routeros:"password"`
	Comment  string `routeros:"comment"`
	Address  string `routeros:"remote-address"`
	Gateway  string `routeros:"local-address"`
	Service  string `routeros:"service"`
}

func (q Secret) GetId() string {
	return q.ID
}

func (q Secret) Where() string {
	return fmt.Sprintf("?name=%s", q.Name)
}

func (q Secret) Path() string {
	return fmt.Sprintf("/ppp/secret")
}

type Arp struct {
	Mac     string `routeros:"mac-address"`
	Address string `routeros:"address"`
}

func (a Arp) GetId() string {
	return ""
}

func (a Arp) Where() string {
	return ""
}

func (a Arp) Path() string {
	return fmt.Sprintf("/ip/arp")
}
