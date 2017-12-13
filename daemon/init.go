package daemon

import (
	"flag"

	Log "github.com/pawelsocha/kryptond/logging"
)

var (
	ConfigFile string
)

func init() {
	flag.StringVar(&ConfigFile, "config", "/etc/lms/lms.ini", "Path to lms config file")
	flag.Parse()
	Log.SetLogLevel()
}
