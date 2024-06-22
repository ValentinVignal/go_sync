package main

import (
	"database/sql"
	"encoding/json"
	"strings"

	"log"

	"math/rand"

	"github.com/aidarkhanov/nanoid"
	"github.com/bxcodec/faker/v3"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func Seed(c *gin.Context) {
	log.Println("Seed")
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

	createTables(database)

	seedDatabase(database)
}

func createTables(database *sql.DB) {
	var databaseExists bool
	database.QueryRow("SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename  = 'project')").Scan(&databaseExists)
	if databaseExists {
		log.Println("Tables already created")
		return
	}
	log.Println("Creating tables...")
	createQueries := []string{
		`CREATE TABLE "project" ("updatedAt" bigint NOT NULL DEFAULT '0', "deletedAt" TIMESTAMP, "id" character varying NOT NULL, "name" character varying NOT NULL, "code" character varying NOT NULL, CONSTRAINT "PK_4d68b1358bb5b766d3e78f32f57" PRIMARY KEY ("id"))`,
		`CREATE TABLE "task" ("updatedAt" bigint NOT NULL DEFAULT '0', "deletedAt" TIMESTAMP, "id" character varying NOT NULL, "projectID" character varying NOT NULL, "name" character varying NOT NULL, "description" character varying, CONSTRAINT "PK_fb213f79ee45060ba925ecd576e" PRIMARY KEY ("id"))`,
		`CREATE TABLE "forms_tasks" ("updatedAt" bigint NOT NULL DEFAULT '0', "deletedAt" TIMESTAMP, "formID" character varying NOT NULL, "taskID" character varying NOT NULL, CONSTRAINT "PK_5cde27784334db1c9530bea6b5f" PRIMARY KEY ("formID", "taskID"))`,
		`CREATE TABLE "form" ("updatedAt" bigint NOT NULL DEFAULT '0', "deletedAt" TIMESTAMP, "id" character varying NOT NULL, "projectID" character varying NOT NULL, "name" character varying NOT NULL, "data" jsonb NOT NULL DEFAULT '{}', CONSTRAINT "PK_8f72b95aa2f8ba82cf95dc7579e" PRIMARY KEY ("id"))`,
		`CREATE TABLE "form_projects_project" ("formId" character varying NOT NULL, "projectId" character varying NOT NULL, CONSTRAINT "PK_0db033acf146ce2e7f99433877a" PRIMARY KEY ("formId", "projectId"))`,
		`CREATE INDEX "IDX_bc419c142f5336f4f3c4849788" ON "form_projects_project" ("formId")`,
		`CREATE INDEX "IDX_99033b0d627d82d697e1b3b08b" ON "form_projects_project" ("projectId")`,
		// cspell: disable-next-line
		`ALTER TABLE "task" ADD CONSTRAINT "FK_464e1e9f04be8ced7e4e878fbcf" FOREIGN KEY ("projectID") REFERENCES "project"("id") ON DELETE NO ACTION ON UPDATE NO ACTION`,
		`ALTER TABLE "forms_tasks" ADD CONSTRAINT "FK_f3ed34ef693480eda462df17b7b" FOREIGN KEY ("formID") REFERENCES "form"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
		`ALTER TABLE "forms_tasks" ADD CONSTRAINT "FK_0bc7355812c3784dd05b38e13f6" FOREIGN KEY ("taskID") REFERENCES "task"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
		`ALTER TABLE "form" ADD CONSTRAINT "FK_793836ec378a587c98a8c72a6b8" FOREIGN KEY ("projectID") REFERENCES "project"("id") ON DELETE NO ACTION ON UPDATE NO ACTION`,
		`ALTER TABLE "form_projects_project" ADD CONSTRAINT "FK_bc419c142f5336f4f3c4849788f" FOREIGN KEY ("formId") REFERENCES "form"("id") ON DELETE CASCADE ON UPDATE CASCADE`,
		`ALTER TABLE "form_projects_project" ADD CONSTRAINT "FK_99033b0d627d82d697e1b3b08bf" FOREIGN KEY ("projectId") REFERENCES "project"("id") ON DELETE CASCADE ON UPDATE CASCADE`,
	}

	for _, query := range createQueries {
		_, err := database.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Done creating tables")
}

type count struct {
	projects   int
	tasks      int
	forms      int
	formToTask int
}

var counts = count{
	projects:   500,
	tasks:      50_000,
	forms:      50_000,
	formToTask: 10_000,
}

func seedDatabase(database *sql.DB) {
	log.Println("Seeding database...")
	projectStatement, err := database.Prepare("INSERT INTO project(id, name, code) VALUES ($1, $2, $3)")
	if err != nil {
		log.Fatal(err)
	}
	defer projectStatement.Close()

	for i := 0; i < counts.projects; i++ {
		projectId := nanoid.New()
		projectName := faker.Word()
		projectCode := faker.Word()
		if _, err := projectStatement.Exec(projectId, projectName, projectCode); err != nil {
			log.Fatal(err)
		}
	}

	projectIds := []string{}
	rows, err := database.Query("SELECT id FROM project")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Fatal(err)
		} else {
			projectIds = append(projectIds, id)
		}
	}

	for i := 0; i < counts.tasks; i++ {
		taskId := nanoid.New()
		taskProjectId := projectIds[rand.Intn(len(projectIds))]
		taskName := faker.Sentence()
		taskDescription := faker.Paragraph()
		updatedAt := faker.UnixTime()
		if _, err := database.Exec(`INSERT INTO task(id, name, "projectID", description, "updatedAt") VALUES ($1, $2, $3, $4, $5)`, taskId, taskName, taskProjectId, taskDescription, updatedAt); err != nil {
			log.Fatal(err)
		}
	}

	fieldIds := strings.Split("0123456789abcdefghijklmnopqrstuvwxyz", "")

	for i := 0; i < counts.forms; i++ {
		formId := nanoid.New()
		formProjectId := projectIds[rand.Intn(len(projectIds))]
		formName := faker.Sentence()
		formData := make(map[string]string)
		for _, id := range fieldIds {
			formData[id] = faker.Sentence()
		}
		updatedAt := faker.UnixTime()
		jsonData, err := json.Marshal(formData)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := database.Exec(`INSERT INTO form(id, name, "projectID", data, "updatedAt") VALUES ($1, $2, $3, $4, $5)`, formId, formName, formProjectId, jsonData, updatedAt); err != nil {
			log.Fatal(err)
		}
	}

	taskIds := []string{}
	rows, err = database.Query("SELECT id FROM task")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Fatal(err)
		} else {
			taskIds = append(taskIds, id)
		}
	}

	formIds := []string{}
	rows, err = database.Query("SELECT id FROM form")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Fatal(err)
		} else {
			formIds = append(formIds, id)
		}
	}

	for i := 0; i < counts.formToTask; i++ {
		formId := formIds[rand.Intn(len(formIds))]
		taskId := taskIds[rand.Intn(len(taskIds))]
		updatedAt := faker.UnixTime()
		if _, err := database.Exec(`INSERT INTO forms_tasks("formID", "taskID", "updatedAt") VALUES ($1, $2, $3)`, formId, taskId, updatedAt); err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Done seeding")
}
