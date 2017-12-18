package daemon

import (
	"github.com/pawelsocha/kryptond/config"
	. "github.com/pawelsocha/kryptond/logging"
)

//Main main funtion to launch kryptond
func Main() {

	config, err := config.New(ConfigFile)
	if err != nil {
		Log.Critical("Can't read configuration. Error: ", err)
		return
	}
	Log.Info("Config: ", config)
	Log.Debug("Debug oh")
}
