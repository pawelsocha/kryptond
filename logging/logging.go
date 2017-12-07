package logging

import (
	"os"

	logging "github.com/op/go-logging"
)

var Log = logging.MustGetLogger("")

func SetLogLevel(verbose bool) {
	console := logging.NewLogBackend(os.Stdout, "", 0)
	formated := logging.NewBackendFormatter(
		console,
		logging.MustStringFormatter("%{time:0102 15:04:05.99} [%{shortfile}] %{level:.1s}: %{message}"))

	leveled := logging.AddModuleLevel(formated)
	leveled.SetLevel(logging.INFO, "")

	if verbose {
		leveled.SetLevel(logging.DEBUG, "")
	}
	Log.SetBackend(leveled)
}
