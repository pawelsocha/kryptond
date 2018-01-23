package mikrotik

import (
	. "github.com/pawelsocha/kryptond/logging"
	routeros "github.com/pawelsocha/routeros"
)

type Workers struct {
	nodes map[string]*Device
}

func NewWorkers() *Workers {
	return &Workers{
		nodes: make(map[string]*Device),
	}
}

func (w *Workers) AddNewDevice(host string) (*Device, error) {
	device, err := NewDevice(host)
	if err != nil {
		Log.Errorf("Can't create routeros device for host %s. Error: %s", host, err)
		return nil, err
	}
	device.Run()
	w.nodes[host] = device
	return device, nil
}

func (w *Workers) ExecuteCommand(cmd string, result chan *routeros.Reply) {
	for host, device := range w.nodes {
		task := Task{
			Command: cmd,
			Result:  result,
		}
		j := device.Task()
		Log.Debugf("sending job %v to %s", task.Command, host)

		j <- task
	}
}

func (w *Workers) GetDevice(host string) *Device {
	if i, ok := w.nodes[host]; ok {
		return i
	}
	return nil
}
