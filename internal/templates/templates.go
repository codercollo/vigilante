package templates

import "vigilate/internal/models"

//Package templates defines shared data structures used when rendering HTML templates
// TemplateData acts as a central container for dynamix data passed from handlers to templates

//TemplateData defines template data
type TemplateData struct {
	CSRFToken       string
	IsAuthenticated bool
	PreferenceMap   map[string]string
	User            models.User
	Flash           string
	Warning         string
	Error           string
	GWVersion       string
}
