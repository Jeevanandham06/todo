package show

import (
	"encoding/json"
	"fmt"
	"net/http"
	"todo/csvf"
)

type ShowResponse struct {
	Status string            `json:"status"`
	Code   string            `json:"code,omitempty"`
	Msg    string            `json:"msg"`
	Tasks  []csvf.SimpleTask `json:"tasks"`
}

func ShowTask(w http.ResponseWriter, r *http.Request) {
	// Check if GET
	if r.Method != http.MethodGet {
		resp := ShowResponse{Status: "error", Code: "TSS001", Msg: "Use GET method", Tasks: []csvf.SimpleTask{}}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Get all tasks from CSV
	tasks, err := csvf.GetAllTasks()
	if err != nil {
		resp := ShowResponse{Status: "error", Code: "TSS002", Msg: err.Error(), Tasks: []csvf.SimpleTask{}}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Update status for each task
	for i := range tasks {
		tasks[i].Status = csvf.SimpleTaskStatus(tasks[i].Time)
	}

	// Return as JSON
	response := ShowResponse{
		Status: "success",
		Msg:    "",
		Tasks:  tasks,
	}
	jsonData, _ := json.Marshal(response)
	fmt.Fprint(w, string(jsonData))
}
