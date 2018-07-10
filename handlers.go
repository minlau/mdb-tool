package main

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"net/http"
)

func initHandlers(r *gin.Engine) {
	r.POST("/query/all", queryAll)
	r.POST("/query", queryById)
}

func queryAll(c *gin.Context) {
	c.JSON(200, queryAllDatabases(c.Query("query")))
}

func queryById(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	data, err := queryDatabase(id, c.PostForm("query"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	c.JSON(200, data)
}