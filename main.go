package main

import (
	"LOIL-server/db"
	"LOIL-server/server"
	"log"
	"os"
)

const portDefault int = 7540

func main() {
	//api.SetPassword()

	logger := log.New(os.Stdout, "server: ", log.LstdFlags)

	err := db.Init(logger)
	if err != nil {
		logger.Fatal("FATAL: error while db load: ", err)
	}
	defer db.CloseDatabase()

	srv := server.StartServer(portDefault, logger)
	if err := srv.HTTPServer.ListenAndServe(); err != nil {
		logger.Fatal("FATAL: error while server start: ", err)
	}
}
