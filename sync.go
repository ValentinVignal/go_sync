package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type SyncedProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type SyncedData struct {
	Projects []SyncedProject `json:"projects"`
}

func Sync(c *gin.Context) {
	log.Println("Sync")

	connString := "postgres://root:password@localhost:5555/go_sync?sslmode=disable"
	database, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	rows, err := database.Query("SELECT id, name, code FROM project")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	projects := []SyncedProject{}
	for rows.Next() {
		var (
			id   string
			name string
			code string
		)

		if err := rows.Scan(&id, &name, &code); err != nil {
			log.Fatal(err)
		} else {
			projects = append(projects, SyncedProject{
				ID:   id,
				Name: name,
				Code: code,
			})
		}
	}

	syncedData := SyncedData{
		Projects: projects,
	}

	c.JSON(http.StatusOK, syncedData)
}
