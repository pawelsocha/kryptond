package mikrotik

import (
	routeros "github.com/pawelsocha/routeros"
)

type Result struct {
	Reply routeros.Reply
	Error error
}

type Task struct {
	Command string
	Result  chan Result
}
