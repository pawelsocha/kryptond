package mikrotik

import (
	routeros "github.com/pawelsocha/routeros"
)

type Result struct {
	Reply *routeros.Reply
	Error error
}

type Task struct {
	Action string
	Entity Entity
	Result chan *Result
}
