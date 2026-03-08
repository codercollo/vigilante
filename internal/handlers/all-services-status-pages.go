package handlers

import (
	"log"
	"net/http"
	"vigilate/internal/helpers"

	"github.com/CloudyKit/jet/v6"
)

//Package handlers contains HTTP handler functions for various routes

// AllHealthyServices renders healthy services page
func (repo *DBRepo) AllHealthyServices(w http.ResponseWriter, r *http.Request) {
	//get all host services (with host info) for status pending
	services, err := repo.DB.GetServicesByStatus("healthy")
	if err != nil {
		log.Println(err)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("services", services)

	err = helpers.RenderPage(w, r, "healthy", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllWarningServices renders warning services page
func (repo *DBRepo) AllWarningServices(w http.ResponseWriter, r *http.Request) {
	//get all host services (with host info) for status pending
	services, err := repo.DB.GetServicesByStatus("warning")
	if err != nil {
		log.Println(err)
		return
	}
	vars := make(jet.VarMap)
	vars.Set("services", services)
	err = helpers.RenderPage(w, r, "warning", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllProblemsServices renders problemn services page
func (repo *DBRepo) AllProblemServices(w http.ResponseWriter, r *http.Request) {
	//get all host services (with host info) for status pending
	services, err := repo.DB.GetServicesByStatus("problem")
	if err != nil {
		log.Println(err)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("services", services)
	err = helpers.RenderPage(w, r, "problems", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllPendingServices renders pending services page
func (repo *DBRepo) AllPendingServices(w http.ResponseWriter, r *http.Request) {
	//get all host services (with host info) for status pending
	services, err := repo.DB.GetServicesByStatus("pending")
	if err != nil {
		log.Println(err)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("services", services)

	err = helpers.RenderPage(w, r, "pending", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}
