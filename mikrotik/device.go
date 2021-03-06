package mikrotik

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/pawelsocha/kryptond/config"
	. "github.com/pawelsocha/kryptond/logging"
	"github.com/pawelsocha/routeros"
)

// Device structure to describe a routeros api instance
type Device struct {
	Host      string
	Community string
	Conn      *routeros.Client
	Job       chan Task
	done      chan bool
}

// NewDevice create routeros api client
func NewDevice(host string) *Device {
	mk := new(Device)
	mk.Host = host
	return mk
}

//Execute wrapper to routeros RunArgs command
func (device *Device) Execute(cmds ...string) (*routeros.Reply, error) {
	return device.Conn.RunArgs(cmds)
}

//Connect - create connection with routeros
func (device *Device) Connect() error {
	tlsConfig := tls.Config{}

	if config.Cfg.Mikrotik.Insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	conn, err := routeros.DialTLSTimeout(
		fmt.Sprintf("%s:8729", device.Host),
		config.Cfg.Mikrotik.Username,
		config.Cfg.Mikrotik.Password,
		&tlsConfig,
		time.Second*5,
	)

	if err != nil {
		return err
	}
	device.Conn = conn
	return nil
}

//Disconnect - release the connetion
func (device *Device) Disconnect() 	{
	device.Conn.Close()
}


func (device *Device) ExecuteEntity(action string, entity Entity) (*routeros.Reply, error) {
	var ret *routeros.Reply = nil
	var err error

	err = device.Connect()

	if err != nil {
		return ret, err
	}

	defer device.Disconnect()
	
	switch action {
	case "print":
		ret, err = device.Conn.Print(entity)
	case "remove":
		err = device.Conn.Remove(entity)
	case "add":
		err = device.Conn.Add(entity)
	case "edit":
		err = device.Conn.Edit(entity)
	default:
		err = fmt.Errorf("Uknown action %s", action)
	}

	if err != nil {
		Log.Errorf("Can't execute %#v on %s. Error: %s", entity, device.Host, err)
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

//Stop async client
func (device *Device) Stop() error {
	device.done <- true
	close(device.done)
	return nil
}
