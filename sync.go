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

type SyncedData struct {
	Projects []SyncedProject `json:"projects"`
	Forms    []SyncedForm    `json:"forms"`
}

func Sync(c *gin.Context) {
	log.Println("Sync")

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

	formsRows, err := database.Query(`SELECT form.id, form.name, form."projectID" as project_id, form.data, array_remove(array_agg(forms_tasks."taskID"), NULL) as task_ids FROM form
        LEFT JOIN forms_tasks ON form.id = forms_tasks."formID"
        WHERE form."deletedAt" IS NULL
        GROUP BY form.id
        ORDER BY form."updatedAt" DESC LIMIT 1000;`)
	if err != nil {
		log.Fatal(err)
	}
	defer formsRows.Close()

	forms := []SyncedForm{}
	for formsRows.Next() {
		var (
			id        string
			name      string
			projectID string
			data      []byte
			taskIDs   []string
		)

		if err := formsRows.Scan(&id, &name, &projectID, &data, pq.Array(&taskIDs)); err != nil {
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

	syncedData := SyncedData{
		Projects: projects,
		Forms:    forms,
	}

	c.JSON(http.StatusOK, syncedData)
}
