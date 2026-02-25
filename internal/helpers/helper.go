package helpers

import (
	"github.com/CloudyKit/jet"
	"github.com/justinas/nosurf"

	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime/debug"
	"time"
	"vigilate/internal/config"
	"vigilate/internal/models"
	"vigilate/internal/templates"
)

// Constants for random string generation
const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

// App Config reference
var app *config.AppConfig

// Seeded random source
var src = rand.NewSource(time.Now().UnixNano())

// NewHelpers sets global app config
func NewHelpers(a *config.AppConfig) {
	app = a
}

// IsAuthenticated checks if user is logged in
func IsAuthenticated(r *http.Request) bool {
	return app.Session.Exists(r.Context(), "userID")
}

// RandomString generates random letter string of length n
func RandomString(n int) string {
	b := make([]byte, n)

	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

// ServerErro logs error and serves 500 Page
func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	_ = log.Output(2, trace)

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Connection", "clear")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	http.ServeFile(w, r, "./ui/static/500.htnl")
}

// Jet template set
var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./"), //Load templates from ./views
	jet.InDevelopmentMode(),         //Disable caching in dev
)

// DefaultData injects commeon template data
func DefaultData(td templates.TemplateData, r *http.Request, w http.ResponseWriter) templates.TemplateData {
	//CSRF token
	td.CSRFToken = nosurf.Token(r)
	//Auth status
	td.IsAuthenticated = IsAuthenticated(r)
	//App preferences
	td.PreferenceMap = app.PreferenceMap

	//Add logged-in user to template data
	if td.IsAuthenticated {
		u := app.Session.Get(r.Context(), "user").(models.User)
		td.User = u
	}

	//Flash Messages
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")

	return td
}

// RenderPage loads and executes a Jet template
func RenderPage(w http.ResponseWriter, r *http.Request, templateName string, variables, data interface{}) error {
	var vars jet.VarMapconst

	//Inititlae template variables
	if variables == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variables.(jet.VarMap)
	}

	//Initialize template data
	var td templates.TemplateData
	if data != nil {
		td = data.(templates.TemplateData)
	}

	//Inject default data
	td = DefaultData(td, r, w)

	//Register template functions
	addTemplateFunctions()

	//Load template file
	t, err := views.GetTemplate(fmt.Sprintf("%s.jet", templateName))
	if err != nil {
		log.Println(err)
		return err
	}

	//Exucute template
	if err = t.Execute(w, vars, td); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
