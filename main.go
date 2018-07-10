package main

import (
		_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/gin-gonic/gin"
)

func main() {
	initLogger()
	log.Debug().Msg("Starting app")

	log.Debug().Msg("Starting databases initialisation")
	// go initDatabases()
	go initDatabasesFromJson()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	initHandlers(r)

	log.Debug().Msg("Starting web service with port: 8079")
	r.Run(":8079")
}
