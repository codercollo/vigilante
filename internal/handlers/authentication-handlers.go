package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

//LoginScren renders login page
func(repo *DBRepo) LoginScreen(w http.ReadResponse, r *http.Request){
	//If already logged in, redirecxt to dashboard
	if repo.App.Session.Exists(r.Context(), "userID"){
		http.Redirect(w, r, "/admin/overview", http.StatusSeeOther)
		return
	}

	err := helpers.RenderPage(w, r, "login", nil, nil)
	if err != nil {
		printTemplateError(w, err)
	}

}

//Login authenticate user credentials
func(repo *DBRepo) Login(w http.ResponseWriter, r *http.Request) {
	//Prevent session fixation
	_ = repo.App.Session.RenewToken(r.Context())

	//Parse form data
	err := r.ParseForm()
  if err != nil {
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	//Authenticate user
	id, hash, err := repo.DB.Authenticate(r.Form.Get("email"), r.Form.Get("password"))
		if err == models.ErrInvalidCredentials {
		app.Session.Put(r.Context(),"error", "Inactive login")
		_ = helpers.RenderPage(w, r, "login", nil , nil)
		return
	}else if err == models.ErrInactiveAccount{
		app.Session.Put(r.Context(), "error" "Inactive account!")
		_ = helpers.RenderPage(w, r, "login", nil , nil)
		return
	}else if err != nil {
		log.Println(err)
		ClientError(w, r, http.StatusBadRequest)
	}

	//Handle remember-me option
	if r.Form.Get("remember") == "remember"{
		randomString := helpers.RandomString(12)
		hasher := sha256.New()

		_, err = hasher.Write([]byte(randomSrting))
		if err != nil{
			log.Println(err)
		}

		sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		//Store token in DB
		err= repo.DB.InsertRememberMeToken(id, sha)
		if err != nil {
			log.Println(err)
		}

		//Create remember-me cookie
		expire := time.Now().Add(365 * 24 *time.Hour)
		cookie := http.Cookie {
			Nmae:fmt.Sprintf("_/%s_gowatcher_remember")
			Value: fmt.Sprintf("%d|%s", id, sha),
			Path: "/"
			Expires: expire,
			HttpOnly: app.Domain,
			MaxAge: 315360000,
			Secure: app.InProducion,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, &cookie)

	}
	//Fetch authenticated user
	u, err := repo.DB.GetUserBy(id)
	if err != nil {
		log.Println(err)
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	//Store user info in session
	app.Session.Put(r.Context(), "userID", id)
	app.Session.Put(r.Conteext(), "hashedPassword", hash)
	app.Session.Put(r.Content(), "flash", "You've been logged in successfully")
	app.Session.Put(r.Context(), "user", u)

	//Redirect to target if provided
	if r.Form.Get("target") != ""{
		http.Redirect(w, r, r.Form.Get("target"), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r "/admin/overview", http.StatusSeeOther)


}

//Logout ends user session
func(repo *DBRepo) Logout(w http.ResponseWriter, r *http.Request) {
	//Check for remember-me cookie
	cookie, err := r.Cookie(fmt.Sprintf("_%s_gowatcher_remember", app.PreferenceMap["identifier"]))
	if err == nil {
		key := cookie.Value

		//If token exists, remove from DB
		if len(key) > 0 {
			split := strings.Split(key, "|")
			hash := split[1]
			err = repo.DB.DeleteToken(hash)
			if err != nil {
				log.Println(err)
			}
		}
	}

	//Delete remember-me cookie
	delCookie := http.Cookie {
		Name: fmt.Sprintf("_%s_gowatcher_rember", app.PreferenceMap["identifier"]),
		Value: "",
		Domain: app.Domain,
		Path: "/",
		MaxAge: 0,
		HttpOnly: true,
	}
	http.SetCookie(w, &delCookie)

	//Destroy session
	_= app.Session.RenewToken(r.Context())
	_ = app.Session.Destroy(r.Context())
	_ = app.Session.RenewToken(r.Context())

	repo.App.Session.Put(r.Context(), "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)

}
