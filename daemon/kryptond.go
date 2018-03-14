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
	db, err := database.Database(config)

	if err != nil {
		Log.Critical("Can't connect to database. Error: ", err)
	}

	routers, err := router.GetRoutersList(db)
	if err != nil {
		Log.Critical("Can't get list of routers from database. Error: ", err)
	}

	for _, device := range routers {
		Log.Debugf("Add router %s", device.PrivateAddress)
		_, err := workers.AddNewDevice(device.PrivateAddress)
		if err != nil {
			Log.Critical("Can't connect with router. Error: ", err)
		}
	}

	db.Disconnect()

	for {
		select {
		case <-time.After(time.Second * 60):
			events, err := client.CheckEvents(config)
			if err != nil {
				Log.Critical("Can't get list of events. Error: ", err)
				return
			}

			for _, event := range events {
				client, err := client.NewClient(event.CustomerId, config)
				if err != nil {
					Log.Critical("Can't create new client. Error: ", err)
					return
				}
				Log.Infof("Event: %#v, client: %#v\n", events, client)

				q := mikrotik.Queue{
					Name:    fmt.Sprintf("Client:%d", client.ID),
					Target:  client.Nodes[0].IP,
					Comment: fmt.Sprintf("%d:%d %s - %s ", client.ID, client.Nodes[0].ID, client.Name, client.Nodes[0].Name),
					Limits:  fmt.Sprintf("%d/%d", client.Rate.Upceil, client.Rate.Downceil),
				}
				qid := mikrotik.Queue{
					Name: fmt.Sprintf("Client:%d", client.ID),
				}

				s := mikrotik.Secret{
					Name:     fmt.Sprintf("Client:%d", client.ID),
					Password: client.Nodes[0].Passwd,
					Address:  client.Nodes[0].IP,
					Gateway:  client.Nodes[0].Gateway,
					Comment:  fmt.Sprintf("%d: %s", client.ID, client.Name),
				}

				sid := mikrotik.Secret{
					Name: fmt.Sprintf("Client:%d", client.ID),
				}
				// WIP loop, PoC
				for host, device := range workers.Nodes {
					ret, err := device.ExecuteEntity("print", qid)
					var action string

					if err != nil {
						Log.Fatalf("Can't get queue data from %s. Client: %d, Error: %s", host, client.ID, err)
						continue
					}

					action = "add"
					if len(ret.Re) > 0 {
						ret.Fetch(&qid)
						q.ID = qid.ID
						action = "edit"
					}

					_, err = device.ExecuteEntity(action, q)
					if err != nil {
						Log.Fatalf("Can't %s queue for user %d. Error: %s", action, client.ID, err)
						continue
					}

					ret, err = device.ExecuteEntity("print", sid)
					if err != nil {
						Log.Fatalf("Can't get secret data from %s. Client: %d, Error: %s", host, client.ID, err)
						continue
					}

					action = "add"
					if len(ret.Re) > 0 {
						ret.Fetch(&sid)
						s.ID = sid.ID
						action = "edit"
					}

					_, err = device.ExecuteEntity(action, s)
					if err != nil {
						Log.Fatalf("Can't %s secret for user %d. Error: %s", action, client.ID, err)
						continue
					}

				}
				Log.Infof("Q: %#v", q)
			}
		}
	}
}
