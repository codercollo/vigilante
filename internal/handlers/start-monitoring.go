package handlers

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

// job represents a monitoring task for a specific service
type job struct {
	HostServiceID int //ID of service we want to monitor
}

// Run is executed by the scheduler whenever the job is trggered
func (j job) Run() {
	Repo.ScheduledCheck(j.HostServiceID)
}

func (repo *DBRepo) StartMonitoring() {

	//Only start monitoring if the configuration allows it
	if app.PreferenceMap["monitoring_live"] == "1" {
		log.Println("***********starting monitor")

		//Prepare websocket message data
		data := make(map[string]string)
		data["message"] = "Monitoring is starting"

		//Notify all connected clients that monitoring has started
		err := app.WsClient.Trigger("public-channel", "app-starting", data)
		if err != nil {
			log.Println(err)
		}

		//Get all the services that should be monitored from the database
		servicesToMonitor, err := repo.DB.GetServicesToMonitor()
		if err != nil {
			log.Println(err)
		}

		//Loop through each service to chedule monitoring jobs
		for _, x := range servicesToMonitor {

			//Build the scheduler interval string ie: @every 5s
			var sch string

			if x.ScheduleUnit == "d" {
				//Scheduler doesn't support days directly so convert days to hrs
				sch = fmt.Sprintf("@every %s%s", x.ScheduleNumber*24, "h")
			} else {
				sch = fmt.Sprintf("@every %s%s", x.ScheduleNumber, x.ScheduleUnit)

			}

			//Create a job for this specific service
			var j job
			j.HostServiceID = x.ID

			//Register the job with the scheduler
			scheduleID, err := app.Scheduler.AddJob(sch, j)
			if err != nil {
				log.Println(err)
			}

			//Store the scheduler job ID so we can stop or modify it later
			app.MonitorMap[x.ID] = scheduleID

			// Prepare websocket payload describing the scheduled job
			payload := make(map[string]string)
			payload["message"] = "scheduling"
			payload["host_service_id"] = strconv.Itoa(x.ID)

			// This is a "zero-like" reference time used to check if a time is valid
			yearone := time.Date(0001, 11, 17, 20, 34, 58, 65138737, time.UTC)

			//If scheduler already calculated the next run time
			if app.Scheduler.Entry(app.MonitorMap[x.ID]).Next.After(yearone) {
				//Send formatted next run time
				data["next_run"] = app.Scheduler.Entry(app.MonitorMap[x.ID]).Next.Format("2006-01-02 3:04:05 PM")
			} else {
				//Scheduler has not scheduled the run yet
				data["next_run"] = "Pending..."
			}

			//Add more information about the service
			payload["host"] = x.HostName
			payload["service"] = x.Service.ServiceName

			//If the service has been checked before
			if x.LastCheck.After(yearone) {
				//Send the last run time
				payload["last_run"] = x.LastCheck.Format("2006-01-02 3:04:05 PM")
			} else {
				//No previous check
				payload["last_run"] = "Pending..."
			}
			//Send the schedule interval used
			payload["schedule"] = fmt.Sprintf("@every %d%s", x.ScheduleNumber, x.ScheduleUnit)

			//Notify clients about the next sheduled run
			err = app.WsClient.Trigger("puclic-channel", "next-run-event", payload)
			if err != nil {
				log.Println(err)
			}

			//Notify clients that the schedule has changed
			err = app.WsClient.Trigger("public-channel", "schedule-changed-event", payload)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
