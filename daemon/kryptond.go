package daemon

import (
	"fmt"
	"log"
	"time"

	"github.com/pawelsocha/kryptond/client"
	"github.com/pawelsocha/kryptond/config"
	"github.com/pawelsocha/kryptond/database"
	. "github.com/pawelsocha/kryptond/logging"
	"github.com/pawelsocha/kryptond/mikrotik"
	"github.com/pawelsocha/kryptond/router"

	"github.com/jinzhu/gorm"
	//mysql
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

//Main main funtion to launch kryptond
func Main() {

	config, err := config.New(ConfigFile)
	if err != nil {
		Log.Critical("Can't read configuration. Error: ", err)
		return
	}
	Log.Info("Starting")
	workers := mikrotik.NewWorkers()
	database.Connection, err = gorm.Open("mysql", config.GetDatabaseDSN())

	if err != nil {
		Log.Critical("Can't connect to database. Error: ", err)
		return
	}

	routers, err := router.GetRoutersList()
	if err != nil {
		Log.Critical("Can't get list of routers from database. Error: ", err)
		return
	}

	for _, device := range routers {
		Log.Debugf("Add router %s", device.PrivateAddress)
		_, err := workers.AddNewDevice(device.PrivateAddress)
		if err != nil {
			Log.Critical("Can't connect with router. Error: ", err)
			return
		}
	}
	for {
		select {
		case <-time.After(time.Second * 60):

			events, err := client.CheckEvents()
			if err != nil {
				Log.Critical("Can't get list of events. Error: ", err)
				return
			}

			for _, event := range events {
				host := event.AddrNtoa()
				device := workers.GetDevice(host)
				Log.Infof("Got event %s for client %d", event.Operation, event.CustomerId)
				switch event.Operation {
				case "CHANGE":
					err = processUpdate(&event, device)
				case "DELETE":
					err = processRemove(&event, device)
				}
				if err != nil {
					log.Fatalf("Error processing %s. Client: %d, Host: %s, Error: %s",
						event.Operation,
						event.CustomerId,
						host,
						err,
					)
				}
			}
		}
	}
}

func processRemove(event *client.Event, device *mikrotik.Device) error {
	sid := mikrotik.Secret{
		Name: fmt.Sprintf("Client:%d", event.CustomerId),
	}

	qid := mikrotik.Queue{
		Name: fmt.Sprintf("Client:%d", event.CustomerId),
	}

	ret, err := device.ExecuteEntity("print", qid)
	if err != nil && err.Error() != "Id is empty." {
		return fmt.Errorf("Queue get error")
	}

	ret.Fetch(&qid)

	ret, err = device.ExecuteEntity("print", sid)
	if err != nil {
		if err.Error() == "Id is empty." {
			return nil
		}
		return fmt.Errorf("Secret get error")
	}

	ret.Fetch(&sid)

	if _, err := device.ExecuteEntity("remove", qid); err != nil {
		return fmt.Errorf("Queue remove error")
	}

	if _, err := device.ExecuteEntity("remove", sid); err != nil {
		return fmt.Errorf("Secret remove error")
	}

	return nil
}

func processUpdate(event *client.Event, device *mikrotik.Device) error {
	Log.Infof("Processing change event for client %d", event.CustomerId)

	client, err := client.NewClient(event.CustomerId)
	if err != nil {
		Log.Critical("Can't create new client. Error: ", err)
		return err
	}

	sid := mikrotik.Secret{
		Name: fmt.Sprintf("Client:%d", client.ID),
	}

	qid := mikrotik.Queue{
		Name: fmt.Sprintf("Client:%d", client.ID),
	}

	q := mikrotik.Queue{
		Name:    fmt.Sprintf("Client:%d", client.ID),
		Target:  client.Nodes[0].IP,
		Comment: fmt.Sprintf("%d:%d %s - %s ", client.ID, client.Nodes[0].ID, client.Name, client.Nodes[0].Name),
		Limits:  fmt.Sprintf("%d/%d", client.Rate.Upceil, client.Rate.Downceil),
	}

	s := mikrotik.Secret{
		Name:     fmt.Sprintf("Client:%d", client.ID),
		Password: client.Nodes[0].Passwd,
		Address:  client.Nodes[0].IP,
		Gateway:  client.Nodes[0].Gateway,
		Comment:  fmt.Sprintf("%d: %s", client.ID, client.Name),
	}

	ret, err := device.ExecuteEntity("print", qid)
	var action string

	if err != nil {
		return fmt.Errorf("Queue get error")
	}

	action = "add"
	if len(ret.Re) > 0 {
		ret.Fetch(&qid)
		q.ID = qid.ID
		action = "edit"
	}

	_, err = device.ExecuteEntity(action, q)
	if err != nil {
		return fmt.Errorf("Queue set error")
	}

	ret, err = device.ExecuteEntity("print", sid)
	if err != nil {
		return fmt.Errorf("Secret get error")
	}

	action = "add"
	if len(ret.Re) > 0 {
		ret.Fetch(&sid)
		s.ID = sid.ID
		action = "edit"
	}

	_, err = device.ExecuteEntity(action, s)
	if err != nil {
		return fmt.Errorf("Secret set error")
	}

	event.Finish = time.Now()
	event.Status = "DONE"
	event.Save()

	return nil
}
