package main

import (
	"errors"
	"log"
	"os"

	"github.com/Wikia/helios/helios"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("Provide mysql data source, like: user:pass@tcp(host:port)/dbname")
		panic(errors.New("No data source provided"))
	}
	dataSourceName := os.Args[1]
	helios.NewHelios().Run(dataSourceName)
}
