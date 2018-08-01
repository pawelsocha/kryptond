package mikrotik

import (
	"github.com/pawelsocha/routeros"
)

//Client
type Client interface {
	Close()
	RunArgs(sentence []string) (*routeros.Reply, error)
}

type ClientMock struct {
	Sentece []string
}

func (c *ClientMock) Close() {}
func (c *ClientMock) RunArgs(sentence []string) (*routeros.Reply, error) {
	c.Sentece = sentence
	return new(routeros.Reply), nil
}
