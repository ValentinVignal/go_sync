package main

import (
	"net/http"

	"log"

	"github.com/gin-gonic/gin"
)

// https://pkg.go.dev/github.com/lib/pq
// https://medium.com/@dewirahmawatie/connecting-to-postgresql-in-golang-59d7b208bad2

func main() {
	router := gin.Default()
	router.GET("/ping", ping)
	router.POST("/seed", seed)
	router.POST("/drop", drop)
	router.GET("/sync", sync)
	router.Run("localhost:8080")

}

func ping(c *gin.Context) {
	log.Println("ping")
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func seed(c *gin.Context) {

}
func drop(c *gin.Context) {

}
func sync(c *gin.Context) {

}
