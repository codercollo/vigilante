package handlers

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/pusher/pusher-http-go"
)

// PusherAuth handles authentication for Pusher presence channels
func (repo *DBRepo) PusherAuth(w http.ResponseWriter, r *http.Request) {
	//Get the current user's ID from session
	userID := repo.App.Session.GetInt(r.Context(), "userID")

	//Retrieve full user details from the database
	u, _ := repo.DB.GetUserById(userID)

	// Read request body containing Pusher auth params
	params, _ := ioutil.ReadAll(r.Body)

	// Prepare presence data with user info
	presenceData := pusher.MemberData{
		UserID: strconv.Itoa(userID),
		UserInfo: map[string]string{
			"name": u.FirstName,
			"id":   strconv.Itoa(userID),
		},
	}

	// Authenticate the user for the presence channel
	response, err := app.WsClient.AuthenticatePresenceChannel(params, presenceData)
	if err != nil {
		log.Println(err)
		return
	}

	//Return the authentication response as JSON
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(response)
}
