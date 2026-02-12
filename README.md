================================================================================
TODO PROJECT - COMPLETE CODE
================================================================================

PROJECT STRUCTURE:
- go.mod
- main.go
- tasks.csv
- addtaa/Addtask.go
- csvf/simple.go
- show/Showtask.go
- updel/Updatetask.go
- updel/Deletetask.go

================================================================================
1. go.mod
================================================================================
module todo

go 1.25.7

================================================================================
2. main.go
================================================================================
package main

import (
    "log"
    "net/http"
    "todo/addtaa"
    "todo/show"
    "todo/updel"
)

func main() {
    // Routes
    http.HandleFunc("/add", addtaa.AddTask)
    http.HandleFunc("/show", show.ShowTask)
    http.HandleFunc("/upd", updel.UpdateTask)
    http.HandleFunc("/del", updel.DeleteTask) // Now uses PUT method

    // Start server
    port := ":8082"
    log.Printf("Server running at http://localhost%s\n", port)
    log.Println("POST /add   - Add task (unique taskname required)")
    log.Println("GET /show   - Show all active tasks")
    log.Println("PUT /upd    - Update task details")
    log.Println("PUT /del    - Delete task (soft delete, status=N)")
    log.Println("Time format: HH:MM:SS (synced with local timezone)")
    log.Fatal(http.ListenAndServe(port, nil))
}

================================================================================
3. addtaa/Addtask.go
================================================================================
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

================================================================================
4. csvf/simple.go
================================================================================
package csvf

import (
    "encoding/csv"
    "errors"
    "os"
    "time"
)

// SimpleTask is a basic task struct
type SimpleTask struct {
    TaskName string `json:"taskname"`
    Desc     string `json:"desc"`
    Time     string `json:"time"`
    Status   string `json:"status"`
}

