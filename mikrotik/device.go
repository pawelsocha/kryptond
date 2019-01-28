package mikrotik

import (
	"fmt"

	. "github.com/pawelsocha/kryptond/logging"
	"github.com/pawelsocha/routeros"
)

type Device struct {
	Host      string
	Community string
	Conn      Client
	Job       chan Task
	done      chan bool
}

// NewDevice create routeros api client
func NewDevice(client Client, host string) *Device {
	mk := new(Device)
	mk.Host = host
	mk.Conn = client
	return mk
}

//Execute wrapper to routeros RunArgs command
func (device *Device) Execute(cmds ...string) (*routeros.Reply, error) {
	return device.Conn.RunArgs(cmds)
}

//Disconnect - release the connetion
func (device *Device) Disconnect() {
	device.Conn.Close()
}

//ExecuteEntity - perform the mikrotik action on speficic entitity.
func (device *Device) ExecuteEntity(action string, entity Entity) (*routeros.Reply, error) {
	var ret *routeros.Reply = nil
	var err error

	defer device.Disconnect()

	switch action {
	case "print":
		ret, err = device.Print(entity)
	case "remove":
		err = device.Remove(entity)
	case "add":
		err = device.Add(entity)
	case "edit":
		err = device.Edit(entity)
	default:
		err = fmt.Errorf("unknown action %s", action)
	}

	if err != nil {
		Log.Errorf("can't execute %#v on %s. Error: %s", entity, device.Host, err)
	}

	return ret, err
}

func (device *Device) executeTask(task Task) {
	Log.Debugf("%s: Execute %s, Action: %s", device.Host, task.Entity.Path, task.Action)
	ret, err := device.ExecuteEntity(task.Action, task.Entity)

	task.Result <- &Result{
		Reply: ret,
		Error: err,
	}
}

//Run execute command asynchronously
func (device *Device) Run() {
	device.done = make(chan bool, 1)
	device.Job = make(chan Task)

	go func() {
		for {
			select {
			case task := <-device.Job:
				Log.Debugf("Got job %#v", task)
				device.executeTask(task)
			case <-device.done:
				Log.Infof("%s: exiting", device.Host)
				return
			}
		}
	}()
}

//Task return task channel
func (device *Device) TaskChan() chan Task {
	return device.Job
}

func (device *Device) Print(i Entity) (*routeros.Reply, error) {
	sentence := []string{
		fmt.Sprintf("%s/print", i.Path()),
	}

	attrs := i.PrintAttrs()
	if attrs != "" {
		sentence = append(sentence, attrs)
	}

	where := i.Where()
	if where != "" {
		sentence = append(sentence, where)
	}

	plist := PropertyList(i)
	if plist != "" {
		sentence = append(sentence, "=.proplist="+plist)
	}

	return device.Conn.RunArgs(sentence)
}

func (device *Device) Remove(i Entity) error {
	id := i.GetId()
	if id == "" {
		return fmt.Errorf("Id is empty.\n")
	}

	sentence := []string{
		fmt.Sprintf("%s/remove", i.Path()),
		fmt.Sprintf("=.id=%s", id),
	}
	_, err := device.Conn.RunArgs(sentence)
	return err
}

func (device *Device) Edit(i Entity) error {
	id := i.GetId()
	if id == "" {
		return fmt.Errorf("Id is empty.\n")
	}
	sentence := []string{
		fmt.Sprintf("%s/set", i.Path()),
	}

	sentence = append(sentence, ValueList(i)...)

	_, err := device.Conn.RunArgs(sentence)
	return err
}

func (device *Device) Add(i Entity) error {
	sentence := []string{
		fmt.Sprintf("%s/add", i.Path()),
	}

	sentence = append(sentence, ValueList(i)...)

	_, err := device.Conn.RunArgs(sentence)
	return err

}

//Stop async client
func (device *Device) Stop() error {
	device.done <- true
	close(device.done)
	return nil
}
