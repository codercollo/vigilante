package handlers

import (
	"net/http"
	"vigilate/internal/helpers"
)

// AllHealthyServices renders healthy services page
func (repo *DBRepo) AllHealthyServices(w http.ResponseWriter, r *http.Request) {
	err := helpers.RenderPage(w, r, "healthy", nil, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllWarningServices renders warning services page
func (repo *DBRepo) AllWarningServices(w http.ResponseWriter, r *http.Request) {
	err := helpers.RenderPage(w, r, "warning", nil, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllProblemsServices renders problemn services page
func (repo *DBRepo) AllProblemServices(w http.ResponseWriter, r *http.Request) {
	err := helpers.RenderPage(w, r, "problem", nil, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllPendingServices renders pending services page
func (repo *DBRepo) AllPendingServices(w http.ResponseWriter, r *http.Request) {
	err := helpers.RenderPage(w, r, "pending", nil, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}
