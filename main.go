package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	YYMMDD = "2006-01-02 15:04:05"
)

type Configuration struct {
	AirflowURL       string `json:"airflow_url"`
	Interval         uint   `json:"interval"`
	DiscordThreadUrl string `json:"discord_thread_url"`
	DiscordName      string `json:"discord_name"`
	DiscordAvatarUrl string `json:"discord_avatar_url"`
	// UpMsg      string `json:"up_message"`
	// DownMsg    string `json:"down_message"`
}

type Metadatabase struct {
	Status string `json:"status"`
}

type Scheduler struct {
	LatestSchedulerHeartbeat string `json:"latest_scheduler_heartbeat"`
	Status                   string `json:"status"`
}

type AirflowResponse struct {
	Metadatabase Metadatabase `json:"metadatabase"`
	Scheduler    Scheduler    `json:"scheduler"`
}

type AirflowStatus struct {
	WebserverStatus       string `json:"webserver_status"`
	WebserverDownTime     string `json:"webserver_downtime"`
	WebserverRecovered    bool   `json:"webserver_recovered"`
	MetadatabaseStatus    string `json:"metadatabase_status"`
	MetadatabaseDownTime  string `json:"metadatabase_downtime"`
	MetadatabaseRecovered bool   `json:"metadatabase_recovered"`
	SchedulerStatus       string `json:"scheduler_status"`
	SchedulerDownTime     string `json:"scheduler_downtime"`
	SchedulerRecovered    bool   `json:"scheduler_recovered"`
}

type DiscordMessage struct {
	Content   string `json:"content"`
	Username  string `json:"username"`
	AvatarUrl string `json:"avatar_url"`
}

type DiscordResponse struct {
	Code    uint   `json:"code"`
	Message string `json:"message"`
}

func getAirflowStatus(cfg *Configuration) (error, AirflowResponse) {
	var airflowResponse AirflowResponse

	airflowHealthUrl := cfg.AirflowURL + "/health"

	res, err := http.Get(airflowHealthUrl)
	if err != nil {
		return fmt.Errorf("Cannot making request: %v", err), airflowResponse
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&airflowResponse)
	if err != nil {
		return fmt.Errorf("Cannot parsing response: %v", err), airflowResponse
	}

	return nil, airflowResponse
}

func saveAirfloStatus(obj interface{}) {
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}

	file, _ := os.Create("airflow_status.json")
	defer file.Close()

	_, err = file.Write(b)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func sendNotificationDiscord(cfg *Configuration, msg string) {
	discordMessage := DiscordMessage{
		Username:  cfg.DiscordName,
		AvatarUrl: cfg.DiscordAvatarUrl,
		Content:   msg,
	}
	messageJSON, err := json.Marshal(discordMessage)
	if err != nil {
		fmt.Println("error: ", err)
	}

	_, err = http.Post(cfg.DiscordThreadUrl, "application/json", bytes.NewBuffer(messageJSON))
	if err != nil {
		fmt.Println("error: ", err)
		return
	}

	fmt.Println("info: Successfully send alert to Discord")

	// defer resp.Body.Close()

	// discordResponse := DiscordResponse{}
	// err = json.NewDecoder(resp.Body).Decode(&discordResponse)
	// if err != nil {
	// 	fmt.Printf("error: Cannot parsing response %v\n", err)
	// }
}

func minutesDifference(timeFrom string, timeNow string) string {
	parsedTimeFrom, err := time.Parse(YYMMDD, timeFrom)
	if err != nil {
		fmt.Printf("error: parsing time from: %v\n", err)
	}

	parsedTimeNow, err := time.Parse(YYMMDD, timeNow)
	if err != nil {
		fmt.Printf("error: parsing time now: %v\n", err)
	}

	diff := parsedTimeNow.Sub(parsedTimeFrom).Minutes()
	return fmt.Sprintf("%d minutes", uint(diff))
}

