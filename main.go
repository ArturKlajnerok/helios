package main

import (
	"fmt"
	"os"

	"github.com/Wikia/helios/helios"
)

func main() {

	displayHelp := false
	configFile := "./config/config.ini"
	if len(os.Args) == 2 && os.Args[1] == "--help" || len(os.Args) > 2 {
		displayHelp = true
	} else if len(os.Args) == 2 {
		configFile = os.Args[1]
	}
	if displayHelp {
		fmt.Printf("Helios OAuth service. Can be started with one argument - path to config file.\n"+
			"If no arguments are provided the default %s will be used.\n", configFile)
	} else {
		helios.NewHelios().Run(configFile)
	}
}
