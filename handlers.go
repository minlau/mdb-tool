package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func initHandlers(r *gin.Engine, store *DatabaseStore) {
	r.StaticFile("/", "./assets/index.html")
	r.Static("/assets", "./assets")
	r.GET("/databases", getDatabases(store))
	r.GET("/query", query(store))
}

type queryRequest struct {
	GroupId   *int   `form:"groupId" json:"groupId"`
	GroupType string `form:"groupType" json:"groupType" binding:"required"`
	Query     string `form:"query" json:"query" binding:"required"`
}

func query(store *DatabaseStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req queryRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
			return
		}

		if req.GroupType == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type is required"})
			return
		}

		if req.GroupId == nil {
			c.JSON(200, store.QueryMultipleDatabases(req.GroupType, req.Query))
		} else {
			c.JSON(200, store.QueryDatabase(*req.GroupId, req.GroupType, req.Query))
		}
	}
}

func getDatabases(store *DatabaseStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, store.GetDatabaseItems())
	}
}