func listen(cfg *Configuration) {
	var airflowStatus AirflowStatus
	var airflowResponse AirflowResponse
	var err error

	file, _ := os.Open("airflow_status.json")
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&airflowStatus)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	interval := cfg.Interval
	fmt.Printf("info: Scheduling task to run every %d seconds\n", interval)
	for range time.Tick(time.Duration(interval) * time.Second) {
		fmt.Println("info: Executing health check")

		timeNow := time.Now().Format(YYMMDD)

		err, airflowResponse = getAirflowStatus(cfg)
		if err != nil {
			fmt.Println("error: " + err.Error())
			airflowStatus.WebserverStatus = "unhealthy"
			if airflowStatus.WebserverRecovered {
				airflowStatus.WebserverDownTime = timeNow
				fmt.Println("info: Webserver is unhealthy, send alert to Discord ...")
				// send notif down
				sendNotificationDiscord(cfg, "Monitor is DOWN: airflow-webserver. Hi @here, please check service status")
			}
			airflowStatus.WebserverRecovered = false
			saveAirfloStatus(airflowStatus)
			continue
		}

		if airflowStatus.WebserverStatus != "healthy" {
			fmt.Println("Info: Webserver health recovered, send alert to Discord ...")
			// calculate time diff in minutes
			minutesDiff := minutesDifference(airflowStatus.WebserverDownTime, timeNow)
			// send notif down
			sendNotificationDiscord(cfg, "Monitor is UP: airflow-webserver. Hi @here, it was down for "+minutesDiff)
			airflowStatus.WebserverStatus = "healthy"
			airflowStatus.WebserverRecovered = true
			saveAirfloStatus(airflowStatus)
		}

		airflowStatus.MetadatabaseStatus = "" + airflowResponse.Metadatabase.Status
		airflowStatus.SchedulerStatus = "" + airflowResponse.Scheduler.Status

		if airflowStatus.MetadatabaseStatus != "healthy" {
			fmt.Println("info: Scheduler is unhealthy")
			if airflowStatus.MetadatabaseRecovered {
				airflowStatus.MetadatabaseDownTime = timeNow
				fmt.Println("info: Scheduler is unhealthy, send alert to Discord ...")
				// send notif down
				sendNotificationDiscord(cfg, "Monitor is DOWN: airflow-metadatabase. Hi @here, please check service status")
			}
			airflowStatus.MetadatabaseRecovered = false
			saveAirfloStatus(airflowStatus)
		}
		if airflowStatus.SchedulerStatus != "healthy" {
			fmt.Println("info: Metadatabase is unhealthy")
			if airflowStatus.SchedulerRecovered {
				airflowStatus.SchedulerDownTime = timeNow
				fmt.Println("info: Metadatabase is unhealthy, send alert to Discord ...")
				// send notif down
				sendNotificationDiscord(cfg, "Monitor is DOWN: airflow-scheduler. Hi @here, please check service status")
			}
			airflowStatus.SchedulerRecovered = false
			saveAirfloStatus(airflowStatus)
		}
		if airflowStatus.MetadatabaseStatus == "healthy" && !airflowStatus.MetadatabaseRecovered {
			fmt.Println("Info: Metadatabase health recovered, send alert to Discord ...")
			airflowStatus.MetadatabaseRecovered = true
			saveAirfloStatus(airflowStatus)
			// calculate time diff in minutes
			minutesDiff := minutesDifference(airflowStatus.MetadatabaseDownTime, timeNow)
			// send notif up
			sendNotificationDiscord(cfg, "Monitor is UP: airflow-metadatabase. Hi @here, it was down for "+minutesDiff)
		}
		if airflowStatus.SchedulerStatus == "healthy" && !airflowStatus.SchedulerRecovered {
			fmt.Println("Info: Scheduler health recovered, send alert to Discord ...")
			airflowStatus.SchedulerRecovered = true
			saveAirfloStatus(airflowStatus)
			// calculate time diff in minutes
			minutesDiff := minutesDifference(airflowStatus.SchedulerDownTime, timeNow)
			// send notif down
			sendNotificationDiscord(cfg, "Monitor is UP: airflow-scheduler. Hi @here, it was down for "+minutesDiff)
		}
	}
}

func main() {
	file, _ := os.Open("config.json")
	defer file.Close()

	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}

	listen(&configuration)
}
