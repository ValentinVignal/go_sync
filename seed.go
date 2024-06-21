package main

import (
	"database/sql"

	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func seed(c *gin.Context) {
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

}

func createTables(database *sql.DB) {

	var databaseExists bool
	database.QueryRow("SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename  = 'project')").Scan(&databaseExists)
	println(databaseExists)
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
