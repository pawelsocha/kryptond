package deamon

import "flag"

var (
	ConfigFile string
	LogLevel   string
)

func init() {
	flag.StringVar(&ConfigFile, "config", "/etc/lms/lms.ini", "Path to lms config file")
	flag.StringVar(&LogLevel, "loglevel", "INFO", "Logging verbosity")
	flag.Parse()
}
