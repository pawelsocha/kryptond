package mikrotik

import (
	routeros "github.com/pawelsocha/routeros"
)

type Task struct {
	Command string
	Result  chan *routeros.Reply
}
