package daemon

import (
	"flag"

	Log "github.com/pawelsocha/kryptond/logging"
)

var (
	Verbose    bool = false
	ConfigFile string
)

func init() {
	flag.BoolVar(&Verbose, "v", false, "Logging verbosity")
	flag.StringVar(&ConfigFile, "config", "/etc/lms/lms.ini", "Path to lms config file")
	flag.Parse()

	Log.SetLogLevel(Verbose)
}
