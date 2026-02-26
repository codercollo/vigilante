package handlers

import (
	"net/http"
	"vigilate/internal/helpers"
)

//Package handlers provides HTTP handlers for the admin panel and related features

// ListEntries renders the schedule page showing all schedule entries
func (repo *DBRepo) ListEntries(w http.ResponseWriter, r *http.Request) {
	err := helpers.RenderPage(w, r, "schedule", nil, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}
