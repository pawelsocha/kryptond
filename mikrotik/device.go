package mikrotik

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/pawelsocha/kryptond/config"
	. "github.com/pawelsocha/kryptond/logging"
	routeros "github.com/pawelsocha/routeros"
)

// Device structure to describe a routeros api instance
type Device struct {
	Host string
	Conn *routeros.Client
	Job  chan Task
	done chan bool
}

// NewDevice create routeros api client
func NewDevice(host string) (*Device, error) {
	var err error
	mk := new(Device)

	tlsConfig := tls.Config{}

	if config.Cfg.Mikrotik.Insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	mk.Conn, err = routeros.DialTLSTimeout(
		fmt.Sprintf("%s:8729", host),
		config.Cfg.Mikrotik.Username,
		config.Cfg.Mikrotik.Password,
		&tlsConfig,
		time.Second*5,
	)

	if err != nil {
		return nil, err
	}

	mk.Host = host
	return mk, nil
}

func (device *Device) Execute(cmds ...string) (*routeros.Reply, error) {
	return device.Conn.RunArgs(cmds)
}

func (device *Device) executeTask(task Task) {
	Log.Debugf("%s: Execute %s", device.Host, task.Command)
	ret, err := device.Execute(strings.Split(task.Command, " ")...)

	if err != nil {
		Log.Errorf("Can't execute %s on %s. Error: %s", task.Command, device.Host, err)
		task.Result <- nil
		return
	}

	task.Result <- ret
}

//Run execute command asynchronously
func (device *Device) Run() {
	device.done = make(chan bool, 1)
	device.Job = make(chan Task)

	go func() {
		for {
			select {
			case task := <-device.Job:
				Log.Debugf("Got job %v", task.Command)
				device.executeTask(task)
			case <-device.done:
				Log.Infof("%s: exiting", device.Host)
				return
			}
		}
	}()
}

//Task return task channel
func (device *Device) Task() chan Task {
	return device.Job
}

//Stop async client
func (device *Device) Stop() error {
	device.done <- true
	close(device.done)
	return nil
}