// TaskExists checks if task name already exists
func TaskExists(taskname string) bool {
    file, err := os.Open("tasks.csv")
    if err != nil {
        return false
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, _ := reader.ReadAll()

    for i, row := range records {
        if i == 0 {
            continue // Skip header
        }
        if len(row) < 4 {
            continue
        }
        if row[0] == taskname && row[3] != "N" {
            return true
        }
    }
    return false
}

// SaveTask saves a task to CSV
func SaveTask(taskname, desc, tasktime string) error {
    // Check if task already exists
    if TaskExists(taskname) {
        return errors.New("[TCS003] task name already exists")
    }

    file, err := os.OpenFile("tasks.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return errors.New("[TCS001] failed to open CSV file")
    }
    defer file.Close()

    stat, err := file.Stat()
    if err != nil {
        return errors.New("[TCS002] failed to read file stats")
    }
    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Write header if empty
    if stat.Size() == 0 {
        writer.Write([]string{"TaskName", "Description", "Time", "Status"})
    }

    // Write task
    writer.Write([]string{taskname, desc, tasktime, "Y"})
    return nil
}

// GetAllTasks returns all tasks from CSV as simple structs
func GetAllTasks() ([]SimpleTask, error) {
    file, err := os.Open("tasks.csv")
    if err != nil {
        return []SimpleTask{}, errors.New("[TCG001] failed to open tasks file")
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return []SimpleTask{}, errors.New("[TCG002] failed to read CSV")
    }

    var tasks []SimpleTask
    for i, row := range records {
        if i == 0 {
            continue // Skip header
        }
        if len(row) < 4 {
            continue
        }

        // Filter out deleted tasks
        if row[3] == "N" {
            continue
        }

        task := SimpleTask{
            TaskName: row[0],
            Desc:     row[1],
            Time:     row[2],
            Status:   row[3],
        }
        tasks = append(tasks, task)
    }
    return tasks, nil
}

// UpdateTaskInCSV updates a task
func UpdateTaskInCSV(taskname, desc, newtime string) error {
    tasks, err := getAllTasks()
    if err != nil {
        return errors.New("[TCU001] failed to read tasks")
    }

    found := false
    // Find and update
    for i := range tasks {
        if tasks[i].TaskName == taskname {
            tasks[i].Desc = desc
            tasks[i].Time = newtime
            found = true
        }
    }

    if !found {
        return errors.New("[TCU002] task not found")
    }

    // Write all back
    return writeAllTasks(tasks)
}

// DeleteTaskInCSV marks task as deleted
func DeleteTaskInCSV(taskname string) error {
    tasks, err := getAllTasks()
    if err != nil {
        return errors.New("[TCD001] failed to read tasks")
    }

    found := false
    for i := range tasks {
        if tasks[i].TaskName == taskname {
            tasks[i].Status = "N"
            found = true
        }
    }

    if !found {
        return errors.New("[TCD002] task not found")
    }

    return writeAllTasks(tasks)
}

// DeleteTaskByStatusChange updates status to N (soft delete)
func DeleteTaskByStatusChange(taskname string) error {
    all, err := getAllTasksIncludeDeleted()
    if err != nil {
        return errors.New("[TCD001] failed to read tasks")
    }

    found := false
    for i := range all {
        if all[i].TaskName == taskname {
            all[i].Status = "N"
            found = true
        }
    }

    if !found {
        return errors.New("[TCD002] task not found")
    }

    return writeAllTasksIncludeDeleted(all)
}

// SimpleTaskStatus returns simple status
func SimpleTaskStatus(tasktime string) string {
    parsedTime, _ := time.ParseInLocation("2006-01-02 15:04:05", tasktime, time.Local)

    if time.Now().After(parsedTime) {
        return "expired"
    }
    if time.Until(parsedTime) < 15*time.Minute {
        return "active"
    }
    return "pending"
}

// Internal helpers
func getAllTasks() ([]SimpleTask, error) {
    file, err := os.Open("tasks.csv")
    if err != nil {
        return []SimpleTask{}, errors.New("[TCG003] failed to open tasks file")
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return []SimpleTask{}, errors.New("[TCG004] failed to parse CSV")
    }

    var tasks []SimpleTask
    for i, row := range records {
        if i == 0 {
            continue
        }
        if len(row) < 4 {
            continue
        }

        task := SimpleTask{
            TaskName: row[0],
            Desc:     row[1],
            Time:     row[2],
            Status:   row[3],
        }
        tasks = append(tasks, task)
    }
    return tasks, nil
}

// getAllTasksIncludeDeleted reads all tasks including deleted ones
func getAllTasksIncludeDeleted() ([]SimpleTask, error) {
    file, err := os.Open("tasks.csv")
    if err != nil {
        return []SimpleTask{}, errors.New("[TCG005] failed to open tasks file")
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return []SimpleTask{}, errors.New("[TCG006] failed to parse CSV")
    }

    var tasks []SimpleTask
    for i, row := range records {
        if i == 0 {
            continue
        }
        if len(row) < 4 {
            continue
        }

        task := SimpleTask{
            TaskName: row[0],
            Desc:     row[1],
            Time:     row[2],
            Status:   row[3],
        }
        tasks = append(tasks, task)
    }
    return tasks, nil
}

func writeAllTasks(tasks []SimpleTask) error {
    file, err := os.Create("tasks.csv")
    if err != nil {
        return errors.New("[TCW001] failed to create tasks file")
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Write header
    writer.Write([]string{"TaskName", "Description", "Time", "Status"})

    // Write only active tasks
    for _, task := range tasks {
        if task.Status != "N" {
            writer.Write([]string{task.TaskName, task.Desc, task.Time, task.Status})
        }
    }
    return nil
}

// writeAllTasksIncludeDeleted writes all tasks including deleted ones
func writeAllTasksIncludeDeleted(tasks []SimpleTask) error {
    file, err := os.Create("tasks.csv")
    if err != nil {
        return errors.New("[TCW002] failed to create tasks file")
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Write header
    writer.Write([]string{"TaskName", "Description", "Time", "Status"})

    // Write all tasks including deleted
    for _, task := range tasks {
        writer.Write([]string{task.TaskName, task.Desc, task.Time, task.Status})
    }
    return nil
}

================================================================================
5. show/Showtask.go
================================================================================
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

================================================================================
6. updel/Updatetask.go
================================================================================
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

================================================================================
7. updel/Deletetask.go
================================================================================
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

================================================================================
END OF PROJECT CODE
================================================================================
