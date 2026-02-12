package addtaa

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"todo/csvf"
)

type Task struct {
	TaskName string `json:"taskname"`
	Desc     string `json:"desc"`
	Time     string `json:"time"`
}

type APIResponse struct {
	Status string `json:"status"`
	Code   string `json:"code,omitempty"`
	Msg    string `json:"msg"`
}

func AddTask(w http.ResponseWriter, r *http.Request) {
	// Check if POST
	if r.Method != http.MethodPost {
		resp := APIResponse{Status: "error", Code: "TAA001", Msg: "Use POST method"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resp := APIResponse{Status: "error", Code: "TAA002", Msg: "Error reading body"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Parse JSON
	var task Task
	err = json.Unmarshal(body, &task)
	if err != nil {
		resp := APIResponse{Status: "error", Code: "TAA003", Msg: "Invalid JSON format"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Check taskname
	if task.TaskName == "" {
		resp := APIResponse{Status: "error", Code: "TAA004", Msg: "taskname required"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Check desc
	if task.Desc == "" {
		resp := APIResponse{Status: "error", Code: "TAA005", Msg: "desc required"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Check time
	if task.Time == "" {
		resp := APIResponse{Status: "error", Code: "TAA006", Msg: "time required (HH:MM:SS)"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Parse time
	parsedTime, err := time.ParseInLocation("15:04:05", task.Time, time.Local)
	if err != nil {
		resp := APIResponse{Status: "error", Code: "TAA007", Msg: "Invalid time, use HH:MM:SS like 14:30:00"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Make full datetime
	now := time.Now()
	fullDateTime := time.Date(now.Year(), now.Month(), now.Day(),
		parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), 0, time.Local)

	// If past, use tomorrow
	if now.After(fullDateTime) {
		fullDateTime = fullDateTime.AddDate(0, 0, 1)
	}

	// Check 10 minute limit
	if time.Until(fullDateTime) < 10*time.Minute {
		resp := APIResponse{Status: "error", Code: "TAA008", Msg: "Time must be 10 mins or more from now"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Save
	err = csvf.SaveTask(task.TaskName, task.Desc, fullDateTime.Format("2006-01-02 15:04:05"))
	if err != nil {
		resp := APIResponse{Status: "error", Code: "TAA009", Msg: err.Error()}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	resp := APIResponse{Status: "success", Msg: "Task added successfully"}
	jsonData, _ := json.Marshal(resp)
	fmt.Fprint(w, string(jsonData))
}
