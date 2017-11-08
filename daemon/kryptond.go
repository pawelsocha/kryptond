package daemon

import (
	"fmt"
	"github.com/pawelsocha/kryptonlms/config"
)

//Main main funtion to launch kryptond
func Main() {
	fmt.Printf("ConfigFile %s", ConfigFile)
	c, err := config.New(ConfigFile)
	fmt.Printf("C %v %s", c, err)
}
