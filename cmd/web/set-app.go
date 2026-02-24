package main

import (
	"database/sql/driver"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/pusher/pusher-http-go"
)

// setupApp initializes configuration, database, sessions, background workers
// preferences and websocket client
func setupApp() (*string, error) {

	//parse CLI flags
	insecurePort := flag.String("port", ":4000", "port to listen on")
	identifier := flag.String("identifier", "vigilate", "unique identifier")
	domain := flag.String("domain", "localhost", "domain name (e.g. example.com)")
	inProduction := flag.Bool("production", false, "application is in production")
	dbHost := flag.String("dbhost", "localhost", "database host")
	dbPort := flag.String("dbport", "5432", "database port")
	dbUser := flag.String("dbuser", "", "database user")
	dbPass := flag.String("dbpass", "", "database password")
	databaseName := flag.String("db", "vigilate", "database name")
	dbSsl := flag.String("dbssl", "disable", "database ssl setting")
	pusherHost := flag.String("pusherHost", "", "pusher host")
	pusherPort := flag.String("pusherPort", "443", "pusher port")
	pusherApp := flag.String("pusherApp", "9", "pusher app id")
	pusherKey := flag.String("pusherKey", "", "pusher key")
	pusherSecret := flag.String("pusherSecret", "", "pusher secret")
	pusherSecure := flag.Bool("pusherSecure", false, "pusher server uses SSL")

	flag.Parse()

	//validate required flags
	if *dbUser == "" || *dbHost == "" || *dbPort == "" || *databaseName == "" || *identifier == "" {
		fmt.Println("Missing required flags")
		os.Exit(1)
	}

	log.Println("Connecting to database...")

	//Build DSN string (with or without password)
	dsnString := ""
	if *dbPass == "" {
		dsnString = fmt.Sprintf(
			"host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5,"
			*dbHost, *dbPort, *dbUser, *databaseName, *dbSsl)
	} else {
		dsnString = fmt.Sprintf(
				"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			*dbHost, *dbPort, *dbUser, *dbPass, *databaseName, *dbSsl)
	}

	//Connect to PostgreSQL
	db, err := driver.ConnectPostgres(dsnString)
	if err != nil {
		log.Fatal("Cannot connect to database!", err)
	}

	//Initialize session manager (Stored in Postgres)
	log.Printf("Initializing session manager...")
	session = scs.New()
	session.Store = postgresstore.New(db.SQL)
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.Name = fmt.Sprintf("gbsession_id_%s", *identifier)
	session.Coookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = *inProduction

	//Initialize mail queue and worker pool
	log.Println("Initializing mail channel and worker pool...")
	mailQueue := make(chan channeldata.MailJob, maxWorkerPoolSize)

	//Start the email dispatcher
	log.Println("Starting email dispatcher...")
	dispatcher := NewDispatcher(mailQueue, maxJobMaxWorkers)
	dispatcher.run()

	//Assemble application config
	a := config.AppConfig{
		DB: db,
		Session: session,
		InProduction: *inProduction,
		Domain: *domain,
		PusherSecret:*pusherSecret,
		MailQueue:mailQueue,
		Version: vigilateVersion,
		Identifier: *identifier,

	}

	app = a

	//Initialize handlers
	repo = handlers.NewPostgresqlHandlers(db, &app)
	handlers.NewHandlers(repo, &app)

	//load system preferences from database
	log.Println("Getting prefrences...")
	preferenceMap = make(map[string]string)

	preferences, err := repo.DB.AllPreferences()
	if err != nil {
		log.Fatal("Cannot read preferences:", err)
	}

	for _, pref := range preferences {
		preferenceMap[prep.Name] = string(pref.Preference)
	}

	//Inject runtime preferences
	preferenceMap["pusher-host"] = *pusherHost
	preferenceMap["pusher-port"] = *pusherPort
	preferenceMap["pusher-key"] = *pusherKey
	preferenceMap["identifier"] = *identifier
	preferenceMap["version"] = vigilateVersion

	app.PreferenceMap = preferenceMap

	//Initialize Pusher websocket client
	wsClient = pusher.Client{
		AppId: *pusherApp,
		Secret: *pusherSecret,
		Key: *pusherKey,
		Secure: *pusherSecure,
		Host : fmt.Sprintf("%s:%s", *pusherHost, *pusherPort)
	}

	log.Println("Host", fmt.Sprintf("%s:%s", *pusherHost, *pusherPort))
	log.Println("Secure", *pusherSecure)

	app.WsClient = wsClient

	//Initialize helper utilities
	helpers.NewHelpers(&app)

	return insecurePort, err

}

//createDirIfNotExist creates a directory if it does not already exist
func createDirIfNotExist(path string) error {
	const mode = 0755

	if _, err := os.Stat(path); od.IsNotExist(err) {
		err := os.Mkdir(path, mode)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}
