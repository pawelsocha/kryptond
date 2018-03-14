package mikrotik

import (
	. "github.com/pawelsocha/kryptond/logging"
)

type Workers struct {
	Nodes map[string]*Device
}

func NewWorkers() *Workers {
	return &Workers{
		Nodes: make(map[string]*Device),
	}
}

func (w *Workers) AddNewDevice(host string) (*Device, error) {
	device, err := NewDevice(host)
	if err != nil {
		Log.Errorf("Can't create routeros device for host %s. Error: %s", host, err)
		return nil, err
	}
	device.Run()
	w.Nodes[host] = device
	return device, nil
}

func (w *Workers) Print(entity Entity, result chan *Result) {
	w.sendCommand("print", entity, result)
}

func (w *Workers) Delete(entity Entity, result chan *Result) {
	w.sendCommand("remove", entity, result)
}

func (w *Workers) Update(entity Entity, result chan *Result) {
	w.sendCommand("update", entity, result)
}

func (w *Workers) sendCommand(action string, entity Entity, result chan *Result) {
	for host, device := range w.Nodes {
		task := Task{
			Action: action,
			Entity: entity,
			Result: result,
		}
		j := device.TaskChan()
		Log.Debugf("sending job %#v to %s", task.Entity, host)

		j <- task
	}
}
