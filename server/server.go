package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Server struct {
	Logger     *log.Logger
	HTTPServer *http.Server
}

// setting the server port and checking the port value in the environment variable
func setPort(port int, logger *log.Logger) string {

	portTodo := os.Getenv("TODO_PORT")

	if portTodo != "" {
		portNew, err := strconv.Atoi(portTodo)
		if err != nil || portNew < 1 || portNew > 65535 {
			logger.Printf("WARN: invalid port %s, is using port %d\n", portTodo, port)
		} else {
			port = portNew
		}
	}
	logger.Printf("INFO: starting the server on the port %d\n", port)

	return fmt.Sprintf(":%d", port)
}

func StartServer(portDefault int, logger *log.Logger) *Server {

	port := setPort(portDefault, logger)

	mux := http.NewServeMux()

	//api.Init(mux, logger)

	server := &http.Server{
		Addr:         port,
		Handler:      mux,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return &Server{
		Logger:     logger,
		HTTPServer: server,
	}
}
