package main

import (
	"database/sql"

	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func drop(c *gin.Context) {
	log.Println("Drop")
	connString := "postgres://root:password@localhost:5555/go_sync?sslmode=disable"
	database, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Test the connection to the database
	if err := database.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully Connected")
	}

	dropTables(database)

}

func dropTables(database *sql.DB) {
	dropQueries := []string{
		`ALTER TABLE "form_projects_project" DROP CONSTRAINT "FK_99033b0d627d82d697e1b3b08bf"`,
		`ALTER TABLE "form_projects_project" DROP CONSTRAINT "FK_bc419c142f5336f4f3c4849788f"`,
		`ALTER TABLE "form" DROP CONSTRAINT "FK_793836ec378a587c98a8c72a6b8"`,
		`ALTER TABLE "forms_tasks" DROP CONSTRAINT "FK_0bc7355812c3784dd05b38e13f6"`,
		`ALTER TABLE "forms_tasks" DROP CONSTRAINT "FK_f3ed34ef693480eda462df17b7b"`,
		`ALTER TABLE "task" DROP CONSTRAINT "FK_464e1e9f04be8ced7e4e878fbcf"`,
		`DROP INDEX "public"."IDX_99033b0d627d82d697e1b3b08b"`,
		`DROP INDEX "public"."IDX_bc419c142f5336f4f3c4849788"`,
		`DROP TABLE "form_projects_project"`,
		`DROP TABLE "form"`,
		`DROP TABLE "forms_tasks"`,
		`DROP TABLE "task"`,
		`DROP TABLE "project"`,
	}

	for _, query := range dropQueries {
		_, err := database.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}
}
