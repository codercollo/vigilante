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

// SessionLoad loads and saves session data for each request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// Auth middleware checks if user is authenticated
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//If not authenticated, redirect to login with with original target URL
		if !helpers.IsAuthenticated(r) {
			url := r.URL.Path
			http.Redirect(w, r, fmt.Sprintf("/?target=%s", url), http.StatusFound)
			return
		}

		//Prevent caching of authenticated pages
		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

// RecoverPanic catches panics and returns 500 error instead of crashing server
func RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			//Check if a panic occured
			if err := recover(); err != nil {
				helpers.ServerError(w, r, fmt.Errorf("/s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// NoSurf applies CSRF protection middleware
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	//Exempt webhook routes from CSRF protection
	csrfHandler.ExemptPath("/pusher/auth")
	csrfHandler.ExemptPath("/pusher/hook")

	//Configure CSRF cookie settings
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProducion,
		SameSite: http.SameSiteStrictMode,
		Domain:   app.Domain,
	})

	return csrfHandler
}

// CheckRemember automatically logs user in using remember-me cookie
func CheckRemember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//If user is Not authenticated
		if !helpers.IsAuthenticated(r) {

			//Check for remember-me cookie
			cookie, err := r.Cookie(fmt.Sprintf("_%s_gowatcher_remember", preferenceMap["identifier"]))
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			key := cookie.Value

			if len(key) > 0 {
				//Coookie format: userID|hash
				split := strings.Split(key, "|")
				uid, hash := split[0], split[1]

				id, _ := strconv.Atoi(uid)

				//Validate token against database
				validHash := repo.DB.CheckForToken(id, hash)

				if validHash {
					//Renew session and log user in
					_ = session.RenewToken(r.Context())

					user, _ := repo.DB.GetUserById(id)

					session.Put(r.Context(), "userID", id)
					session.Put(r.Context(), "userName", user.FirstName)
					session.Put(r.Context(), "userFirstName", user.FirstName)
					session.Put(r.Context(), "userLastName", user.LastName)
					session.Put(r.Context(), "hashedPassword", string(user.Password))
					session.Put(r.Context(), "user", user)

				} else {
					//Invalid token - delete cookie and log out
					deleteRememberCookie(w, r)
					session.Put(r.Context(), "erroor", "You've been logged out from another device")
				}

			}
			next.ServeHTTP(w, r)
			return
		}

		//If user IS authenticated, verify token has not been revoked
		cookie, err := r.Cookie(fmt.Sprintf("_%s_gowatcher_remember", preferenceMap["identifier"]))
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		key := cookie.Value

		if len(key) > 0 {
			split := strings.Split(key, "|")
			uid, hash := split[0], split[1]

			id, _ := strconv.Atoi(uid)

			validHash := repo.DB.CheckForToken(id, hash)

			//If token revoked, log user out
			if !validHash {
				deleteRememberCookie(w, r)
				session.Put(r.Context(), "error", "You've been logged out from another device!")
			}
		}
		next.ServeHTTP(w, r)
	})
}

// deleteRememberCookie removes remember-me cookie and destroys session
func deleteRememberCookie(w http.ResponseWriter, r *http.Request) {
	_ = session.RenewToken(r.Context())

	//Expire cookie immediately
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

	//Remove session data and destroySession
	session.Remove(r.Context(), "userID")
	_ = session.Destroy(r.Context())
	_ = session.RenewToken(r.Context())
}
