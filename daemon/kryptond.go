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
		Log.Debugf("Add router %#v", device)
		if _, err := workers.AddNewDevice(device); err != nil {
			Log.Critical("Can't connect with router. Error: ", err)
			return
		}
	}

	for {
		select {
		case <-time.After(time.Second * 15):

			events, err := client.CheckEvents()
			if err != nil {
				Log.Critical("Can't get list of events. Error: ", err)
				return
			}

			for _, event := range events {
				host := event.AddrNtoa()
				device := workers.GetDevice(host)

				if device == nil {
					Log.Warningf("Got event %s for client %d with empty device.", event.Operation, event.CustomerId)
					event.Status = "NODEV"
					event.Save()
					continue
				}
				if event.NodeId == 0 {
					Log.Warningf("Got event %s for client %d with empty node id.", event.Operation, event.CustomerId)
					event.Status = "NONODE"
					event.Save()
					continue
				}

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

	client, err := client.NewClient(event.CustomerId)
	if err != nil {
		Log.Critical("Can't create new client. Error: ", err)
		return err
	}

	if len(client.Nodes) == 0 {
		return fmt.Errorf("Node is empty")
	}

	node := client.Nodes[0]

	sid := mikrotik.Secret{
		Name: fmt.Sprintf("%s", node.Name),
	}

	qid := mikrotik.Queue{
		Name: fmt.Sprintf("Client:%d", event.CustomerId),
	}

	nid := mikrotik.Nat{
		Comment: fmt.Sprintf("%s", node.Name),
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

	ret, err = device.ExecuteEntity("print", nid)
	if err != nil {
		if err.Error() == "Id is empty." {
			return nil
		}
		return fmt.Errorf("Nat get error")
	}

	ret.Fetch(&nid)

	if _, err := device.ExecuteEntity("remove", qid); err != nil {
		return fmt.Errorf("Queue remove error")
	}

	if _, err := device.ExecuteEntity("remove", sid); err != nil {
		return fmt.Errorf("Secret remove error")
	}

	if _, err := device.ExecuteEntity("remove", nid); err != nil {
		return fmt.Errorf("Nat remove error")
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

	if len(client.Nodes) == 0 {
		return fmt.Errorf("Node is empty")
	}

	node := client.Nodes[0]

	if node.Access == 0 {
		return processRemove(event, device)
	}

	if node.Public == "" || node.Public == "0.0.0.0" {
		node.Public = device.Host
	}

	nid := mikrotik.Nat{
		Comment: fmt.Sprintf("%s", node.Name),
	}

	sid := mikrotik.Secret{
		Name: fmt.Sprintf("%s", node.Name),
	}

	qid := mikrotik.Queue{
		Name: fmt.Sprintf("Client:%d", client.ID),
	}

	q := mikrotik.Queue{
		Name:    fmt.Sprintf("Client:%d", client.ID),
		Target:  node.IP,
		Comment: fmt.Sprintf("%d:%d %s - %s ", client.ID, node.ID, client.Name, node.Name),
		Limits:  fmt.Sprintf("%dk/%dk", client.Rate.Upceil, client.Rate.Downceil),
	}

	s := mikrotik.Secret{
		Name:     fmt.Sprintf("%s", node.Name),
		Password: node.Passwd,
		Address:  node.IP,
		Gateway:  node.Gateway,
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

	ret, err = device.ExecuteEntity("print", nid)
	if err != nil {
		return fmt.Errorf("Secret get error")
	}

	if len(ret.Re) > 0 {
		ret.Fetch(&nid)
		if _, err := device.ExecuteEntity("remove", nid); err != nil {
			return fmt.Errorf("Nat remove error")
		}
	}

	var nat mikrotik.Nat
	if node.Warning == 1 {
		nat = mikrotik.Nat{
			Action:     "dst-nat",
			ToAddress:  device.Community,
			SrcAddress: node.IP,
			Chain:      "dstnat",
			Comment:    fmt.Sprintf("%s", node.Name),
		}
	} else {
		nat = mikrotik.Nat{
			Action:     "src-nat",
			ToAddress:  node.Public,
			SrcAddress: node.IP,
			Chain:      "srcnat",
			Comment:    fmt.Sprintf("%s", node.Name),
		}
	}

	_, err = device.ExecuteEntity("add", nat)
	event.Finish = time.Now()
	event.Status = "DONE"
	event.Save()

	return nil
}
