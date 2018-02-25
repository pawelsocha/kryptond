package daemon

import (
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
				Log.Fatal("Can't get list of events. Error: ", err)
				continue
			}
			for _, event := range events {
				client, err := client.NewClient(event.CustomerId, config)
				if err != nil {
					Log.Fatal("Can't create new client. Error: ", err)
					continue
				}
				Log.Infof("Event: %v, client: %v", events, client)
			}
		}
	}
}
