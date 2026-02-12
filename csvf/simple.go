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
