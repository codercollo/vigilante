package main

import (
	"net/http"
	"vigilate/internal/handlers"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5"
)

// routes sets up all application routes and middleware
func routes() http.Handler {

	//Create Router
	mux := chi.NewRouter

	//Global middleware

	//Load session data
	mux.Use(SessionLoad)
	//Recover from panics
	mux.Use(RecoverPanic)
	//CSRF protection
	mux.Use(NoSurf)
	//Remember-me authenticate
	mux.Use(CheckRemeber)

	//Authentication routes
	mux.Get("/", handlers.Repo.LoginScreen)
	mux.Post("/", handlers.Repo.Login)
	mux.Get("/user/logout", handlers.Repo.Logout)

	//Admin routes (protected)
	mux.Route("/admin", func(mux chi.Router) {
		//Require authentication
		mux.Use(Auth)

		//Dashboard
		mux.Get("/overview", handlers.Repo.AdminDashboard)

		//Events
		mux.Get("/events", handler.Repo.Events)

		//Settings
		mux.Get("/settings", handlers.Repo.Settings)
		mux.Post("/settings", handlers.Repo.PostSettings)

		//Service status pages
		mux.Get("/all-health", handlers.Repo.AllHealthServices)
		mux.Get("/all-warning", handlers.Repo.AllWarningServices)
		mux.Get("/all-problems", handlers.Repo.AlProblemsServices)
		mux.Get("/all-pending", handlers.Repo.AllPendingServices)

		//User managament
		mux.Get("/users", handlers.Repo.AllUsers)
		mux.Get("/users/{id}", handlers.Repo.OneUser)
		mux.Post("/user/{id}", handlers.Repo.PostOneUser)
		mux.Get("/user/delete/{id}", handlers.DeleteUser)

		//Schedule
		mux.Get("/schedule", handlers.Repo.ListEntries)

		//Host management
		mux.Get("/host/all", handlers.Repo.AllHosts)
		mux.Get("/host/{id}", handlers.Repo.Hosts)

	})

	//Serve static files
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static", http.StripPrefix("/static", fileServer))

	return mux
}
