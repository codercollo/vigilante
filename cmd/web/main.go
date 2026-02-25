package main

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
	"vigilate/internal/config"
	"vigilate/internal/handlers"

	"github.com/alexedwards/scs/v2"
	"github.com/pusher/pusher-http-go"
)

// Application-wide configuration
var app config.AppConfig

// Database repository
var repo *handlers.DBRepo

// Session manager
var session *scs.SessionManager
var preferenceMap map[string]string

// Websocket client (Pusher)
var wsClient pusher.Client

const vigilateVersion = "1.0.0"
const maxWorkerPoolSize = 5
const maxJobMaxWorkers = 5

// init runs before main and performs application-level setup
// register session types and configuring environment settings
func init() {
	gob.Register(models.User{})
	_ = os.Setenv("TZ", "Africa/Nairobi")

}

// main is the application entry point
func main() {
	//Initialize application configuration
	insecurePort, err := setupApp()
	if err != nil {
		log.Fatal(err)
	}

	//Ensure cleanup on shutdown
	defer close(app.MailQueue)
	defer app.DB.SQL.Close()

	//startup information
	log.Printf("******************************************")
	log.Printf("** %sVigilate%s v%s built in %s", "\033[31m", "\033[0m", vigilateVersion, runtime.Version())
	log.Printf("**----------------------------------------")
	log.Printf("** Running with %d Processors", runtime.NumCPU())
	log.Printf("** Running on %s", runtime.GOOS)
	log.Printf("******************************************")

	//configure HTTP server
	srv := &http.Server{
		Addr:              *insecurePort,
		Handler:           routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	log.Printf("Starting HTTP server on port %s...", *insecurePort)

	//Start HTTP server
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
