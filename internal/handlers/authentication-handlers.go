package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"vigilate/internal/helpers"
	"vigilate/internal/models"
)

//Package handler contains HTTP handler functions for user authentication

// LoginScreen shows the home (login) screen
func (repo *DBRepo) LoginScreen(w http.ResponseWriter, r *http.Request) {
	// if already logged in, take to dashboard
	if repo.App.Session.Exists(r.Context(), "userID") {
		http.Redirect(w, r, "/admin/overview", http.StatusSeeOther)
		return
	}

	//Render the login template
	err := helpers.RenderPage(w, r, "login", nil, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// Login handles user login form submission and session management
func (repo *DBRepo) Login(w http.ResponseWriter, r *http.Request) {
	//Renew session token for security
	_ = repo.App.Session.RenewToken(r.Context())

	//Parse form input
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	//Authenticate user credentials
	id, hash, err := repo.DB.Authenticate(r.Form.Get("email"), r.Form.Get("password"))
	if err == models.ErrInvalidCredentials {
		app.Session.Put(r.Context(), "error", "Invalid login")
		err := helpers.RenderPage(w, r, "login", nil, nil)
		if err != nil {
			printTemplateError(w, err)
		}
		return
	} else if err == models.ErrInactiveAccount {
		app.Session.Put(r.Context(), "error", "Inactive account!")
		err := helpers.RenderPage(w, r, "login", nil, nil)
		if err != nil {
			printTemplateError(w, err)
		}
		return
	} else if err != nil {
		log.Println(err)
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	//Handle "remember me"
	if r.Form.Get("remember") == "remember" {
		//Generate a random token for persistent login
		randomString := helpers.RandomString(12)
		hasher := sha256.New()

		_, err = hasher.Write([]byte(randomString))
		if err != nil {
			log.Println(err)
		}

		sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		//Save token to a database
		err = repo.DB.InsertRememberMeToken(id, sha)
		if err != nil {
			log.Println(err)
		}

		// Set remember-me cookie in browser
		expire := time.Now().Add(365 * 24 * 60 * 60 * time.Second)
		cookie := http.Cookie{
			Name:     fmt.Sprintf("_%s_gowatcher_remember", app.PreferenceMap["identifier"]),
			Value:    fmt.Sprintf("%d|%s", id, sha),
			Path:     "/",
			Expires:  expire,
			HttpOnly: true,
			Domain:   app.Domain,
			MaxAge:   315360000, // seconds in year
			Secure:   app.InProducion,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, &cookie)
	}

	// Fetch full user info from database
	u, err := repo.DB.GetUserById(id)
	if err != nil {
		log.Println(err)
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	//Store user data and authentication info in session
	app.Session.Put(r.Context(), "userID", id)
	app.Session.Put(r.Context(), "hashedPassword", hash)
	app.Session.Put(r.Context(), "flash", "You've been logged in successfully!")
	app.Session.Put(r.Context(), "user", u)

	//Redirect to target page or default dashboard
	if r.Form.Get("target") != "" {
		http.Redirect(w, r, r.Form.Get("target"), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/overview", http.StatusSeeOther)
}

// Logout logs the user out and ends user's session and clears cookies
func (repo *DBRepo) Logout(w http.ResponseWriter, r *http.Request) {

	// delete the remember me token, if any
	cookie, err := r.Cookie(fmt.Sprintf("_%s_gowatcher_remember", app.PreferenceMap["identifier"]))
	if err != nil {
	} else {
		key := cookie.Value
		// have a remember token, so get the token
		if len(key) > 0 {
			// key length > 0, so it might be a valid token
			split := strings.Split(key, "|")
			hash := split[1]
			err = repo.DB.DeleteToken(hash)
			if err != nil {
				log.Println(err)
			}
		}
	}

	// delete the remember me cookie, if any
	delCookie := http.Cookie{
		Name:     fmt.Sprintf("_%s_gowatcher_remember", app.PreferenceMap["identifier"]),
		Value:    "",
		Domain:   app.Domain,
		Path:     "/",
		MaxAge:   0,
		HttpOnly: true,
	}
	http.SetCookie(w, &delCookie)

	//Destroy session data
	_ = app.Session.RenewToken(r.Context())
	_ = app.Session.Destroy(r.Context())
	_ = app.Session.RenewToken(r.Context())

	//Flash logout message
	repo.App.Session.Put(r.Context(), "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
