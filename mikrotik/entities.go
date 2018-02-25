package mikrotik

import "fmt"

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

type Secrets struct {
	ID       string `routeros:".id"`
	Name     string `routeros:"name"`
	Password string `routeros:"password"`
	Comment  string `routeros:"comment"`
	Address  string `routeros:"remote-address"`
	Service  string `routeros:"service"`
}

func (q Secrets) GetId() string {
	return q.ID
}

func (q Secrets) Where() string {
	return fmt.Sprintf("?name=%s", q.Name)
}

func (q Secrets) Path() string {
	return fmt.Sprintf("/ppp/secret")
}
