package mikrotik

import (
	. "github.com/pawelsocha/kryptond/logging"
	"github.com/pawelsocha/kryptond/router"
)

type Workers struct {
	Nodes map[string]*Device
}

func NewWorkers() *Workers {
	return &Workers{
		Nodes: make(map[string]*Device),
	}
}

func (w *Workers) AddNewDevice(r router.Router) (*Device, error) {

	ip := r.PrivateAddress
	if r.PublicAddress != "" && r.PublicAddress != "0.0.0.0" {
		ip = r.PublicAddress
	}

	device := NewDevice(ip)
	device.Run()
	device.Community = r.Community

	w.Nodes[ip] = device
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

func (w *Workers) GetDevice(host string) *Device {
	if device, ok := w.Nodes[host]; ok {
		return device
	}

	return nil
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
