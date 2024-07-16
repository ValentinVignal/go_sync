package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type SyncedProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type SyncedForm struct {
	ID        string                 `json:"id"`
	ProjectID string                 `json:"projectID"`
	Name      string                 `json:"name"`
	Data      map[string]interface{} `json:"data"`
	TaskIDs   []string               `json:"taskIDs"`
}

type SyncedTask struct {
	ID          string   `json:"id"`
	ProjectID   string   `json:"projectID"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	FormIDs     []string `json:"formIDs"`
}

type SyncedData struct {
	Projects []SyncedProject `json:"projects"`
	Forms    []SyncedForm    `json:"forms"`
	Tasks    []SyncedTask    `json:"tasks"`
}

func LocalSync(c *gin.Context) {
	log.Println("Local Sync")

	connString := "postgres://root:password@localhost:5555/go_sync?sslmode=disable"
	database, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	projectRows, err := database.Query(`SELECT id, name, code FROM project  WHERE "deletedAt" IS NULL ORDER BY "updatedAt" DESC LIMIT 1000`)
	if err != nil {
		log.Fatal(err)
	}
	defer projectRows.Close()

	projects := []SyncedProject{}
	for projectRows.Next() {
		var (
			id   string
			name string
			code string
		)

		if err := projectRows.Scan(&id, &name, &code); err != nil {
			log.Fatal(err)
		} else {
			projects = append(projects, SyncedProject{
				ID:   id,
				Name: name,
				Code: code,
			})
		}
	}

	formRows, err := database.Query(`SELECT form.id, form.name, form."projectID" as project_id, form.data, array_remove(array_agg(forms_tasks."taskID"), NULL) as task_ids FROM form
        LEFT JOIN forms_tasks ON form.id = forms_tasks."formID"
        WHERE form."deletedAt" IS NULL
        GROUP BY form.id
        ORDER BY form."updatedAt" DESC LIMIT 1000;`)
	if err != nil {
		log.Fatal(err)
	}
	defer formRows.Close()

	forms := []SyncedForm{}
	for formRows.Next() {
		var (
			id        string
			name      string
			projectID string
			data      []byte
			taskIDs   []string
		)

		if err := formRows.Scan(&id, &name, &projectID, &data, pq.Array(&taskIDs)); err != nil {
			log.Fatal(err)
		} else {
			unMarshaledData := map[string]interface{}{}
			json.Unmarshal(data, &unMarshaledData)
			forms = append(forms, SyncedForm{
				ID:        id,
				Name:      name,
				ProjectID: projectID,
				Data:      unMarshaledData,
				TaskIDs:   taskIDs,
			})
		}
	}

	taskRows, err := database.Query(`SELECT task.id, task.name, task.description, task."projectID" as project_id, array_remove(array_agg(forms_tasks."formID"), NULL) as form_ids FROM task
        LEFT JOIN forms_tasks ON task.id = forms_tasks."taskID"
        WHERE task."deletedAt" IS NULL
        GROUP BY task.id
        ORDER BY task."updatedAt" DESC LIMIT 1000;`)
	if err != nil {
		log.Fatal(err)
	}
	defer taskRows.Close()

	tasks := []SyncedTask{}
	for taskRows.Next() {
		var (
			id          string
			name        string
			projectID   string
			description string
			formIDs     []string
		)

		if err := taskRows.Scan(&id, &name, &description, &projectID, pq.Array(&formIDs)); err != nil {
			log.Fatal(err)
		} else {
			tasks = append(tasks, SyncedTask{
				ID:          id,
				Name:        name,
				ProjectID:   projectID,
				Description: description,
				FormIDs:     formIDs,
			})
		}
	}

	syncedData := SyncedData{
		Projects: projects,
		Forms:    forms,
		Tasks:    tasks,
	}

	c.JSON(http.StatusOK, syncedData)
}
