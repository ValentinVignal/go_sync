package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Channel struct {
	Name    string     `json:"name"`
	Columns []string   `json:"columns"`
	Rows    [][]string `json:"rows"`
}
type SyncResponse struct {
	Channels []Channel `json:"channels"`
}

func Test(c *gin.Context) {
	log.Println("Sync")

	connString := "postgres://postgres:password@localhost:5432/novadelite?sslmode=disable"
	database, err := sql.Open("postgres", connString)
	if err != nil {
		log.Panic(err)
	}
	defer database.Close()

	response := SyncResponse{
		Channels: []Channel{},
	}

	projectChannel := Channel{
		Name:    "projects",
		Columns: []string{"id", "name", "code"},
		Rows:    [][]string{},
	}

	projectRows, err := database.Query(`SELECT id, name, code FROM projects WHERE "deletedAt" IS NULL ORDER BY "updatedAt" DESC LIMIT 1000`)
	if err != nil {
		log.Panic(err)
	}
	defer projectRows.Close()

	for projectRows.Next() {
		var (
			id   string
			name string
			code string
		)

		if err := projectRows.Scan(&id, &name, &code); err != nil {
			log.Fatal(err)
		} else {
			projectChannel.Rows = append(projectChannel.Rows, []string{id, name, code})
		}
	}
	response.Channels = append(response.Channels, projectChannel)
	c.JSON(http.StatusOK, response)

}
