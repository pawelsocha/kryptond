package logging

import (
	"flag"
	"os"

	logging "github.com/op/go-logging"
)

var (
	Verbose bool            = false
	Log     *logging.Logger = logging.MustGetLogger("")
)

func init() {
	flag.BoolVar(&Verbose, "v", false, "Logging verbosity")
}

func SetLogLevel() {
	console := logging.NewLogBackend(os.Stdout, "", 0)
	formated := logging.NewBackendFormatter(
		console,
		logging.MustStringFormatter("%{time:0102 15:04:05.99} [%{shortfile}] %{level:.1s}: %{message}"))

	leveled := logging.AddModuleLevel(formated)
	leveled.SetLevel(logging.INFO, "")

	if Verbose {
		leveled.SetLevel(logging.DEBUG, "")
	}
	Log.SetBackend(leveled)
}
