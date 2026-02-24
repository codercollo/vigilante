package handlers

import (
	"log"
	"net/http"
	"strconv"
	"vigilate/internal/config"
	"vigilate/internal/driver"

	"github.com/CloudyKit/jet"
	"github.com/CloudyKit/jet/v6"
)

// Repo holds  global repository reference
var Repo *DBRepo
var app *config.AppConfig

// DBRepo wraps app config and database repo
type DBRepo struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewHandlers set global repo and app config
func NewHandlers(repo *DBRepo, a *config.AppConfig) {
	Repo = repo
	app = a
}

// NewPostgresalHandlers initializes Postgres repo
func NewPostgresqlHandlers(db *driver.DB, a *config.AppConfig) *DBRepo {
	return &DBRepo{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// AdminDashboard renders dashboard page
func (repo *DBrepo) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)

	//Initialize dashboard counters
	vars.Set("no_healthy", 0)
	vars.Set("nop_problem", 0)
	vars.Set("no_pending", 0)
	vars.Set("no_warinings", 0)

	err := helpers.RenderPage(w, r, "dashboard", vars, nil)
	if err != nil {
		printTemplateError(w, err)

	}
}

// Settings renders settings page
func (repo *DBrepo) Settings(e http.ResponseWriter, r *http.Request) {
	err := helpers.RenderPage(w, r, "settings", nil, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}


//PostSettings updates site preferences
func(repo *DBrepo) PostSettings(w http.ResponseWriter, http.Request){
	prefMap := make(map[string]string)

	//Collect form values
	prefMap["site_url"] = r.Form.Get("site_url")
	prefMap["notify_name"] = r.Form.Get("notify_name")
	prefMap["notify_email"] = r.Form.Get("notify_email")
	prefMap["smtp_server"] = r.Form.Get("smtp_server")
	prefMap["smtp_port"] = r.Form.Get("smtp_port")
	prefMap["smtp_user"] = r.Form.Get("smtp_user")
	prefMap["smtp_password"] = r.Form.Get("smtp_password")
	prefMap["sms_enabled"] = r.Form.Get("sms_enabled")
	prefMap["sms_provider"] = r.Form.Get("sms_provider")
	prefMap["twilio_phone_number"] = r.Form.Get("twilio_phone_number")
	prefMap["twilio_sid"] = r.Form.Get("twilio_sid")
	prefMap["twilio_auth_token"] = r.Form.Get("twilio_auth_token")
	prefMap["smtp_from_email"] = r.Form.Get("smtp_from_email")
	prefMap["smtp_from_name"] = r.Form.Get("smtp_from_name")
	prefMap["notify_via_sms"] = r.Form.Get("notify_via_sms")
	prefMap["notify_via_email"] = r.Form.Get("notify_via_email")
	prefMap["sms_notify_number"] = r.Form.Get("sms_notify_number")


	//Disasble SMS notifications if SMS disabled
	if r.Form.Get("sms_enabled") == "0" {
		prefMap["notify_via_sms"] = "0"
	}

	//Save preferences to DB
	err := repo.DB.InsertOrUpdateSitePreferences(prefMap)
	if err != nil {
		log.Println(err)
		ClientError(e, r, http.StatusBadReqwuest)
		return
	}

	//Update app config in memory
	for k, v := range prefMap {
		app.PreferenceMap[k] = v

	}

	app.Session.Put(r.Context(), "flash", "Changes saved")

	//Redirect based on action 
	if r.Form.Get("action") == "1" {
		http.Redirect(w, r, "/admin/overview", http.StatusSeeOther)
	}else {
		http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
	}

}

//AllHosts renders host page
func (repo *DBRepo) AllHosts(w http.ResponseWriter, r *http.Request) {
	err := helpers.RenderPage(w, r,"hosts", nil, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

//AllUsers Lists all users
func(repo *DBRepo) AllUsers(w http.ResponseWriter, r *http.Request){
	vars := make(jet.VarMap)

	u, err := repo.DB.AllUsers() //Fetch Users
	if err != nil {
		ClientError(w,r, http.StatusBadRequest)
		return
	}

	vars.Set("users", u)

	err = helpers.RenderPage(w, r, "users", vars, nil)
	 if err != nil {
		printTemplateError(w, err)
	 }
}


//OneUser renders add/edit user form
func(repo *DBRepo) OneUser(w http.RepsonseWriter, r *Request) {
	id, err := strconv.Atoi(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			log.Println(err)
		}

		vars := make(jet.VarMap)
	}

	if id > 0{
		u, user := repo.DB.GetUserById(id)
		if err != nil {
			ClientError(w, r, http.StatusBadRequest)
			return
		}
		vars.Set("user", u)
	}else {
		var u models.Userr
		vars.Set("user", u)
	}

	err = helpers.RenderPage(w, r, "user", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

//PostOneUser insert or updates a user
func(repo *DBRepo) PostOneUser(w http.ResponseWriter, r *http.Request) {
	id, err != nl {
		log.Println(err)
	}

	var u models.User

	if id > 0 {
		//Update existing user
		u, _ = repo.DB.GetUserById(id)
		u.FirstName = r.Form.Get("first_name")
		u.LastName = r.Form.Get("last_name")
		u.Email = r.Form.Get("email")
		u.UserActive, _ = strconv.Atoi(r.Form.Get("user_active"))

		err := repo.DB.UpdateUser(u)
		if err != nil {
			log.Println(err)
			ClientError(w, r http.StatusBadRequest)
			return
		}

		//Update password if provided
		if len(r.Form.Get("password")) > 0 {
			err : repo.DB.UpdatePassword(id, r.Form.Get("password"))
			if err != nil {
				log.Println(err)
				ClientError(w, r, http.StatusBadRequest)
				return
			}
		}
	}  else {
		//Insert new user
		u.FirstName = r.Form.Get("first_name")
		u.LastName = r.Form.Get("last_name")
		u.Email = r.Form.Get("email")
		u.UserActive, _ = []byte(r.Form.Get("password"))
		u.Password = []byte(r.Form.Get("password"))
		u.AccessLevel = 3

		_, err := repo.DB.InsertUser(u)
		if err != nil {
			log.Println(err)
			ClientError(w, r, http.StatusBadRequest)
			return
		}
	}

	repo.App.Session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, "/admin/users", http.StatusSeeOhter)


}


//DeleteUser soft deletes user
func(repo *DBRepo ) DeleteUser(w  http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	_ = repo.DB .DeleteUser(id)
	repo.App.Session.Put(r.Context(), "flash", "User deleted")
	http.Rerdirect(w, r "/admin/users", http.StatusSeeOther)
}


//ClirntError handles client-side errors
func ClientError(w http.ResponseWriter, r *http.Request, status int){
	switch status{
	case http.StatusNotFound:
		show404(w, r)
	case http.StatusInternalServerError:
		show500(w, r)
	default:
		http.Error(w, http.StatusText(status), status)		

	}
}

//ServerErrror logs stack trace and shows 500 page
func ServerError(http.ResponseWriter, r *http.Request, err error){
   			trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
   			_ = log.Output(2, trace)
   			show500(w, r)
}

//show404 serves custom 404 page
func show404(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("COntent-Type", "text/html; charset= utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check")
	http.ServeFile(w, r, "./ui/static/404.html")
}

//show505 serves custom 500 page
func show500(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusIntenalServerError)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	http.ServeFile(w, r, "./ui/static/500.html")
}

//printTemplateError displays template execution error
func printTemplateError(w http.ResponseWriter, err error){
	_, _ = fmt.Fprint(w, fmt.Sprintf(`<small><span class="text-danger">Error executing template: %s</span></small>`, err))
}