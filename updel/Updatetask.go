package updel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"todo/csvf"
)

type UpdateReq struct {
	TaskName string `json:"taskname"`
	Desc     string `json:"desc"`
	Time     string `json:"time"`
}

type APIResponse struct {
	Status string `json:"status"`
	Code   string `json:"code,omitempty"`
	Msg    string `json:"msg"`
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	// Check if PUT
	if r.Method != http.MethodPut {
		resp := APIResponse{Status: "error", Code: "TUU001", Msg: "Use PUT method"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resp := APIResponse{Status: "error", Code: "TUU002", Msg: "Error reading body"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Parse JSON
	var req UpdateReq
	err = json.Unmarshal(body, &req)
	if err != nil {
		resp := APIResponse{Status: "error", Code: "TUU003", Msg: "Invalid JSON format"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Check inputs
	if req.TaskName == "" {
		resp := APIResponse{Status: "error", Code: "TUU004", Msg: "taskname required"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}
	if req.Desc == "" {
		resp := APIResponse{Status: "error", Code: "TUU005", Msg: "desc required"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}
	if req.Time == "" {
		resp := APIResponse{Status: "error", Code: "TUU006", Msg: "time required (HH:MM:SS)"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Parse time
	parsedTime, err := time.ParseInLocation("15:04:05", req.Time, time.Local)
	if err != nil {
		resp := APIResponse{Status: "error", Code: "TUU007", Msg: "Invalid time, use HH:MM:SS"}
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
		resp := APIResponse{Status: "error", Code: "TUU008", Msg: "Time must be 10 mins or more"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Update in CSV
	err = csvf.UpdateTaskInCSV(req.TaskName, req.Desc, fullDateTime.Format("2006-01-02 15:04:05"))
	if err != nil {
		resp := APIResponse{Status: "error", Code: "TUU009", Msg: err.Error()}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	resp := APIResponse{Status: "success", Msg: "Task updated"}
	jsonData, _ := json.Marshal(resp)
	fmt.Fprint(w, string(jsonData))
}
