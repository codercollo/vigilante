package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"vigilate/internal/helpers"

	"github.com/justinas/nosurf"
)

//This file contains HTTP middleware for the Vigilate application
//It includes session loading, authentication checks, CSRF protection
//panic recovery and "remember me" functionality

// SessionLoad loads the session from every incoming requests
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// Auth checks if a user is authenitcated before allowing access to a route
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			url := r.URL.Path
			http.Redirect(w, r, fmt.Sprintf("/?target=%s", url), http.StatusFound)
			return
		}
		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

// RecoverPanic recovers from panic(s) during HTTP request handling
// and returns a 500 Internal Server Error page instead of crashing the server
func RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// Check if there has been a panic
			if err := recover(); err != nil {
				// return a 500 Internal Server response
				helpers.ServerError(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// NoSurf adds CSRF protection to HTTP requests using the nosurf package
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	//Exempt certain routes from CSRF checks
	csrfHandler.ExemptPath("/pusher/auth")
	csrfHandler.ExemptPath("/pusher/hook")

	//Configure CSRF cookie
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProducion,
		SameSite: http.SameSiteStrictMode,
		Domain:   app.Domain,
	})

	return csrfHandler
}

// CheckRemember checks for "remember me" cookies and logs the user in automatically
// if a valid token exists
func CheckRemember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			cookie, err := r.Cookie(fmt.Sprintf("_%s_gowatcher_remember", preferenceMap["identifier"]))
			if err != nil {
				next.ServeHTTP(w, r)
			} else {
				key := cookie.Value
				if len(key) > 0 {
					// split the cookie value to get user ID and hash
					split := strings.Split(key, "|")
					uid, hash := split[0], split[1]
					id, _ := strconv.Atoi(uid)

					//validate the remember token
					validHash := repo.DB.CheckForToken(id, hash)
					if validHash {
						// valid remember me token, so log the user in;
						// populate session with user info

						_ = session.RenewToken(r.Context())
						user, _ := repo.DB.GetUserById(id)
						hashedPassword := user.Password
						session.Put(r.Context(), "userID", id)
						session.Put(r.Context(), "userName", user.FirstName)
						session.Put(r.Context(), "userFirstName", user.FirstName)
						session.Put(r.Context(), "userLastName", user.LastName)
						session.Put(r.Context(), "hashedPassword", string(hashedPassword))
						session.Put(r.Context(), "user", user)
						next.ServeHTTP(w, r)
					} else {
						// invalid token, so delete the cookie and log-out
						deleteRememberCookie(w, r)
						session.Put(r.Context(), "error", "You've been logged out from another device!")
						next.ServeHTTP(w, r)
					}
				} else {
					// key length is zero, so it's a leftover cookie (user has not closed browser)
					next.ServeHTTP(w, r)
				}
			}
		} else {
			// they are logged in, but make sure that the remember token has not been revoked
			cookie, err := r.Cookie(fmt.Sprintf("_%s_gowatcher_remember", preferenceMap["identifier"]))
			if err != nil {
				// no cookie
				next.ServeHTTP(w, r)
			} else {
				key := cookie.Value
				// have a remember token, but make sure it's valid
				if len(key) > 0 {
					split := strings.Split(key, "|")
					uid, hash := split[0], split[1]
					id, _ := strconv.Atoi(uid)
					validHash := repo.DB.CheckForToken(id, hash)
					if !validHash {
						deleteRememberCookie(w, r)
						session.Put(r.Context(), "error", "You've been logged out from another device!")
						next.ServeHTTP(w, r)
					} else {
						next.ServeHTTP(w, r)
					}
				} else {
					next.ServeHTTP(w, r)
				}
			}
		}
	})
}

// deleteRememberCookie deletes the remember me cookie, and logs the user out, clears the session
func deleteRememberCookie(w http.ResponseWriter, r *http.Request) {
	_ = session.RenewToken(r.Context())
	// delete the cookie
	newCookie := http.Cookie{
		Name:     fmt.Sprintf("_%s_ggowatcher_remember", preferenceMap["identifier"]),
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		Domain:   app.Domain,
		MaxAge:   -1,
		Secure:   app.InProducion,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &newCookie)

	// log them out
	session.Remove(r.Context(), "userID")
	_ = session.Destroy(r.Context())
	_ = session.RenewToken(r.Context())
}
