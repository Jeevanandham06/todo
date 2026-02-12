package updel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"todo/csvf"
)

type DeleteReq struct {
	TaskName string `json:"taskname"`
}

type DeleteResponse struct {
	Status string `json:"status"`
	Code   string `json:"code,omitempty"`
	Msg    string `json:"msg"`
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	// Check PUT method (soft delete via status change)
	if r.Method != http.MethodPut {
		resp := DeleteResponse{Status: "error", Code: "TUD001", Msg: "Use PUT method"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resp := DeleteResponse{Status: "error", Code: "TUD002", Msg: "Error reading body"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	var req DeleteReq
	err = json.Unmarshal(body, &req)
	if err != nil {
		resp := DeleteResponse{Status: "error", Code: "TUD003", Msg: "Invalid JSON format"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Check taskname
	if req.TaskName == "" {
		resp := DeleteResponse{Status: "error", Code: "TUD004", Msg: "taskname required"}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	// Delete from CSV
	err = csvf.DeleteTaskByStatusChange(req.TaskName)
	if err != nil {
		resp := DeleteResponse{Status: "error", Code: "TUD005", Msg: err.Error()}
		jsonData, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonData))
		return
	}

	resp := DeleteResponse{Status: "success", Msg: "Task deleted successfully"}
	jsonData, _ := json.Marshal(resp)
	fmt.Fprint(w, string(jsonData))
}
