package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"vigilate/internal/models"

	"github.com/go-chi/chi"
)

// Types of services
const (
	HTTP           = 1
	HTTPS          = 2
	SSLCertificate = 3
)

// JSON resp sent to client
type jsonResp struct {
	OK            bool      `json:"ok"`
	Message       string    `json:"message"`
	ServiceID     int       `json:"service_id"`
	HostServiceID int       `json:"host_service_id"`
	HostID        int       `json:"host_id"`
	OldStatus     string    `json:"old_status"`
	NewStatus     string    `json:"new_status"`
	LastCheck     time.Time `json:"last_check"`
}

// ScheduledCheck performs a scheduled check on a host service by id
func (repo *DBRepo) ScheduledCheck(hostServiceID int) {
	log.Println("**************** Running check for", hostServiceID)

	hs, err := repo.DB.GetHostServiceByID(hostServiceID)
	if err != nil {
		log.Println(err)
		return
	}
	h, err := repo.DB.GetHostByID(hs.HostID)
	if err != nil {
		log.Println(err)
		return
	}

	newStatus, msg := repo.testServiceForHost(h, hs)

	if newStatus != hs.Status {
		data := make(map[string]string)
		data["message"] = fmt.Sprintf("host services %s on %s has changed to %s", hs.Service.ServiceName, h.HostName, newStatus)
		repo.broadcastMessage("public-channel", "host-service-status-changed", data)

	}

	hs.Status = newStatus
	hs.LastCheck = time.Now()
	err = repo.DB.UpdateHostService(hs)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("New status is", newStatus, "and msg is ", msg)

}

func (repop *DBRepo) broadcastMessage(channel, messageType string, data map[string]string) {
	err := app.WsClient.Trigger(channel, messageType, data)
	if err != nil {
		log.Println(err)
	}
}

// TestCheck does manual service checks via HTTP endpoint
func (repo *DBRepo) TestCheck(w http.ResponseWriter, r *http.Request) {
	//Extract host service record and old status from URL
	hostServiceID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	oldStatus := chi.URLParam(r, "oldStatus")
	okay := true
	log.Println(hostServiceID, oldStatus)

	//Fetch host_services record from DB
	hs, err := repo.DB.GetHostServiceByID(hostServiceID)
	if err != nil {
		log.Println(err)
		okay = false
	}
	log.Println("Service name is", hs.Service.ServiceName)

	//Fetch the host record for this service
	h, err := repo.DB.GetHostByID(hs.HostID)
	if err != nil {
		log.Println(err)
		okay = false

	}

	//Run the actual service test based on service type
	msg, newStatus := repo.testServiceForHost(h, hs)

	//update the host service in the database with status (if changed) and last check
	hs.Status = newStatus
	hs.LastCheck = time.Now()
	hs.UpdatedAt = time.Now()
	err = repo.DB.UpdateHostService(hs)
	if err != nil {
		log.Println(err)
		okay = false
	}
	//broadcast service status changed event

	var resp jsonResp

	//send json to client
	if okay {
		resp = jsonResp{
			OK:            true,
			Message:       msg,
			ServiceID:     hs.ServiceID,
			HostServiceID: hs.ID,
			HostID:        hs.HostID,
			OldStatus:     oldStatus,
			NewStatus:     newStatus,
			LastCheck:     time.Now(),
		}
	} else {
		resp.OK = false
		resp.Message = "Something went wrong!"
	}

	// Marshal JSON and sent to client
	out, _ := json.MarshalIndent(resp, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// testServiceForHost determines which test to run depending on service type
func (repo *DBRepo) testServiceForHost(h models.Host, hs models.HostService) (string, string) {
	var msg, newStatus string

	switch hs.ServiceID {
	case HTTP:
		//Run HTTP test for this host
		msg, newStatus = testHTTPForHost(h.URL)

	}
	return msg, newStatus
}

// testHTTPForHost performs actual HTTP GET request to check if host is reachable
func testHTTPForHost(url string) (string, string) {
	//Normalize URL : remove trailing slash
	if strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}

	//Convert https to http
	url = strings.Replace(url, "https://", "http://", -1)

	//Make GET request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("%s - %s", url, "error connecting"), "problem"
	}
	defer resp.Body.Close()

	//Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("%s - %s", url, resp.Status), "problem"
	}

	//Status 200 = healthy
	return fmt.Sprintf("%s - %s", url, resp.Status), "healthy"
}
