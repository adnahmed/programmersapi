package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

type Invite struct {
	Login string `json:"login"`
}

func addCommonResponseHeaders(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		panic("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/users", func(c *gin.Context) {
		addCommonResponseHeaders(c)
		c.JSON(http.StatusOK, GetUserLists())
	})

	router.POST("/users/invite", func(c *gin.Context) {
		addCommonResponseHeaders(c)
		var invite Invite
		c.BindJSON(&invite)
		InviteUser(invite.Login)
		c.JSON(http.StatusAccepted, nil)
	})

	router.Run(":" + port)
}
