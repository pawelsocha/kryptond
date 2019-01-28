package daemon

import (
	"fmt"
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
					Log.Errorf("Error processing %s. Client: %d, Host: %s, Error: %s",
						event.Operation,
						event.CustomerId,
						host,
						err,
					)
					event.Status = "ERR"
					event.Save()
				}
			}
		}
	}
}

func processRemove(event *client.Event, device *mikrotik.Device) error {

	client, err := client.NewClient(event.CustomerId)
	if err != nil {
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
		Log.Errorf("Can't get secret details during removing process. Error: %s", err)
		return err
	}

	ret.Fetch(&sid)

	ret, err = device.ExecuteEntity("print", nid)
	if err != nil {
		if err.Error() == "Id is empty." {
			return nil
		}
		Log.Errorf("Can't get ip>nat details during removing process. Error: %s", err)
		return err
	}

	ret.Fetch(&nid)

	if _, err := device.ExecuteEntity("remove", qid); err != nil {
		if err.Error() != "Id is empty" {
			Log.Errorf("Can't remove queue. Error: %s", err)
			return err
		}
	}

	if _, err := device.ExecuteEntity("remove", sid); err != nil {
		Log.Errorf("Can't remove secret. Error: %s", err)
		return err
	}

	if _, err := device.ExecuteEntity("remove", nid); err != nil {
		Log.Errorf("Can't remove ip>nat. Error: %s", err)
		return err
	}

	return nil
}

func processUpdate(event *client.Event, device *mikrotik.Device) error {
	Log.Infof("Processing change event for client %d", event.CustomerId)

	client, err := client.NewClient(event.CustomerId)
	if err != nil {
		return err
	}

	if len(client.Nodes) == 0 {
		return fmt.Errorf("Node is empty")
	}

	//here for for client nodes
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
		Log.Errorf("Can't get queue details. Error: %s", err)
		return err
	}

	action = "add"
	if len(ret.Re) > 0 {
		ret.Fetch(&qid)
		q.ID = qid.ID
		action = "edit"
	}

	_, err = device.ExecuteEntity(action, q)
	if err != nil {
		Log.Errorf("Can't set queue details. Error: %s", err)
		return err
	}

	ret, err = device.ExecuteEntity("print", sid)
	if err != nil {
		Log.Errorf("Can't get secret details. Error: %s", err)
		return err
	}

	action = "add"
	if len(ret.Re) > 0 {
		ret.Fetch(&sid)
		s.ID = sid.ID
		action = "edit"
	}

	_, err = device.ExecuteEntity(action, s)
	if err != nil {
		Log.Errorf("Can't set secret details. Error: %s", err)
		return err
	}

	ret, err = device.ExecuteEntity("print", nid)
	if err != nil {
		Log.Errorf("Can't get ip>nat details. Error: %s", err)
		return err
	}

	if len(ret.Re) > 0 {
		ret.Fetch(&nid)
		if _, err := device.ExecuteEntity("remove", nid); err != nil {
			Log.Errorf("Can't remove ip>nat. Error: %s", err)
			return err
		}
	}

	nat := mikrotik.Nat{
		SrcAddress: node.IP,
		Comment:    fmt.Sprintf("%s", node.Name),
	}

	if node.Warning == 1 {
		nat.Action = "dst-nat"
		nat.ToAddress = device.Community
		nat.Chain = "dstnat"
	} else {
		nat.Action = "src-nat"
		nat.ToAddress = node.Public
		nat.Chain = "srcnat"
	}

	_, err = device.ExecuteEntity("add", nat)
	if err != nil {
		Log.Errorf("Can't add ip>nat details. Error: %s", err)
		return err
	}
	event.Finish = time.Now()
	event.Status = "DONE"
	event.Save()

	return nil
}
